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
