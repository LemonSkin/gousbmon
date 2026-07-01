# GoUSBMonitor

GoUSBMonitor is a cross-platform library for USB device monitoring that simplifies tracking of connections, disconnections, and examination of connected device attributes on Windows, Linux and macOS (eventually).

It is heavily inspired by the [USBMonitor](https://github.com/Lemon-Skin/USBMonitor) library for Python.

Currently it supports Windows and Linux as that is what I have needed it for, but I plan on adding support for macOS at some point as required.

I built this because other solutions are either single-platform ([USBMon](https://github.com/rubiojr/go-usbmon)) or require additional dependencies and/or do more than I need ([gousb](https://github.com/google/gousb)).

Pull requests, bug reports and feature requests are very welcome.

## Installation

```bash
go get github.com/LemonSkin/gousbmon
```

## Usage

GoUSBMonitor has been designed to be very reminiscent of the [USBMonitor](https://github.com/Lemon-Skin/USBMonitor) library for Python. See the [example code](./examples/basic/main.go) for basic setup and usage.

### Filtering devices

You can restrict which devices the monitor reports by passing `Filter` objects to `gousbmon.WithFilters`. A device is kept if it matches **any** of the provided filters. Chain methods on a single `Filter` to require **all** of those conditions.

For example, the following monitor will report devices that **either** have a vendor ID of `046d` **or** have a vendor ID of `1234` and a model ID of `5678`.


```go
f1 := gousbmon.NewFilter().MatchVendorID("046d")
f2 := gousbmon.NewFilter().MatchVendorID("1234").MatchModelID("5678")
monitor, err := gousbmon.NewMonitor(gousbmon.WithFilters(f1, f2))
```


## API Reference

### `gousbmon.NewMonitor(opts ...Option) (*Monitor, error)`

Creates a `Monitor` for the current platform. Options configure the detector, logger, handler, and filters.

- `opts`: `...Option` - Optional configuration options. See the [Options](#options) below.
- Returns `*Monitor` and `error`.

### Options

- `gousbmon.WithDetector(d device.Detector) Option` - Use a custom `Detector` instead of the platform-specific one.
- `gousbmon.WithLogger(l *slog.Logger) Option` - Set the logger used for diagnostic output.
- `gousbmon.WithHandler(h slog.Handler) Option` - Set the logger from an `slog.Handler`.
- `gousbmon.WithFilters(filters ...*Filter) Option` - Restrict monitored devices to those matching the supplied filters.

### `(*Monitor) StartMonitoring(onConnect, onDisconnect func(deviceID string, info DeviceInfo)) error`

Starts a background goroutine that polls for device changes, invoking `onConnect`/`onDisconnect` as devices appear and disappear. Polls every 500ms.

- `onConnect`: `func(deviceID string, info DeviceInfo)` - Invoked when a device is connected.
- `onDisconnect`: `func(deviceID string, info DeviceInfo)` - Invoked when a device is disconnected.
- Returns `ErrAlreadyMonitoring` if monitoring is already running.

### `(*Monitor) StartMonitoringWithInterval(onConnect, onDisconnect func(deviceID string, info DeviceInfo), interval time.Duration) error`

Starts a background goroutine that polls for device changes at the given interval, invoking `onConnect`/`onDisconnect` as devices appear and disappear.

- `onConnect`: `func(deviceID string, info DeviceInfo)` - Invoked when a device is connected.
- `onDisconnect`: `func(deviceID string, info DeviceInfo)` - Invoked when a device is disconnected.
- `interval`: `time.Duration` - Polling interval.
- Returns `ErrInvalidInterval` if the interval is invalid.
- Returns `ErrAlreadyMonitoring` if monitoring is already running.

### `(*Monitor) StopMonitoring()`

Stops the background monitoring goroutine.

### `(*Monitor) GetAvailableDevices() (map[string]DeviceInfo, error)`

Returns the currently connected devices keyed by device ID, after applying any configured filters.

### `(*Monitor) ChangesFromLastCheck(update bool) (removed, added map[string]DeviceInfo, err error)`

Returns the devices removed and added since the previous check. When `update` is `true`, the internal snapshot is saved to the current state.

### `(*Monitor) CheckChanges(onConnect, onDisconnect func(deviceID string, info DeviceInfo), update bool) error`

Runs a single change-detection pass, invoking the callbacks for each removed and added device. When `update` is `true`, the internal snapshot is saved to the current state.

### `Filter` builder

Create a `Filter` with `gousbmon.NewFilter()` and chain methods to add criteria. All criteria on a single `Filter` must match (AND logic). Pass multiple filters to `WithFilters` for OR logic.

- `(*Filter) MatchVendorID(id string) *Filter`
- `(*Filter) MatchModelID(id string) *Filter`
- `(*Filter) MatchVendor(name string) *Filter`
- `(*Filter) MatchModel(name string) *Filter`
- `(*Filter) MatchSerial(serial string) *Filter`
- `(*Filter) MatchUSBInterfaces(interfaces string) *Filter`
- `(*Filter) MatchRevision(revision string) *Filter`
- `(*Filter) MatchUSBClassFromDatabase(class string) *Filter`
- `(*Filter) MatchVendorFromDatabase(vendor string) *Filter`
- `(*Filter) MatchModelFromDatabase(model string) *Filter`
- `(*Filter) MatchDevName(name string) *Filter`
- `(*Filter) MatchDevType(devType string) *Filter`

## Device Properties

The `DeviceInfo` struct returned by most functions contains the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `IDVendorID` | `string` | Vendor ID (e.g. `"046d"`) |
| `IDVendor` | `string` | Vendor name |
| `IDModel` | `string` | Model name |
| `IDModelID` | `string` | Model ID (e.g. `"c077"`) |
| `IDSerial` | `string` | Serial number |
| `IDUSBInterfaces` | `string` | USB interfaces (comma-separated) |
| `IDRevision` | `string` | Device revision |
| `IDUSBClassFromDatabase` | `string` | USB class from database |
| `IDVendorFromDatabase` | `string` | Vendor from database |
| `IDModelFromDatabase` | `string` | Model from database |
| `DevName` | `string` | Device name/path |
| `DevType` | `string` | Device type (e.g. `"usb_device"`) |

Note that, depending on the device and the OS, some of this information may be incomplete or certain attributes may overlap with others.

## About

GoUSBMonitor is a Go port of the Python [USBMonitor](https://github.com/Eric-Canas/USBMonitor) library.

