//go:build linux

package platform

import "testing"

func TestIsRootHub(t *testing.T) {
	tests := []struct {
		name    string
		sysname string
		want    bool
	}{
		{"usb1 root hub", "usb1", true},
		{"usb20 root hub", "usb20", true},
		{"external hub", "1-1", false},
		{"nested device", "2-1.4", false},
		{"interface", "1-1:1.0", false},
		{"empty", "", false},
		{"usb prefix only", "usb", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRootHub(tt.sysname); got != tt.want {
				t.Errorf("isRootHub(%q) = %v, want %v", tt.sysname, got, tt.want)
			}
		})
	}
}

func TestBuildDevName(t *testing.T) {
	tests := []struct {
		name           string
		busnum, devnum int
		want           string
	}{
		{"standard", 1, 4, "/dev/bus/usb/001/004"},
		{"padding", 12, 137, "/dev/bus/usb/012/137"},
		{"zero bus", 0, 4, ""},
		{"zero dev", 1, 0, ""},
		{"negative", -1, -1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildDevName(tt.busnum, tt.devnum); got != tt.want {
				t.Errorf("buildDevName(%d, %d) = %q, want %q", tt.busnum, tt.devnum, got, tt.want)
			}
		})
	}
}

func TestInfoFromProperties(t *testing.T) {
	props := map[string]string{
		"ID_VENDOR_ID":               "046D",
		"ID_MODEL_ID":                "C077",
		"ID_VENDOR":                  "Logitech",
		"ID_MODEL":                   "USB Optical Mouse",
		"ID_SERIAL":                  "Logitech_USB_Optical_Mouse",
		"ID_SERIAL_SHORT":            "ABC123",
		"ID_REVISION":                "7200",
		"ID_USB_INTERFACES":          ":030102:",
		"ID_USB_CLASS_FROM_DATABASE": "Human Interface Device",
		"ID_VENDOR_FROM_DATABASE":    "Logitech, Inc.",
		"ID_MODEL_FROM_DATABASE":     "M105 Optical Mouse",
		"DEVNAME":                    "/dev/bus/usb/001/004",
		"DEVTYPE":                    "usb_device",
	}

	got := infoFromProperties(props)

	checks := map[string]struct{ got, want string }{
		"IDVendorID":             {got.IDVendorID, "046d"},
		"IDModelID":              {got.IDModelID, "c077"},
		"IDVendor":               {got.IDVendor, "Logitech"},
		"IDModel":                {got.IDModel, "USB Optical Mouse"},
		"IDSerial":               {got.IDSerial, "ABC123"},
		"IDRevision":             {got.IDRevision, "7200"},
		"IDUSBInterfaces":        {got.IDUSBInterfaces, ":030102:"},
		"IDUSBClassFromDatabase": {got.IDUSBClassFromDatabase, "Human Interface Device"},
		"IDVendorFromDatabase":   {got.IDVendorFromDatabase, "Logitech, Inc."},
		"IDModelFromDatabase":    {got.IDModelFromDatabase, "M105 Optical Mouse"},
		"DevName":                {got.DevName, "/dev/bus/usb/001/004"},
		"DevType":                {got.DevType, "usb_device"},
	}
	for field, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", field, c.got, c.want)
		}
	}
}

func TestInfoFromProperties_SerialFallback(t *testing.T) {
	// When ID_SERIAL_SHORT is absent, ID_SERIAL is used.
	got := infoFromProperties(map[string]string{"ID_SERIAL": "full-serial"})
	if got.IDSerial != "full-serial" {
		t.Errorf("IDSerial = %q, want %q", got.IDSerial, "full-serial")
	}
}
