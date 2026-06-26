//go:build windows

package platform

import (
	"testing"
)

// TestNew tests the platform-specific constructor.
func TestNew(t *testing.T) {
	d, err := New()
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
