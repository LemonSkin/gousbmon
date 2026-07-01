package internal

import (
	"testing"

	"github.com/LemonSkin/gousbmon/device"
)

// TestDiffNewDevice tests that a new device is detected
func TestDiffNewDevice(t *testing.T) {
	current := map[string]device.DeviceInfo{
		"device1": {},
	}
	prev := map[string]device.DeviceInfo{}
	removed, added := Diff(prev, current)
	if len(removed) != 0 || len(added) != 1 {
		t.Errorf(`diff(%v, %v) = %v, %v, want %v, %v`, prev, current, removed, added, map[string]device.DeviceInfo{}, map[string]device.DeviceInfo{"device1": {}})
	}
}

// TestDiffNewDevice tests that a device has been removed
func TestDiffDeviceRemoved(t *testing.T) {
	current := map[string]device.DeviceInfo{}
	prev := map[string]device.DeviceInfo{
		"device1": {},
	}
	removed, added := Diff(prev, current)
	if len(removed) != 1 || len(added) != 0 {
		t.Errorf(`diff(%v, %v) = %v, %v, want %v, %v`, prev, current, removed, added, map[string]device.DeviceInfo{"device1": {}}, map[string]device.DeviceInfo{})
	}
}

// TestDiffSecondDeviceAdded tests that a second device has been added
func TestDiffSecondDeviceAdded(t *testing.T) {
	current := map[string]device.DeviceInfo{
		"device1": {}, "device2": {},
	}
	prev := map[string]device.DeviceInfo{
		"device1": {},
	}
	removed, added := Diff(prev, current)
	if len(removed) != 0 || len(added) != 1 {
		t.Errorf(`diff(%v, %v) = %v, %v, want %v, %v`, prev, current, removed, added, map[string]device.DeviceInfo{}, map[string]device.DeviceInfo{"device2": {}})
	}
}

// TestDiffSwapDevice tests that a device has been added and a device has been removed
func TestDiffSwapDevice(t *testing.T) {
	current := map[string]device.DeviceInfo{
		"device2": {},
	}
	prev := map[string]device.DeviceInfo{
		"device1": {},
	}
	removed, added := Diff(prev, current)
	if len(removed) != 1 || len(added) != 1 {
		t.Errorf(`diff(%v, %v) = %v, %v, want %v, %v`, prev, current, removed, added, map[string]device.DeviceInfo{"device1": {}}, map[string]device.DeviceInfo{"device2": {}})
	}
}
