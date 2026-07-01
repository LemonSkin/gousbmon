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
