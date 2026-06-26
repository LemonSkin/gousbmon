//go:build windows

package platform

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/LemonSkin/gousbmon/device"
)

// SetupAPI bindings. We declare the procedures ourselves but rely on golang.org/x/sys/windows for the typed helpers
// (GUID, UTF-16 conversion, well-known error values, lazy system DLL loading).
var (
	modSetupAPI = windows.NewLazySystemDLL("setupapi.dll")

	procSetupDiGetClassDevsW              = modSetupAPI.NewProc("SetupDiGetClassDevsW")
	procSetupDiEnumDeviceInfo             = modSetupAPI.NewProc("SetupDiEnumDeviceInfo")
	procSetupDiGetDeviceInstanceIdW       = modSetupAPI.NewProc("SetupDiGetDeviceInstanceIdW")
	procSetupDiGetDeviceRegistryPropertyW = modSetupAPI.NewProc("SetupDiGetDeviceRegistryPropertyW")
	procSetupDiDestroyDeviceInfoList      = modSetupAPI.NewProc("SetupDiDestroyDeviceInfoList")
)

const (
	digcfPresent    = 0x00000002
	digcfAllClasses = 0x00000004

	// SPDRP_* device registry property indices.
	spdrpDeviceDesc    = 0x00000000
	spdrpHardwareID    = 0x00000001
	spdrpCompatibleIDs = 0x00000002
	spdrpClass         = 0x00000007
	spdrpMfg           = 0x0000000B
	spdrpFriendlyName  = 0x0000000C
)

const invalidHandle = ^uintptr(0)

// spDevInfoData mirrors the Win32 SP_DEVINFO_DATA structure.
type spDevInfoData struct {
	cbSize    uint32
	classGUID windows.GUID
	devInst   uint32
	reserved  uintptr
}

// nonUSBDeviceIDs are PNPDeviceID substrings for entries that are reported under the USB tree but are not real USB
// peripherals.
var nonUSBDeviceIDs = []string{"ROOT_HUB20", "ROOT_HUB30", "VIRTUAL_POWER_PDO"}

var (
	vidRe  = regexp.MustCompile(`(?i)VID_([0-9A-Fa-f]{4})`)
	pidRe  = regexp.MustCompile(`(?i)PID_([0-9A-Fa-f]{4})`)
	venRe  = regexp.MustCompile(`VEN_([a-zA-Z0-9._/\-]{2,8})&`)
	prodRe = regexp.MustCompile(`PROD_([a-zA-Z0-9_/.\-]{2,16})&`)
)

// windowsDetector implements device.Detector for Windows.
type windowsDetector struct{}

// New returns the Windows USB detector.
func New() (device.Detector, error) {
	return &windowsDetector{}, nil
}

// winDevice holds the raw device properties read from SetupAPI before they are normalised into a device.Info.
type winDevice struct {
	DeviceID     string
	PNPDeviceID  string
	Name         string
	Caption      string
	Manufacturer string
	PNPClass     string
	HardwareID   []string
	CompatibleID []string
}

func (d *windowsDetector) GetAvailableDevices() (map[string]device.Info, error) {
	// Restrict enumeration to the "USB" enumerator and to devices that are currently present.
	enumerator, err := windows.UTF16PtrFromString("USB")
	if err != nil {
		return nil, err
	}

	hDevInfo, _, callErr := procSetupDiGetClassDevsW.Call(
		0,
		uintptr(unsafe.Pointer(enumerator)),
		0,
		uintptr(digcfPresent|digcfAllClasses),
	)
	if hDevInfo == invalidHandle {
		return nil, fmt.Errorf("gousbmon: SetupDiGetClassDevs failed: %w", callErr)
	}
	defer procSetupDiDestroyDeviceInfoList.Call(hDevInfo)

	result := make(map[string]device.Info)
	for index := uint32(0); ; index++ {
		var did spDevInfoData
		did.cbSize = uint32(unsafe.Sizeof(did))

		ok, _, e := procSetupDiEnumDeviceInfo.Call(hDevInfo, uintptr(index), uintptr(unsafe.Pointer(&did)))
		if ok == 0 {
			if errors.Is(e, windows.ERROR_NO_MORE_ITEMS) {
				break
			}
			return nil, fmt.Errorf("gousbmon: SetupDiEnumDeviceInfo failed: %w", e)
		}

		instanceID, err := getDeviceInstanceID(hDevInfo, &did)
		if err != nil || instanceID == "" {
			continue
		}
		if isNonUSBDevice(instanceID) {
			continue
		}

		friendly := getStringProperty(hDevInfo, &did, spdrpFriendlyName)
		desc := getStringProperty(hDevInfo, &did, spdrpDeviceDesc)
		name := friendly
		if name == "" {
			name = desc
		}

		dev := winDevice{
			DeviceID:     instanceID,
			PNPDeviceID:  instanceID,
			Name:         name,
			Caption:      desc,
			Manufacturer: getStringProperty(hDevInfo, &did, spdrpMfg),
			PNPClass:     getStringProperty(hDevInfo, &did, spdrpClass),
			HardwareID:   getMultiStringProperty(hDevInfo, &did, spdrpHardwareID),
			CompatibleID: getMultiStringProperty(hDevInfo, &did, spdrpCompatibleIDs),
		}
		result[instanceID] = toDeviceInfo(dev)
	}
	return result, nil
}

// getDeviceInstanceID returns the device instance ID (e.g. "USB\\VID_046D&PID_C077\\...") for the given device.
func getDeviceInstanceID(hDevInfo uintptr, did *spDevInfoData) (string, error) {
	var required uint32
	procSetupDiGetDeviceInstanceIdW.Call(
		hDevInfo,
		uintptr(unsafe.Pointer(did)),
		0,
		0,
		uintptr(unsafe.Pointer(&required)),
	)
	if required == 0 {
		return "", nil
	}

	buf := make([]uint16, required)
	ok, _, e := procSetupDiGetDeviceInstanceIdW.Call(
		hDevInfo,
		uintptr(unsafe.Pointer(did)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(required),
		uintptr(unsafe.Pointer(&required)),
	)
	if ok == 0 {
		return "", e
	}
	return windows.UTF16ToString(buf), nil
}

// getDeviceProperty fetches a raw SPDRP_* device registry property as a UTF-16 buffer. It returns false if the property
// is absent.
func getDeviceProperty(hDevInfo uintptr, did *spDevInfoData, prop uint32) ([]uint16, bool) {
	var required uint32
	procSetupDiGetDeviceRegistryPropertyW.Call(
		hDevInfo,
		uintptr(unsafe.Pointer(did)),
		uintptr(prop),
		0,
		0,
		0,
		uintptr(unsafe.Pointer(&required)),
	)
	if required == 0 {
		return nil, false
	}

	// required is a byte count; the buffer holds UTF-16 code units.
	buf := make([]uint16, (required+1)/2)
	ok, _, _ := procSetupDiGetDeviceRegistryPropertyW.Call(
		hDevInfo,
		uintptr(unsafe.Pointer(did)),
		uintptr(prop),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(required),
		uintptr(unsafe.Pointer(&required)),
	)
	if ok == 0 {
		return nil, false
	}
	return buf, true
}

func getStringProperty(hDevInfo uintptr, did *spDevInfoData, prop uint32) string {
	buf, ok := getDeviceProperty(hDevInfo, did, prop)
	if !ok {
		return ""
	}
	return windows.UTF16ToString(buf)
}

func getMultiStringProperty(hDevInfo uintptr, did *spDevInfoData, prop uint32) []string {
	buf, ok := getDeviceProperty(hDevInfo, did, prop)
	if !ok {
		return nil
	}
	return parseMultiSz(buf)
}

// parseMultiSz splits a REG_MULTI_SZ buffer (NUL-separated, double-NUL terminated) into individual strings.
func parseMultiSz(buf []uint16) []string {
	var out []string
	start := 0
	for i, v := range buf {
		if v != 0 {
			continue
		}
		if i == start {
			break
		}
		out = append(out, windows.UTF16ToString(buf[start:i]))
		start = i + 1
	}
	return out
}

func isNonUSBDevice(pnpDeviceID string) bool {
	for _, substr := range nonUSBDeviceIDs {
		if strings.Contains(pnpDeviceID, substr) {
			return true
		}
	}
	return false
}

func toDeviceInfo(dev winDevice) device.Info {
	info := device.Info{
		IDModel:                dev.Name,
		IDModelFromDatabase:    dev.Caption,
		IDVendor:               dev.Name,
		IDVendorFromDatabase:   dev.Manufacturer,
		IDUSBInterfaces:        strings.Join(dev.CompatibleID, ", "),
		IDUSBClassFromDatabase: dev.PNPClass,
		DevName:                dev.DeviceID,
	}

	segments := strings.Split(dev.DeviceID, `\`)
	driver := ""
	if len(segments) > 0 {
		driver = strings.ToUpper(segments[0])
	}
	info.DevType = driver
	if len(segments) > 1 {
		info.IDSerial = segments[len(segments)-1]
	}

	switch driver {
	case "USBSTOR":
		info.IDModelID = firstMatch(prodRe, dev.DeviceID)
		info.IDVendorID = firstMatch(venRe, dev.DeviceID)
	default:
		info.IDModelID = firstMatch(pidRe, dev.DeviceID)
		info.IDVendorID = firstMatch(vidRe, dev.DeviceID)
	}

	info.IDModelID = strings.ToUpper(info.IDModelID)
	info.IDVendorID = strings.ToUpper(info.IDVendorID)
	return info
}

func firstMatch(re *regexp.Regexp, value string) string {
	m := re.FindStringSubmatch(value)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}
