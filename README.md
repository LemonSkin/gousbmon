# GoUSBMonitor

GoUSBMonitor is a cross-platform (eventually) library for USB device monitoring that simplifies tracking of connections, disconnections, and examination of connected device attributes on Windows, Linux and macOS (eventually).

It is heavily inspired by the [USBMonitor](https://github.com/Lemon-Skin/USBMonitor) library for Python.

Currently it only supports Windows as that is what I have needed it for, but I plan on adding support for Linux in the near future and macOS at some point, probably in the far future.

I built this because other solutions are either single-platform ([USBMon](https://github.com/rubiojr/go-usbmon)) or require additional dependencies and/or do more than I need ([gousb](https://github.com/google/gousb)).

Pull requests, bug reports and feature requests are very welcome.

## Installation

```bash
go get github.com/LemonSkin/gousbmon
```

## Usage

GoUSBMonitor has been designed to be very reminiscent of the [USBMonitor](https://github.com/Lemon-Skin/USBMonitor) library for Python. See the [example code](./examples/basic/main.go) for basic setup and usage.

### Filtering devices

You can restrict which devices the monitor reports by passing filter predicates to `New`. A device is kept if it matches **any** of the provided filters. Use `MatchAll` to require **all** conditions within a single filter.

For example, the following monitor will report devices that **either** have a vendor ID of `046d` (Logitech) **or** simultaneously have a vendor ID of `1234` and a model ID of `5678`.


```go
monitor, err := gousbmon.New(
	gousbmon.MatchVendorID("046d"),
	gousbmon.MatchAll(
		gousbmon.MatchVendorID("1234"),
		gousbmon.MatchModelID("5678"),
	),
)
```


## API Reference

### `gousbmon.New(filters ...Filter) (*Monitor, error)`

Creates a `Monitor` for the current platform. Optional filters restrict the devices that are reported and monitored. A device is kept if it matches any one of the provided filters.

- `filters`: `...Filter` — Optional predicates to filter devices. See the [Match helpers](#match-helpers) below.
- Returns `*Monitor` and `error`.

### `gousbmon.NewWithDetector(detector Detector, filters ...Filter) (*Monitor, error)`

Creates a `Monitor` backed by a caller-supplied `Detector`. Can be used to provide a custom detector implementation.

- `detector`: `Detector` — An implementation of `device.Detector`.
- `filters`: `...Filter` — Optional filters.
- Returns `*Monitor` and `error`.

### `(*Monitor) StartMonitoring(onConnect, onDisconnect Callback, opts ...Option) error`

Starts a background goroutine that polls for device changes, invoking `onConnect`/`onDisconnect` as devices appear and disappear. By default it polls every 500ms; use `WithInterval` to override.

- `onConnect`: `Callback` — Invoked when a device is connected.
- `onDisconnect`: `Callback` — Invoked when a device is disconnected.
- `opts`: `...Option` — Optional configuration. See [Options](#options).
- Returns `ErrAlreadyMonitoring` if monitoring is already running.

### `(*Monitor) StopMonitoring()`

Stops the background monitoring goroutine.

### `(*Monitor) GetAvailableDevices() (map[string]device.Info, error)`

Returns the currently connected devices keyed by device ID, after applying any configured filters.

### `(*Monitor) ChangesFromLastCheck(update bool) (removed, added map[string]device.Info, err error)`

Returns the devices removed and added since the previous check. When `update` is `true`, the internal snapshot is saved to the current state.

### `(*Monitor) CheckChanges(onConnect, onDisconnect Callback, update bool) error`

Runs a single change-detection pass, invoking the callbacks for each removed and added device. When `update` is `true`, the internal snapshot is saved to the current state.

### Options

#### `WithInterval(d time.Duration) Option`

Sets the polling interval. Values <= 0 use `DefaultCheckInterval` (500ms).

```go
monitor.StartMonitoring(onConnect, onDisconnect, gousbmon.WithInterval(1*time.Second))
```

### Match helpers

All match functions return a `Filter` predicate:

- `MatchVendorID(id string) Filter`
- `MatchModelID(id string) Filter`
- `MatchVendor(name string) Filter`
- `MatchModel(name string) Filter`
- `MatchSerial(serial string) Filter`
- `MatchUSBInterfaces(interfaces string) Filter`
- `MatchRevision(revision string) Filter`
- `MatchUSBClassFromDatabase(class string) Filter`
- `MatchVendorFromDatabase(vendor string) Filter`
- `MatchModelFromDatabase(model string) Filter`
- `MatchDevName(name string) Filter`
- `MatchDevType(devType string) Filter`

Combine filters with AND logic:

- `MatchAll(filters ...Filter) Filter`

## Device Properties

The `device.Info` struct returned by most functions contains the following fields:

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

