// Package gousbmon is a cross-platform library for monitoring USB device connections and disconnections. It is a Go
// port of the Python USBMonitor library and exposes the same Linux-style device attributes on every platform.
package gousbmon

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/LemonSkin/gousbmon/device"
	"github.com/LemonSkin/gousbmon/internal"
	"github.com/LemonSkin/gousbmon/internal/platform"
)

const DefaultCheckInterval = 500 * time.Millisecond

var errAlreadyMonitoring = errors.New("gousbmon: monitoring is already running")
var errInvalidInterval = errors.New("gousbmon: invalid interval")

// newPlatformDetector is the function New uses to obtain a platform detector. Made overridable here for testing.
var newPlatformDetector = platform.New

type DeviceInfo = device.DeviceInfo
type Detector = device.Detector

// Monitor inspects and monitors the USB devices connected to the system.
type Monitor struct {
	detector     Detector
	filterGroups [][]filter
	logger       *slog.Logger

	mu        sync.Mutex
	lastCheck map[string]DeviceInfo

	threadMu sync.Mutex
	stop     chan struct{}
	wg       sync.WaitGroup
}

// NewMonitor creates a Monitor for the current platform. Configuration options include WithDetector, WithFilter, WithLogger, and WithHandler.
func NewMonitor(opts ...Option) (*Monitor, error) {
	cfg := newConfig(opts)
	var detector device.Detector
	var err error
	if cfg.detector != nil {
		detector = cfg.detector
	} else {
		detector, err = newPlatformDetector(cfg.logger)
		if err != nil {
			return nil, err
		}
	}

	m := &Monitor{detector: detector, filterGroups: cfg.filterGroups, logger: cfg.logger}
	devices, err := m.detector.GetAvailableDevices()
	if err != nil {
		return nil, err
	}
	m.lastCheck = devices
	return m, nil

}

// GetAvailableDevices returns the currently connected devices. The map key is the device's ID reported to the system.
func (m *Monitor) GetAvailableDevices() (map[string]DeviceInfo, error) {
	devices, err := m.detector.GetAvailableDevices()
	if err != nil {
		return nil, err
	}
	return applyFilters(devices, m.filterGroups), nil
}

// ChangesFromLastCheck returns the devices removed and added since the last check.
// When update is true, the internal snapshot is saved to the current state.
func (m *Monitor) ChangesFromLastCheck(update bool) (removed, added map[string]DeviceInfo, err error) {
	current, err := m.detector.GetAvailableDevices()
	if err != nil {
		return nil, nil, err
	}
	current = applyFilters(current, m.filterGroups)

	m.mu.Lock()
	prev := m.lastCheck
	if update {
		m.lastCheck = current
	}
	m.mu.Unlock()

	removedInternal, addedInternal := internal.Diff(prev, current)
	return removedInternal, addedInternal, nil
}

// CheckChanges runs a single check for changes, invoking the callbacks for each removed and added device.
// When update is true, the internal snapshot is saved to the current state.
func (m *Monitor) CheckChanges(onConnect, onDisconnect func(deviceID string, info DeviceInfo), update bool) error {
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
func (m *Monitor) StartMonitoring(onConnect, onDisconnect func(deviceID string, info DeviceInfo)) error {
	interval := DefaultCheckInterval

	m.threadMu.Lock()
	defer m.threadMu.Unlock()
	if m.stop != nil {
		return errAlreadyMonitoring
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
func (m *Monitor) StartMonitoringWithInterval(onConnect, onDisconnect func(deviceID string, info DeviceInfo), interval time.Duration) error {
	if interval <= 0 {
		return errInvalidInterval
	}

	m.threadMu.Lock()
	defer m.threadMu.Unlock()
	if m.stop != nil {
		return errAlreadyMonitoring
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

// applyFilters keeps only the devices that match at least one filter group. If no filter groups are provided, devices
// is returned unchanged. Within each group, all filters must match (AND). Between groups, any group may match (OR).
func applyFilters(devices map[string]device.DeviceInfo, filterGroups [][]filter) map[string]device.DeviceInfo {
	if len(filterGroups) == 0 {
		return devices
	}
	out := make(map[string]device.DeviceInfo, len(devices))
	for id, info := range devices {
		// A device matches if ANY group matches (OR logic)
		for _, group := range filterGroups {
			if matchesAll(info, group) {
				out[id] = info
				break
			}
		}
	}
	return out
}

// matchesAll reports whether all filters in the group match the device.
func matchesAll(info DeviceInfo, filters []filter) bool {
	for _, f := range filters {
		if !f(info) {
			return false
		}
	}
	return true
}
