// Package gousbmon is a cross-platform library for monitoring USB device connections and disconnections. It is a Go
// port of the Python USBMonitor library and exposes the same Linux-style device attributes on every platform.
package gousbmon

import (
	"sync"
	"time"

	"github.com/LemonSkin/gousbmon/device"
	"github.com/LemonSkin/gousbmon/filter"
	"github.com/LemonSkin/gousbmon/internal"
	"github.com/LemonSkin/gousbmon/internal/errors"
	"github.com/LemonSkin/gousbmon/internal/platform"
)

const DefaultCheckInterval = 500 * time.Millisecond

// newPlatformDetector is the function New uses to obtain a platform detector. Made overridable here for testing.
var newPlatformDetector = platform.New

// Re-exported errors for end-users.
var (
	// ErrAlreadyMonitoring is returned when StartMonitoring is called while a monitoring goroutine is already running.
	ErrAlreadyMonitoring = errors.ErrAlreadyMonitoring
	// ErrUnsupportedPlatform is returned by New on a platform without a backend.
	ErrUnsupportedPlatform = errors.ErrUnsupportedPlatform
	// ErrInvalidInterval is returned when an invalid interval is provided to StartMonitoringWithInterval.
	ErrInvalidInterval = errors.ErrInvalidInterval
)

// Detector produces the raw set of connected USB devices. Provide a custom implementation to NewWithDetector,
// e.g. for testing.
type Detector = device.Detector

// Callback is invoked with the device ID and its information when a device is connected or disconnected.
type Callback func(deviceID string, info device.Info)

// Monitor inspects and monitors the USB devices connected to the system.
type Monitor struct {
	detector device.Detector
	filters  []filter.Filter

	mu        sync.Mutex
	lastCheck map[string]device.Info

	threadMu sync.Mutex
	stop     chan struct{}
	wg       sync.WaitGroup
}

// New creates a Monitor for the current platform. Optional filters restrict the devices that are reported and
// monitored.
func New(filters ...filter.Filter) (*Monitor, error) {
	detector, err := newPlatformDetector()
	if err != nil {
		return nil, err
	}
	return NewWithDetector(detector, filters...)
}

// NewWithDetector creates a Monitor backed by a caller-supplied Detector, bypassing platform detection.
// Mostly used for testing or for users who want to provide their own custom backend.
func NewWithDetector(detector Detector, filters ...filter.Filter) (*Monitor, error) {
	m := &Monitor{detector: detector, filters: filters}
	devices, err := m.GetAvailableDevices()
	if err != nil {
		return nil, err
	}
	m.lastCheck = devices
	return m, nil
}

// GetAvailableDevices returns the currently connected devices. The map key is the device's ID reported to the system.
func (m *Monitor) GetAvailableDevices() (map[string]device.Info, error) {
	devices, err := m.detector.GetAvailableDevices()
	if err != nil {
		return nil, err
	}
	return applyFilters(devices, m.filters), nil
}

// ChangesFromLastCheck returns the devices removed and added since the last check.
// When update is true, the internal snapshot is saved to the current state.
func (m *Monitor) ChangesFromLastCheck(update bool) (removed, added map[string]device.Info, err error) {
	current, err := m.GetAvailableDevices()
	if err != nil {
		return nil, nil, err
	}

	m.mu.Lock()
	prev := m.lastCheck
	if update {
		m.lastCheck = current
	}
	m.mu.Unlock()

	removed, added = internal.Diff(prev, current)
	return removed, added, nil
}

// CheckChanges runs a single check for changes, invoking the callbacks for each removed and added device.
// When update is true, the internal snapshot is saved to the current state.
func (m *Monitor) CheckChanges(onConnect, onDisconnect Callback, update bool) error {
	removed, added, err := m.ChangesFromLastCheck(update)
	if err != nil {
		return err
	}
	if onDisconnect != nil {
		for id, info := range removed {
			onDisconnect(id, info)
		}
	}
	if onConnect != nil {
		for id, info := range added {
			onConnect(id, info)
		}
	}
	return nil
}

// StartMonitoring starts a background goroutine that polls for device changes, invoking callbacks as devices appear and
// disappear. By default it polls every DefaultCheckInterval (500ms); use WithInterval to override.
func (m *Monitor) StartMonitoring(onConnect, onDisconnect Callback) error {
	interval := DefaultCheckInterval

	m.threadMu.Lock()
	defer m.threadMu.Unlock()
	if m.stop != nil {
		return ErrAlreadyMonitoring
	}
	stop := make(chan struct{})
	m.stop = stop
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				_ = m.CheckChanges(onConnect, onDisconnect, true)
			}
		}
	}()
	return nil
}

// StartMonitoringWithInterval starts a background goroutine that polls for device changes every given interval, invoking callbacks as devices appear and
// disappear.
func (m *Monitor) StartMonitoringWithInterval(onConnect, onDisconnect Callback, interval time.Duration) error {
	if interval <= 0 {
		return ErrInvalidInterval
	}

	m.threadMu.Lock()
	defer m.threadMu.Unlock()
	if m.stop != nil {
		return ErrAlreadyMonitoring
	}
	stop := make(chan struct{})
	m.stop = stop
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				_ = m.CheckChanges(onConnect, onDisconnect, true)
			}
		}
	}()
	return nil
}

// StopMonitoring stops the background monitoring goroutine.
func (m *Monitor) StopMonitoring() {
	m.threadMu.Lock()
	stop := m.stop
	m.stop = nil
	m.threadMu.Unlock()
	if stop == nil {
		return
	}
	close(stop)
	m.wg.Wait()
}

// applyFilters keeps only the devices that match at least one of the filters. If no filters are provided, devices is
// returned unchanged.
func applyFilters(devices map[string]device.Info, filters []filter.Filter) map[string]device.Info {
	if len(filters) == 0 {
		return devices
	}
	out := make(map[string]device.Info, len(devices))
	for id, info := range devices {
		for _, f := range filters {
			if f(info) {
				out[id] = info
				break
			}
		}
	}
	return out
}
