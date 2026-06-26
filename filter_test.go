package gousbmon

import (
	"testing"

	"github.com/LemonSkin/gousbmon/device"
)

// TestApplyFilters_NoFilter tests that devices pass through when no filters are provided.
func TestApplyFilters_NoFilter(t *testing.T) {
	devices := map[string]device.Info{"device1": {IDVendorID: "1234"}}
	result := applyFilters(devices, []Filter{})
	if len(result) != 1 {
		t.Errorf("applyFilters(%v, %v) = %v, want %v", devices, []Filter{}, result, devices)
	}
}

// TestApplyFilters_OneMatchingFilter tests that a matching predicate keeps the device.
func TestApplyFilters_OneMatchingFilter(t *testing.T) {
	devices := map[string]device.Info{"device1": {IDVendorID: "1234"}}
	filter := MatchVendorID("1234")
	result := applyFilters(devices, []Filter{filter})
	if len(result) != 1 {
		t.Errorf("applyFilters(%v, %v) = %v, want %v", devices, []Filter{filter}, result, devices)
	}
}

// TestApplyFilters_OneNonMatchingFilter tests that a non-matching predicate removes the device.
func TestApplyFilters_OneNonMatchingFilter(t *testing.T) {
	devices := map[string]device.Info{"device1": {IDVendorID: "1234"}}
	filter := MatchVendorID("5678")
	result := applyFilters(devices, []Filter{filter})
	if len(result) != 0 {
		t.Errorf("applyFilters(%v, %v) = %v, want %v", devices, []Filter{filter}, result, map[string]device.Info{})
	}
}

// TestMatchAll_Match tests that MatchAll passes when all predicates match.
func TestMatchAll_Match(t *testing.T) {
	info := device.Info{IDVendorID: "1234", IDModelID: "5678"}
	filter := MatchAll(MatchVendorID("1234"), MatchModelID("5678"))
	if !filter(info) {
		t.Errorf("MatchAll(MatchVendorID(\"1234\"), MatchModelID(\"5678\"))(%v) = false, want true", info)
	}
}

// TestMatchAll_NoMatch tests that MatchAll fails when any predicate fails.
func TestMatchAll_NoMatch(t *testing.T) {
	info := device.Info{IDVendorID: "1234", IDModelID: "5678"}
	filter := MatchAll(MatchVendorID("5678"), MatchModelID("5678"))
	if filter(info) {
		t.Errorf("MatchAll(MatchVendorID(\"5678\"), MatchModelID(\"5678\"))(%v) = true, want false", info)
	}
}

// TestMatchVendorID tests the MatchVendorID helper.
func TestMatchVendorID(t *testing.T) {
	info := device.Info{IDVendorID: "1234"}
	if !MatchVendorID("1234")(info) {
		t.Errorf("MatchVendorID(\"1234\")(%v) = false, want true", info)
	}
}

// TestMatchModelID tests the MatchModelID helper.
func TestMatchModelID(t *testing.T) {
	info := device.Info{IDModelID: "5678"}
	if !MatchModelID("5678")(info) {
		t.Errorf("MatchModelID(\"5678\")(%v) = false, want true", info)
	}
}

// TestMatchVendor tests the MatchVendor helper.
func TestMatchVendor(t *testing.T) {
	info := device.Info{IDVendor: "Acme"}
	if !MatchVendor("Acme")(info) {
		t.Errorf("MatchVendor(\"Acme\")(%v) = false, want true", info)
	}
}

// TestMatchModel tests the MatchModel helper.
func TestMatchModel(t *testing.T) {
	info := device.Info{IDModel: "Widget"}
	if !MatchModel("Widget")(info) {
		t.Errorf("MatchModel(\"Widget\")(%v) = false, want true", info)
	}
}

// TestMatchSerial tests the MatchSerial helper.
func TestMatchSerial(t *testing.T) {
	info := device.Info{IDSerial: "ABC123"}
	if !MatchSerial("ABC123")(info) {
		t.Errorf("MatchSerial(\"ABC123\")(%v) = false, want true", info)
	}
}

// TestMatchUSBInterfaces tests the MatchUSBInterfaces helper.
func TestMatchUSBInterfaces(t *testing.T) {
	info := device.Info{IDUSBInterfaces: "iface0"}
	if !MatchUSBInterfaces("iface0")(info) {
		t.Errorf("MatchUSBInterfaces(\"iface0\")(%v) = false, want true", info)
	}
}

// TestMatchRevision tests the MatchRevision helper.
func TestMatchRevision(t *testing.T) {
	info := device.Info{IDRevision: "1.0"}
	if !MatchRevision("1.0")(info) {
		t.Errorf("MatchRevision(\"1.0\")(%v) = false, want true", info)
	}
}

// TestMatchUSBClassFromDatabase tests the MatchUSBClassFromDatabase helper.
func TestMatchUSBClassFromDatabase(t *testing.T) {
	info := device.Info{IDUSBClassFromDatabase: "HID"}
	if !MatchUSBClassFromDatabase("HID")(info) {
		t.Errorf("MatchUSBClassFromDatabase(\"HID\")(%v) = false, want true", info)
	}
}

// TestMatchVendorFromDatabase tests the MatchVendorFromDatabase helper.
func TestMatchVendorFromDatabase(t *testing.T) {
	info := device.Info{IDVendorFromDatabase: "AcmeCorp"}
	if !MatchVendorFromDatabase("AcmeCorp")(info) {
		t.Errorf("MatchVendorFromDatabase(\"AcmeCorp\")(%v) = false, want true", info)
	}
}

// TestMatchModelFromDatabase tests the MatchModelFromDatabase helper.
func TestMatchModelFromDatabase(t *testing.T) {
	info := device.Info{IDModelFromDatabase: "WidgetDB"}
	if !MatchModelFromDatabase("WidgetDB")(info) {
		t.Errorf("MatchModelFromDatabase(\"WidgetDB\")(%v) = false, want true", info)
	}
}

// TestMatchDevName tests the MatchDevName helper.
func TestMatchDevName(t *testing.T) {
	info := device.Info{DevName: "usb0"}
	if !MatchDevName("usb0")(info) {
		t.Errorf("MatchDevName(\"usb0\")(%v) = false, want true", info)
	}
}

// TestMatchDevType tests the MatchDevType helper.
func TestMatchDevType(t *testing.T) {
	info := device.Info{DevType: "usb_device"}
	if !MatchDevType("usb_device")(info) {
		t.Errorf("MatchDevType(\"usb_device\")(%v) = false, want true", info)
	}
}
