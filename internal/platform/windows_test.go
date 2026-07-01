//go:build windows

package platform

import (
	"reflect"
	"regexp"
	"testing"
)

// utf16Multi builds a REG_MULTI_SZ-style buffer (NUL-separated, double-NUL terminated) from the given strings.
func utf16Multi(values ...string) []uint16 {
	var buf []uint16
	for _, v := range values {
		for _, r := range v {
			buf = append(buf, uint16(r))
		}
		buf = append(buf, 0)
	}
	buf = append(buf, 0)
	return buf
}

// TestParseMultiSz tests splitting of REG_MULTI_SZ buffers.
func TestParseMultiSz(t *testing.T) {
	tests := []struct {
		name string
		in   []uint16
		want []string
	}{
		{"empty", []uint16{0}, nil},
		{"single", utf16Multi("abc"), []string{"abc"}},
		{"multiple", utf16Multi("abc", "def", "ghi"), []string{"abc", "def", "ghi"}},
		{"nil buffer", nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseMultiSz(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseMultiSz(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestIsNonUSBDevice tests the non-USB device filter.
func TestIsNonUSBDevice(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"root hub 2.0", `USB\ROOT_HUB20\4&ABCD`, true},
		{"root hub 3.0", `USB\ROOT_HUB30\4&ABCD`, true},
		{"virtual power", `USB\VIRTUAL_POWER_PDO\3&XYZ`, true},
		{"real device", `USB\VID_046D&PID_C077\6&1234ABCD&0&1`, false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNonUSBDevice(tt.id); got != tt.want {
				t.Errorf("isNonUSBDevice(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

// TestFirstMatch tests regex submatch extraction.
func TestFirstMatch(t *testing.T) {
	tests := []struct {
		name string
		re   *regexp.Regexp
		in   string
		want string
	}{
		{"vid match", vidRe, `USB\VID_046D&PID_C077`, "046D"},
		{"pid match", pidRe, `USB\VID_046D&PID_C077`, "C077"},
		{"ven match", venRe, `USBSTOR\VEN_SANDISK&PROD_CRUZER`, "SANDISK"},
		{"prod match", prodRe, `USBSTOR\VEN_SANDISK&PROD_CRUZER&`, "CRUZER"},
		{"no match", vidRe, `nothing here`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := firstMatch(tt.re, tt.in); got != tt.want {
				t.Errorf("firstMatch(%v, %q) = %q, want %q", tt.re, tt.in, got, tt.want)
			}
		})
	}
}

// TestToDeviceInfo tests normalization of raw Windows devices into device.DeviceInfo.
func TestToDeviceInfo(t *testing.T) {
	t.Run("standard USB device", func(t *testing.T) {
		dev := winDevice{
			DeviceID:     `USB\VID_046D&PID_C077\6&1234ABCD&0&1`,
			Name:         "USB Optical Mouse",
			Caption:      "HID-compliant mouse",
			Manufacturer: "Logitech",
			PNPClass:     "Mouse",
			CompatibleID: []string{"USB\\Class_03", "USB\\Class_03&SubClass_01"},
		}
		got := toDeviceInfo(dev)

		if got.IDModel != "USB Optical Mouse" {
			t.Errorf("IDModel = %q, want %q", got.IDModel, "USB Optical Mouse")
		}
		if got.IDVendorID != "046D" {
			t.Errorf("IDVendorID = %q, want %q", got.IDVendorID, "046D")
		}
		if got.IDModelID != "C077" {
			t.Errorf("IDModelID = %q, want %q", got.IDModelID, "C077")
		}
		if got.DevType != "USB" {
			t.Errorf("DevType = %q, want %q", got.DevType, "USB")
		}
		if got.IDSerial != "6&1234ABCD&0&1" {
			t.Errorf("IDSerial = %q, want %q", got.IDSerial, "6&1234ABCD&0&1")
		}
		if got.IDUSBInterfaces != "USB\\Class_03, USB\\Class_03&SubClass_01" {
			t.Errorf("IDUSBInterfaces = %q", got.IDUSBInterfaces)
		}
	})

	t.Run("USBSTOR device", func(t *testing.T) {
		dev := winDevice{
			DeviceID: `USBSTOR\VEN_SANDISK&PROD_CRUZER&REV_1.00\1234567890`,
			Name:     "SanDisk Cruzer",
		}
		got := toDeviceInfo(dev)

		if got.DevType != "USBSTOR" {
			t.Errorf("DevType = %q, want %q", got.DevType, "USBSTOR")
		}
		if got.IDVendorID != "SANDISK" {
			t.Errorf("IDVendorID = %q, want %q", got.IDVendorID, "SANDISK")
		}
		if got.IDModelID != "CRUZER" {
			t.Errorf("IDModelID = %q, want %q", got.IDModelID, "CRUZER")
		}
	})

	t.Run("device with no segments", func(t *testing.T) {
		dev := winDevice{DeviceID: "BADID"}
		got := toDeviceInfo(dev)

		if got.DevType != "BADID" {
			t.Errorf("DevType = %q, want %q", got.DevType, "BADID")
		}
		if got.IDSerial != "" {
			t.Errorf("IDSerial = %q, want empty", got.IDSerial)
		}
	})
}

// TestNew tests the platform-specific constructor.
func TestNew(t *testing.T) {
	d, err := New(nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if d == nil {
		t.Fatal("New() returned nil detector")
	}
}

// TestGetAvailableDevices_Smoke is an integration test that ensures the Windows detector can enumerate devices without
// crashing.
func TestGetAvailableDevices_Smoke(t *testing.T) {
	d := &windowsDetector{}
	devs, err := d.GetAvailableDevices()
	if err != nil {
		t.Fatalf("GetAvailableDevices failed: %v", err)
	}

	// Test assumes at least one device is connected. Mostly testing that the thing doesn't crash or freak out.
	if devs == nil {
		t.Fatal("expected non-nil map")
	}
}
