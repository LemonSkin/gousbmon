// Package gousbmon is a cross-platform library for monitoring USB device connections and disconnections. It is a Go
// port of the Python USBMonitor library and exposes the same Linux-style device attributes on every platform.
package gousbmon

import (
	"sync"
	"time"

	"github.com/LemonSkin/gousbmon/device"
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
)

// Detector produces the raw set of connected USB devices. Provide a custom implementation to NewWithDetector,
// e.g. for testing.
type Detector = device.Detector

// Callback is invoked with the device ID and its information when a device is connected or disconnected.
type Callback func(deviceID string, info device.Info)

// Option configures a monitoring session started by StartMonitoring.
type Option func(*monitorConfig)

type monitorConfig struct {
	interval time.Duration
}

// WithInterval sets the polling interval. Values <= 0 use DefaultCheckInterval (500ms).
func WithInterval(d time.Duration) Option {
	return func(c *monitorConfig) { c.interval = d }
}

// Monitor inspects and monitors the USB devices connected to the system.
type Monitor struct {
	detector device.Detector
	filters  []Filter

	mu        sync.Mutex
	lastCheck map[string]device.Info

	threadMu sync.Mutex
	stop     chan struct{}
	wg       sync.WaitGroup
}

// New creates a Monitor for the current platform. Optional filters restrict the devices that are reported and
// monitored.
func New(filters ...Filter) (*Monitor, error) {
	detector, err := newPlatformDetector()
	if err != nil {
		return nil, err
	}
	return NewWithDetector(detector, filters...)
}

// NewWithDetector creates a Monitor backed by a caller-supplied Detector, bypassing platform detection.
// Mostly used for testing or for users who want to provide their own custom backend.
func NewWithDetector(detector Detector, filters ...Filter) (*Monitor, error) {
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

	removed, added = diff(prev, current)
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
func (m *Monitor) StartMonitoring(onConnect, onDisconnect Callback, opts ...Option) error {
	cfg := monitorConfig{interval: DefaultCheckInterval}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.interval <= 0 {
		cfg.interval = DefaultCheckInterval
	}
	interval := cfg.interval

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
