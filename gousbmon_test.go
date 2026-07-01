package gousbmon

import (
	"errors"
	"log/slog"
	"testing"
	"time"
)

// mockDetector is a fake detector for testing.
type mockDetector struct {
	devices map[string]DeviceInfo
	err     error
}

func (m *mockDetector) GetAvailableDevices() (map[string]DeviceInfo, error) {
	return m.devices, m.err
}

// TestNew tests the public New constructor by overriding the platform detector.
func TestNew(t *testing.T) {
	old := newPlatformDetector
	defer func() { newPlatformDetector = old }()

	newPlatformDetector = func(_ *slog.Logger) (Detector, error) {
		return &mockDetector{
			devices: map[string]DeviceInfo{
				"dev1": {IDVendorID: "1234"},
			},
		}, nil
	}

	m, err := NewMonitor()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.lastCheck) != 1 {
		t.Errorf("len(lastCheck) = %d, want 1", len(m.lastCheck))
	}
}

// TestNew_Error tests that errors from the platform detector propagate through New.
func TestNew_Error(t *testing.T) {
	old := newPlatformDetector
	defer func() { newPlatformDetector = old }()

	newPlatformDetector = func(_ *slog.Logger) (Detector, error) {
		return nil, errors.New("gousbmon: platform not supported")
	}

	_, err := NewMonitor()
	if err.Error() != "gousbmon: platform not supported" {
		t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

// TestNewWithDetector_Success tests creating a monitor with a working detector.
func TestNewWithDetector_Success(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorID: "1234"},
		},
	}
	m, err := NewMonitor(WithDetector(d))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.lastCheck) != 1 {
		t.Errorf("len(lastCheck) = %d, want 1", len(m.lastCheck))
	}
}

// TestNewWithDetector_Error tests that detector errors propagate.
func TestNewWithDetector_Error(t *testing.T) {
	d := &mockDetector{err: errors.New("gousbmon: platform not supported")}
	_, err := NewMonitor(WithDetector(d))
	if err.Error() != "gousbmon: platform not supported" {
		t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

// TestGetAvailableDevices_Filtered tests that filters are applied.
func TestGetAvailableDevices_Filtered(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorID: "1234"},
			"dev2": {IDVendorID: "5678"},
		},
	}
	f := NewFilter().MatchVendorID("1234")
	m, err := NewMonitor(WithDetector(d), WithFilters(f))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByModelID tests filtering by model ID.
func TestGetAvailableDevices_FilterByModelID(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDModelID: "5678"},
			"dev2": {IDModelID: "9999"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchModelID("5678")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByVendor tests filtering by vendor name.
func TestGetAvailableDevices_FilterByVendor(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendor: "Acme"},
			"dev2": {IDVendor: "Other"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchVendor("Acme")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByModel tests filtering by model name.
func TestGetAvailableDevices_FilterByModel(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDModel: "Widget"},
			"dev2": {IDModel: "Gadget"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchModel("Widget")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterBySerial tests filtering by serial number.
func TestGetAvailableDevices_FilterBySerial(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDSerial: "ABC123"},
			"dev2": {IDSerial: "XYZ789"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchSerial("ABC123")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByUSBInterfaces tests filtering by USB interfaces.
func TestGetAvailableDevices_FilterByUSBInterfaces(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDUSBInterfaces: "HID"},
			"dev2": {IDUSBInterfaces: "MassStorage"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchUSBInterfaces("HID")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByRevision tests filtering by revision.
func TestGetAvailableDevices_FilterByRevision(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDRevision: "1.0"},
			"dev2": {IDRevision: "2.0"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchRevision("1.0")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByUSBClassFromDatabase tests filtering by USB class from database.
func TestGetAvailableDevices_FilterByUSBClassFromDatabase(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDUSBClassFromDatabase: "HID"},
			"dev2": {IDUSBClassFromDatabase: "Audio"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchUSBClassFromDatabase("HID")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByVendorFromDatabase tests filtering by vendor from database.
func TestGetAvailableDevices_FilterByVendorFromDatabase(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorFromDatabase: "AcmeCorp"},
			"dev2": {IDVendorFromDatabase: "OtherCorp"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchVendorFromDatabase("AcmeCorp")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByModelFromDatabase tests filtering by model from database.
func TestGetAvailableDevices_FilterByModelFromDatabase(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDModelFromDatabase: "WidgetDB"},
			"dev2": {IDModelFromDatabase: "GadgetDB"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchModelFromDatabase("WidgetDB")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByDevName tests filtering by device name.
func TestGetAvailableDevices_FilterByDevName(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {DevName: "usb0"},
			"dev2": {DevName: "usb1"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchDevName("usb0")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterByDevType tests filtering by device type.
func TestGetAvailableDevices_FilterByDevType(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {DevType: "usb_device"},
			"dev2": {DevType: "usb_interface"},
		},
	}
	m, err := NewMonitor(WithDetector(d), WithFilters(NewFilter().MatchDevType("usb_device")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterAND tests that chained criteria on a single Filter enforce AND logic.
func TestGetAvailableDevices_FilterAND(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorID: "1234", IDModelID: "5678"},
			"dev2": {IDVendorID: "1234", IDModelID: "9999"},
			"dev3": {IDVendorID: "0000", IDModelID: "5678"},
		},
	}
	f := NewFilter().MatchVendorID("1234").MatchModelID("5678")
	m, err := NewMonitor(WithDetector(d), WithFilters(f))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 1 {
		t.Errorf("len(devs) = %d, want 1", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
}

// TestGetAvailableDevices_FilterOR tests that multiple Filters enforce OR logic between them.
func TestGetAvailableDevices_FilterOR(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorID: "1234"},
			"dev2": {IDModelID: "5678"},
			"dev3": {IDVendorID: "9999", IDModelID: "9999"},
		},
	}
	f1 := NewFilter().MatchVendorID("1234")
	f2 := NewFilter().MatchModelID("5678")
	m, err := NewMonitor(WithDetector(d), WithFilters(f1, f2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 2 {
		t.Errorf("len(devs) = %d, want 2", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
	if _, ok := devs["dev2"]; !ok {
		t.Errorf("expected dev2 to be present")
	}
	if _, ok := devs["dev3"]; ok {
		t.Errorf("expected dev3 to be absent")
	}
}

// TestGetAvailableDevices_FilterANDOR tests a combination of AND criteria within Filters and OR between Filters.
func TestGetAvailableDevices_FilterANDOR(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorID: "1234", IDModelID: "5678"},
			"dev2": {IDVendorID: "1234", IDModelID: "9999"},
			"dev3": {IDVendorID: "18D1", IDUSBInterfaces: "HID"},
			"dev4": {IDVendorID: "18D1", IDUSBInterfaces: "MassStorage"},
		},
	}
	f1 := NewFilter().MatchVendorID("1234").MatchModelID("5678")
	f2 := NewFilter().MatchVendorID("18D1").MatchUSBInterfaces("HID")
	m, err := NewMonitor(WithDetector(d), WithFilters(f1, f2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	devs, err := m.GetAvailableDevices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devs) != 2 {
		t.Errorf("len(devs) = %d, want 2", len(devs))
	}
	if _, ok := devs["dev1"]; !ok {
		t.Errorf("expected dev1 to be present")
	}
	if _, ok := devs["dev3"]; !ok {
		t.Errorf("expected dev3 to be present")
	}
	if _, ok := devs["dev2"]; ok {
		t.Errorf("expected dev2 to be absent")
	}
	if _, ok := devs["dev4"]; ok {
		t.Errorf("expected dev4 to be absent")
	}
}

// TestChangesFromLastCheck_Added tests detecting a newly added device.
func TestChangesFromLastCheck_Added(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorID: "1234"},
		},
	}
	m, _ := NewMonitor(WithDetector(d))

	// Add a device
	d.devices = map[string]DeviceInfo{
		"dev1": {IDVendorID: "1234"},
		"dev2": {IDVendorID: "5678"},
	}
	removed, added, err := m.ChangesFromLastCheck(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(removed) != 0 {
		t.Errorf("len(removed) = %d, want 0", len(removed))
	}
	if len(added) != 1 {
		t.Errorf("len(added) = %d, want 1", len(added))
	}
	if _, ok := added["dev2"]; !ok {
		t.Errorf("expected dev2 to be added")
	}
}

// TestChangesFromLastCheck_Removed tests detecting a removed device.
func TestChangesFromLastCheck_Removed(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorID: "1234"},
			"dev2": {IDVendorID: "5678"},
		},
	}
	m, _ := NewMonitor(WithDetector(d))

	// Remove a device
	d.devices = map[string]DeviceInfo{
		"dev1": {IDVendorID: "1234"},
	}
	removed, added, err := m.ChangesFromLastCheck(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(removed) != 1 {
		t.Errorf("len(removed) = %d, want 1", len(removed))
	}
	if len(added) != 0 {
		t.Errorf("len(added) = %d, want 0", len(added))
	}
	if _, ok := removed["dev2"]; !ok {
		t.Errorf("expected dev2 to be removed")
	}
}

// TestCheckChanges_Callbacks tests that callbacks are invoked.
func TestCheckChanges_Callbacks(t *testing.T) {
	d := &mockDetector{
		devices: map[string]DeviceInfo{
			"dev1": {IDVendorID: "1234"},
		},
	}
	m, _ := NewMonitor(WithDetector(d))

	d.devices = map[string]DeviceInfo{
		"dev2": {IDVendorID: "5678"},
	}

	var connectedID, disconnectedID string
	onConnect := func(id string, info DeviceInfo) { connectedID = id }
	onDisconnect := func(id string, info DeviceInfo) { disconnectedID = id }

	err := m.CheckChanges(onConnect, onDisconnect, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connectedID != "dev2" {
		t.Errorf("connectedID = %q, want dev2", connectedID)
	}
	if disconnectedID != "dev1" {
		t.Errorf("disconnectedID = %q, want dev1", disconnectedID)
	}
}

// TestStartMonitoring_AlreadyRunning tests ErrAlreadyMonitoring.
func TestStartMonitoring_AlreadyRunning(t *testing.T) {
	d := &mockDetector{devices: map[string]DeviceInfo{}}
	m, _ := NewMonitor(WithDetector(d))

	if err := m.StartMonitoring(nil, nil); err != nil {
		t.Fatalf("first start failed: %v", err)
	}
	defer m.StopMonitoring()

	if err := m.StartMonitoring(nil, nil); err != errAlreadyMonitoring {
		t.Fatalf("expected ErrAlreadyMonitoring, got %v", err)
	}
}

// TestStartMonitoring_Stop tests that StopMonitoring cleanly shuts down.
func TestStartMonitoring_Stop(t *testing.T) {
	d := &mockDetector{devices: map[string]DeviceInfo{}}
	m, _ := NewMonitor(WithDetector(d))

	if err := m.StartMonitoring(nil, nil); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	m.StopMonitoring()

	// Restart should succeed after stop
	if err := m.StartMonitoring(nil, nil); err != nil {
		t.Fatalf("restart failed: %v", err)
	}
	m.StopMonitoring()
}

// TestChangesFromLastCheck_Error tests that detector errors are returned.
func TestChangesFromLastCheck_Error(t *testing.T) {
	d := &mockDetector{devices: map[string]DeviceInfo{}}
	m, _ := NewMonitor(WithDetector(d))

	d.err = errors.New("gousbmon: platform not supported")
	_, _, err := m.ChangesFromLastCheck(true)
	if err.Error() != "gousbmon: platform not supported" {
		t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

// TestCheckChanges_Error tests that CheckChanges propagates detector errors.
func TestCheckChanges_Error(t *testing.T) {
	d := &mockDetector{devices: map[string]DeviceInfo{}}
	m, _ := NewMonitor(WithDetector(d))

	d.err = errors.New("gousbmon: platform not supported")
	err := m.CheckChanges(nil, nil, true)
	if err.Error() != "gousbmon: platform not supported" {
		t.Fatalf("expected ErrUnsupportedPlatform, got %v", err)
	}
}

// TestStopMonitoring_NotRunning tests that StopMonitoring is safe when not running.
func TestStopMonitoring_NotRunning(t *testing.T) {
	d := &mockDetector{devices: map[string]DeviceInfo{}}
	m, _ := NewMonitor(WithDetector(d))

	// Should not panic or block
	m.StopMonitoring()
}

// TestStartMonitoring_WithInterval tests that monitoring starts with a custom interval.
func TestStartMonitoring_WithInterval(t *testing.T) {
	d := &mockDetector{devices: map[string]DeviceInfo{}}
	m, _ := NewMonitor(WithDetector(d))

	// Use a very short interval so the ticker fires quickly
	if err := m.StartMonitoringWithInterval(nil, nil, 10*time.Millisecond); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	m.StopMonitoring()
}

// TestStartMonitoring_ZeroInterval tests for an invalid interval.
func TestStartMonitoring_ZeroInterval(t *testing.T) {
	d := &mockDetector{devices: map[string]DeviceInfo{}}
	m, _ := NewMonitor(WithDetector(d))

	if err := m.StartMonitoringWithInterval(nil, nil, 0); err != errInvalidInterval {
		t.Fatalf("expected ErrInvalidInterval, got %v", err)
	}
	m.StopMonitoring()
}

// TestStartMonitoring_TickerFires tests that the background goroutine ticks and runs CheckChanges.
func TestStartMonitoring_TickerFires(t *testing.T) {
	d := &mockDetector{devices: map[string]DeviceInfo{}}
	m, _ := NewMonitor(WithDetector(d))

	if err := m.StartMonitoringWithInterval(nil, nil, 10*time.Millisecond); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	// Wait for the ticker to fire at least once
	time.Sleep(50 * time.Millisecond)

	m.StopMonitoring()
}

// TestApplyFilters_NoFilter tests that devices pass through when no filters are provided.
func TestApplyFilters_NoFilter(t *testing.T) {
	devices := map[string]DeviceInfo{"device1": {IDVendorID: "1234"}}
	result := applyFilters(devices, [][]filter{})
	if len(result) != 1 {
		t.Errorf("applyFilters(%v, %v) = %v, want %v", devices, [][]filter{}, result, devices)
	}
}

// TestApplyFilters_OneMatchingFilter tests that a matching predicate keeps the device.
func TestApplyFilters_OneMatchingFilter(t *testing.T) {
	devices := map[string]DeviceInfo{"device1": {IDVendorID: "1234"}}
	f := NewFilter().MatchVendorID("1234")
	filterGroups := [][]filter{f.filters}
	result := applyFilters(devices, filterGroups)
	if len(result) != 1 {
		t.Errorf("applyFilters(%v, %v) = %v, want %v", devices, filterGroups, result, devices)
	}
}

// TestApplyFilters_OneNonMatchingFilter tests that a non-matching predicate removes the device.
func TestApplyFilters_OneNonMatchingFilter(t *testing.T) {
	devices := map[string]DeviceInfo{"device1": {IDVendorID: "1234"}}
	f := NewFilter().MatchVendorID("5678")
	filterGroups := [][]filter{f.filters}
	result := applyFilters(devices, filterGroups)
	if len(result) != 0 {
		t.Errorf("applyFilters(%v, %v) = %v, want %v", devices, filterGroups, result, map[string]DeviceInfo{})
	}
}
