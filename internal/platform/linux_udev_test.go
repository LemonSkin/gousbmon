//go:build linux && udev && !sd_device

package platform

import "testing"

func TestNew(t *testing.T) {
	d, err := New(nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	if d == nil {
		t.Fatal("New() returned nil detector")
	}
}

// TestGetAvailableDevices_Smoke ensures the libudev detector enumerates without crashing.
func TestGetAvailableDevices_Smoke(t *testing.T) {
	devices, err := (&udevDetector{}).GetAvailableDevices()
	if err != nil {
		t.Fatalf("GetAvailableDevices failed: %v", err)
	}
	if devices == nil {
		t.Fatal("expected non-nil map")
	}
	for key, info := range devices {
		if info.DevType != "usb_device" {
			t.Errorf("device %q has DevType %q, want usb_device", key, info.DevType)
		}
	}
}
