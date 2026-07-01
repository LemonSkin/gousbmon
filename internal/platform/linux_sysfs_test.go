//go:build linux && !udev && !sd_device

package platform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LemonSkin/gousbmon/device"
)

// writeSysfsFixture builds a fake /sys/bus/usb/devices tree in a temp directory and returns its root. The fixture is
// generated at runtime rather than committed because interface directories are named with a colon (e.g. "1-1:1.0"),
// and these break Windows checkouts.
func writeSysfsFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write := func(rel, content string) {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		// sysfs attribute files end in a trailing newline; readSysfsAttr trims it.
		if err := os.WriteFile(path, []byte(content+"\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	// Root hub (filtered out by name).
	write("usb1/idVendor", "1d6b")
	write("usb1/idProduct", "0002")
	write("usb1/busnum", "1")
	write("usb1/devnum", "1")
	write("usb1/bDeviceClass", "09")

	// Fully-enumerated mouse with a single HID interface.
	write("1-1/idVendor", "046d")
	write("1-1/idProduct", "c077")
	write("1-1/manufacturer", "Logitech")
	write("1-1/product", "USB Optical Mouse")
	write("1-1/serial", "ABC123")
	write("1-1/bcdDevice", "7200")
	write("1-1/busnum", "1")
	write("1-1/devnum", "4")
	write("1-1/bDeviceClass", "00")
	write("1-1/1-1:1.0/bInterfaceClass", "03")
	write("1-1/1-1:1.0/bInterfaceSubClass", "01")
	write("1-1/1-1:1.0/bInterfaceProtocol", "02")

	// Top-level interface directory (skipped because its name contains a colon).
	write("1-1:1.0/bInterfaceClass", "03")

	// Half-enumerated device missing idVendor/idProduct (skipped).
	write("2-1/manufacturer", "Half-enumerated Device")

	return root
}

func TestBuildUSBInterfaces(t *testing.T) {
	tests := []struct {
		name   string
		ifaces []sysfsInterface
		want   string
	}{
		{"none", nil, ""},
		{"single", []sysfsInterface{{"03", "01", "02"}}, ":030102:"},
		{
			"multiple",
			[]sysfsInterface{{"08", "06", "50"}, {"03", "00", "00"}},
			":080650:030000:",
		},
		{"padding single digit", []sysfsInterface{{"3", "1", "2"}}, ":030102:"},
		{"missing fields default zero", []sysfsInterface{{"ff", "", ""}}, ":ff0000:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildUSBInterfaces(tt.ifaces); got != tt.want {
				t.Errorf("buildUSBInterfaces(%v) = %q, want %q", tt.ifaces, got, tt.want)
			}
		})
	}
}

func TestSysfsToDeviceInfo(t *testing.T) {
	ids := newFixtureIDs(t)
	dev := sysfsDevice{
		sysname:      "1-1",
		idVendor:     "046d",
		idProduct:    "c077",
		manufacturer: "Logitech",
		product:      "USB Optical Mouse",
		serial:       "ABC123",
		bcdDevice:    "7200",
		deviceClass:  "00",
		busnum:       1,
		devnum:       4,
		interfaces:   []sysfsInterface{{"03", "01", "02"}},
	}

	got := sysfsToDeviceInfo(dev, ids)

	checks := map[string]struct{ got, want string }{
		"IDVendorID":           {got.IDVendorID, "046d"},
		"IDModelID":            {got.IDModelID, "c077"},
		"IDVendor":             {got.IDVendor, "Logitech"},
		"IDModel":              {got.IDModel, "USB Optical Mouse"},
		"IDSerial":             {got.IDSerial, "ABC123"},
		"IDRevision":           {got.IDRevision, "7200"},
		"DevType":              {got.DevType, "usb_device"},
		"DevName":              {got.DevName, "/dev/bus/usb/001/004"},
		"IDUSBInterfaces":      {got.IDUSBInterfaces, ":030102:"},
		"IDVendorFromDatabase": {got.IDVendorFromDatabase, "Logitech, Inc."},
		"IDModelFromDatabase":  {got.IDModelFromDatabase, "M105 Optical Mouse"},
		// bDeviceClass 00 falls back to the first interface's class (03 -> HID).
		"IDUSBClassFromDatabase": {got.IDUSBClassFromDatabase, "Human Interface Device"},
	}
	for field, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", field, c.got, c.want)
		}
	}
}

func TestSysfsToDeviceInfo_ExplicitClass(t *testing.T) {
	ids := newFixtureIDs(t)
	dev := sysfsDevice{
		idVendor:    "1d6b",
		idProduct:   "0002",
		deviceClass: "09",
		busnum:      1,
		devnum:      1,
	}
	got := sysfsToDeviceInfo(dev, ids)
	if got.IDUSBClassFromDatabase != "Hub" {
		t.Errorf("IDUSBClassFromDatabase = %q, want %q", got.IDUSBClassFromDatabase, "Hub")
	}
}

func TestReadSysfsDevice(t *testing.T) {
	root := writeSysfsFixture(t)

	t.Run("usable device", func(t *testing.T) {
		dev, ok := readSysfsDevice(filepath.Join(root, "1-1"), "1-1")
		if !ok {
			t.Fatal("expected ok=true for fully enumerated device")
		}
		if dev.idVendor != "046d" || dev.idProduct != "c077" {
			t.Errorf("ids = %q:%q", dev.idVendor, dev.idProduct)
		}
		if dev.busnum != 1 || dev.devnum != 4 {
			t.Errorf("bus/dev = %d/%d", dev.busnum, dev.devnum)
		}
		if len(dev.interfaces) != 1 || dev.interfaces[0].class != "03" {
			t.Errorf("interfaces = %+v", dev.interfaces)
		}
	})

	t.Run("uninitialised device skipped", func(t *testing.T) {
		if _, ok := readSysfsDevice(filepath.Join(root, "2-1"), "2-1"); ok {
			t.Error("expected ok=false for device missing idVendor/idProduct")
		}
	})
}

func TestGetAvailableDevices(t *testing.T) {
	d := &sysfsDetector{root: writeSysfsFixture(t)}
	devices, err := d.GetAvailableDevices()
	if err != nil {
		t.Fatalf("GetAvailableDevices failed: %v", err)
	}

	// Only the mouse should remain: usb1 is a root hub, 2-1 is uninitialised, and "1-1:1.0" is an interface.
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d: %v", len(devices), devices)
	}

	info, ok := devices["/dev/bus/usb/001/004"]
	if !ok {
		t.Fatalf("expected device keyed by DevName, got keys: %v", keysOf(devices))
	}
	if info.IDVendorID != "046d" || info.IDModelID != "c077" {
		t.Errorf("ids = %q:%q", info.IDVendorID, info.IDModelID)
	}
	if info.IDModel != "USB Optical Mouse" {
		t.Errorf("IDModel = %q", info.IDModel)
	}
	if info.IDUSBInterfaces != ":030102:" {
		t.Errorf("IDUSBInterfaces = %q", info.IDUSBInterfaces)
	}
	if info.DevType != "usb_device" {
		t.Errorf("DevType = %q", info.DevType)
	}
}

func TestGetAvailableDevices_BadRoot(t *testing.T) {
	d := &sysfsDetector{root: filepath.Join(t.TempDir(), "does-not-exist")}
	if _, err := d.GetAvailableDevices(); err == nil {
		t.Error("expected error for non-existent sysfs root")
	}
}

func TestNew(t *testing.T) {
	d, err := New(nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if d == nil {
		t.Fatal("New() returned nil detector")
	}
}

func keysOf(m map[string]device.DeviceInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
