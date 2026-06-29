//go:build linux && !udev && !sd_device

package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LemonSkin/gousbmon/device"
)

// defaultSysfsRoot is the standard sysfs location listing USB devices and interfaces.
const defaultSysfsRoot = "/sys/bus/usb/devices"

// sysfsDetector implements device.Detector by reading /sys/bus/usb/devices directly.
type sysfsDetector struct {
	root string
	ids  *usbIDs
}

// New returns the Linux USB detector.
func New() (device.Detector, error) {
	fmt.Printf("Using sysfs detector\n")
	return &sysfsDetector{root: defaultSysfsRoot, ids: loadUSBIDs()}, nil
}

func (d *sysfsDetector) GetAvailableDevices() (map[string]device.Info, error) {
	entries, err := os.ReadDir(d.root)
	if err != nil {
		return nil, fmt.Errorf("gousbmon: read %s: %w", d.root, err)
	}

	result := make(map[string]device.Info)
	for _, entry := range entries {
		name := entry.Name()
		// Interfaces (e.g. "1-1:1.0") contain a colon, only want device nodes
		if strings.Contains(name, ":") {
			continue
		}
		if isRootHub(name) {
			continue
		}

		raw, ok := readSysfsDevice(filepath.Join(d.root, name), name)
		if !ok {
			// Missing core descriptors, skip it
			continue
		}

		info := sysfsToDeviceInfo(raw, d.ids)
		key := info.DevName
		if key == "" {
			key = name
		}
		result[key] = info
	}
	return result, nil
}

// sysfsInterface holds the class triple read from a single USB interface directory.
type sysfsInterface struct {
	class    string
	subclass string
	protocol string
}

// sysfsDevice holds the raw attributes read from a USB device directory in sysfs.
type sysfsDevice struct {
	sysname      string
	idVendor     string
	idProduct    string
	manufacturer string
	product      string
	serial       string
	bcdDevice    string
	deviceClass  string
	busnum       int
	devnum       int
	interfaces   []sysfsInterface
}

// readSysfsDevice reads a USB device directory. It returns false if the directory lacks the core descriptor files
// (idVendor/idProduct), which indicates the device is not fully enumerated.
func readSysfsDevice(dir, sysname string) (sysfsDevice, bool) {
	idVendor := readSysfsAttr(dir, "idVendor")
	idProduct := readSysfsAttr(dir, "idProduct")
	if idVendor == "" || idProduct == "" {
		return sysfsDevice{}, false
	}

	dev := sysfsDevice{
		sysname:      sysname,
		idVendor:     strings.ToLower(idVendor),
		idProduct:    strings.ToLower(idProduct),
		manufacturer: readSysfsAttr(dir, "manufacturer"),
		product:      readSysfsAttr(dir, "product"),
		serial:       readSysfsAttr(dir, "serial"),
		bcdDevice:    readSysfsAttr(dir, "bcdDevice"),
		deviceClass:  strings.ToLower(readSysfsAttr(dir, "bDeviceClass")),
		busnum:       atoiSafe(readSysfsAttr(dir, "busnum")),
		devnum:       atoiSafe(readSysfsAttr(dir, "devnum")),
	}
	dev.interfaces = readSysfsInterfaces(dir, sysname)
	return dev, true
}

// readSysfsInterfaces reads the interface sub-directories (e.g. "<sysname>:c.i") within a device directory.
func readSysfsInterfaces(dir, sysname string) []sysfsInterface {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	prefix := sysname + ":"
	var ifaces []sysfsInterface
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		ifaceDir := filepath.Join(dir, entry.Name())
		ifaces = append(ifaces, sysfsInterface{
			class:    strings.ToLower(readSysfsAttr(ifaceDir, "bInterfaceClass")),
			subclass: strings.ToLower(readSysfsAttr(ifaceDir, "bInterfaceSubClass")),
			protocol: strings.ToLower(readSysfsAttr(ifaceDir, "bInterfaceProtocol")),
		})
	}
	return ifaces
}

// sysfsToDeviceInfo normalises a raw sysfs device into a device.Info.
func sysfsToDeviceInfo(dev sysfsDevice, ids *usbIDs) device.Info {
	info := device.Info{
		IDVendorID:           dev.idVendor,
		IDModelID:            dev.idProduct,
		IDVendor:             dev.manufacturer,
		IDModel:              dev.product,
		IDSerial:             dev.serial,
		IDRevision:           dev.bcdDevice,
		IDUSBInterfaces:      buildUSBInterfaces(dev.interfaces),
		DevType:              "usb_device",
		DevName:              buildDevName(dev.busnum, dev.devnum),
		IDVendorFromDatabase: ids.vendor(dev.idVendor),
		IDModelFromDatabase:  ids.product(dev.idVendor, dev.idProduct),
	}

	// The device class is "00" when the class is declared per-interface. Fall back to the first interface's class.
	classCode := dev.deviceClass
	if (classCode == "" || classCode == "00") && len(dev.interfaces) > 0 {
		classCode = dev.interfaces[0].class
	}
	info.IDUSBClassFromDatabase = ids.class(classCode)
	return info
}

// buildUSBInterfaces renders the udev-style ID_USB_INTERFACES string, e.g. ":030102:080650:". Each interface is encoded
// as six hex digits (class, subclass, protocol) and the whole string is colon-delimited with leading/trailing colons.
func buildUSBInterfaces(ifaces []sysfsInterface) string {
	if len(ifaces) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteByte(':')
	for _, iface := range ifaces {
		fmt.Fprintf(&b, "%s%s%s:", pad2(iface.class), pad2(iface.subclass), pad2(iface.protocol))
	}
	return b.String()
}

// pad2 normalises a hex byte string into two lowercase digits.
func pad2(s string) string {
	if s == "" {
		return "00"
	}
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

// readSysfsAttr reads a single sysfs attribute file, returning its contents or an empty string on error.
func readSysfsAttr(dir, name string) string {
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// atoiSafe parses an integer, returning 0 when the input is empty or invalid.
func atoiSafe(s string) int {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
