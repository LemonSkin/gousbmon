package gousbmon

import (
	"testing"
)

// TestFilter_NewFilter tests creating a new Filter.
func TestFilter_NewFilter(t *testing.T) {
	f := NewFilter()
	if f == nil {
		t.Fatal("NewFilter() returned nil")
	}
	if len(f.filters) != 0 {
		t.Errorf("NewFilter() has %d filters, want 0", len(f.filters))
	}
}

// TestFilter_MatchVendorID tests the MatchVendorID builder method.
func TestFilter_MatchVendorID(t *testing.T) {
	f := NewFilter()
	f.MatchVendorID("1234")
	if len(f.filters) != 1 {
		t.Errorf("After MatchVendorID, filter count = %d, want 1", len(f.filters))
	}
	info := DeviceInfo{IDVendorID: "1234"}
	if !f.filters[0](info) {
		t.Errorf("Filter did not match device with IDVendorID=1234")
	}
}

// TestFilter_Chaining tests method chaining on Filter.
func TestFilter_Chaining(t *testing.T) {
	f := NewFilter()
	f.MatchVendorID("1234").MatchModelID("5678").MatchVendor("Acme")
	if len(f.filters) != 3 {
		t.Errorf("After chaining 3 methods, filter count = %d, want 3", len(f.filters))
	}
}

// TestFilter_MatchModelID tests the MatchModelID builder method.
func TestFilter_MatchModelID(t *testing.T) {
	f := NewFilter().MatchModelID("5678")
	if len(f.filters) != 1 {
		t.Errorf("After MatchModelID, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDModelID: "5678"}) {
		t.Errorf("MatchModelID did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDModelID: "9999"}) {
		t.Errorf("MatchModelID matched unexpected device")
	}
}

// TestFilter_MatchVendor tests the MatchVendor builder method.
func TestFilter_MatchVendor(t *testing.T) {
	f := NewFilter().MatchVendor("Acme")
	if len(f.filters) != 1 {
		t.Errorf("After MatchVendor, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDVendor: "Acme"}) {
		t.Errorf("MatchVendor did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDVendor: "Other"}) {
		t.Errorf("MatchVendor matched unexpected device")
	}
}

// TestFilter_MatchModel tests the MatchModel builder method.
func TestFilter_MatchModel(t *testing.T) {
	f := NewFilter().MatchModel("Widget")
	if len(f.filters) != 1 {
		t.Errorf("After MatchModel, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDModel: "Widget"}) {
		t.Errorf("MatchModel did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDModel: "Gadget"}) {
		t.Errorf("MatchModel matched unexpected device")
	}
}

// TestFilter_MatchSerial tests the MatchSerial builder method.
func TestFilter_MatchSerial(t *testing.T) {
	f := NewFilter().MatchSerial("ABC123")
	if len(f.filters) != 1 {
		t.Errorf("After MatchSerial, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDSerial: "ABC123"}) {
		t.Errorf("MatchSerial did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDSerial: "XYZ789"}) {
		t.Errorf("MatchSerial matched unexpected device")
	}
}

// TestFilter_MatchUSBInterfaces tests the MatchUSBInterfaces builder method.
func TestFilter_MatchUSBInterfaces(t *testing.T) {
	f := NewFilter().MatchUSBInterfaces("HID")
	if len(f.filters) != 1 {
		t.Errorf("After MatchUSBInterfaces, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDUSBInterfaces: "HID"}) {
		t.Errorf("MatchUSBInterfaces did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDUSBInterfaces: "MassStorage"}) {
		t.Errorf("MatchUSBInterfaces matched unexpected device")
	}
}

// TestFilter_MatchRevision tests the MatchRevision builder method.
func TestFilter_MatchRevision(t *testing.T) {
	f := NewFilter().MatchRevision("1.0")
	if len(f.filters) != 1 {
		t.Errorf("After MatchRevision, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDRevision: "1.0"}) {
		t.Errorf("MatchRevision did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDRevision: "2.0"}) {
		t.Errorf("MatchRevision matched unexpected device")
	}
}

// TestFilter_MatchUSBClassFromDatabase tests the MatchUSBClassFromDatabase builder method.
func TestFilter_MatchUSBClassFromDatabase(t *testing.T) {
	f := NewFilter().MatchUSBClassFromDatabase("HID")
	if len(f.filters) != 1 {
		t.Errorf("After MatchUSBClassFromDatabase, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDUSBClassFromDatabase: "HID"}) {
		t.Errorf("MatchUSBClassFromDatabase did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDUSBClassFromDatabase: "Audio"}) {
		t.Errorf("MatchUSBClassFromDatabase matched unexpected device")
	}
}

// TestFilter_MatchVendorFromDatabase tests the MatchVendorFromDatabase builder method.
func TestFilter_MatchVendorFromDatabase(t *testing.T) {
	f := NewFilter().MatchVendorFromDatabase("AcmeCorp")
	if len(f.filters) != 1 {
		t.Errorf("After MatchVendorFromDatabase, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDVendorFromDatabase: "AcmeCorp"}) {
		t.Errorf("MatchVendorFromDatabase did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDVendorFromDatabase: "OtherCorp"}) {
		t.Errorf("MatchVendorFromDatabase matched unexpected device")
	}
}

// TestFilter_MatchModelFromDatabase tests the MatchModelFromDatabase builder method.
func TestFilter_MatchModelFromDatabase(t *testing.T) {
	f := NewFilter().MatchModelFromDatabase("WidgetDB")
	if len(f.filters) != 1 {
		t.Errorf("After MatchModelFromDatabase, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{IDModelFromDatabase: "WidgetDB"}) {
		t.Errorf("MatchModelFromDatabase did not match expected device")
	}
	if f.filters[0](DeviceInfo{IDModelFromDatabase: "GadgetDB"}) {
		t.Errorf("MatchModelFromDatabase matched unexpected device")
	}
}

// TestFilter_MatchDevName tests the MatchDevName builder method.
func TestFilter_MatchDevName(t *testing.T) {
	f := NewFilter().MatchDevName("usb0")
	if len(f.filters) != 1 {
		t.Errorf("After MatchDevName, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{DevName: "usb0"}) {
		t.Errorf("MatchDevName did not match expected device")
	}
	if f.filters[0](DeviceInfo{DevName: "usb1"}) {
		t.Errorf("MatchDevName matched unexpected device")
	}
}

// TestFilter_MatchDevType tests the MatchDevType builder method.
func TestFilter_MatchDevType(t *testing.T) {
	f := NewFilter().MatchDevType("usb_device")
	if len(f.filters) != 1 {
		t.Errorf("After MatchDevType, filter count = %d, want 1", len(f.filters))
	}
	if !f.filters[0](DeviceInfo{DevType: "usb_device"}) {
		t.Errorf("MatchDevType did not match expected device")
	}
	if f.filters[0](DeviceInfo{DevType: "usb_interface"}) {
		t.Errorf("MatchDevType matched unexpected device")
	}
}
