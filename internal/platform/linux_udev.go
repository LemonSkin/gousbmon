//go:build linux && udev && !sd_device

package platform

/*
#cgo LDFLAGS: -ludev
#include <libudev.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"log/slog"
	"unsafe"

	"github.com/LemonSkin/gousbmon/device"
)

// udevDetector implements device.Detector using libudev.
type udevDetector struct {
	log *slog.Logger
}

// New returns the Linux USB detector.
func New(logger *slog.Logger) (device.Detector, error) {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	logger.Debug("Creating udev detector")
	return &udevDetector{log: logger}, nil
}

func (d *udevDetector) GetAvailableDevices() (map[string]device.Info, error) {
	ctx := C.udev_new()
	if ctx == nil {
		return nil, fmt.Errorf("gousbmon: udev_new failed")
	}
	defer C.udev_unref(ctx)

	enum := C.udev_enumerate_new(ctx)
	if enum == nil {
		return nil, fmt.Errorf("gousbmon: udev_enumerate_new failed")
	}
	defer C.udev_enumerate_unref(enum)

	subsystem := C.CString("usb")
	C.udev_enumerate_add_match_subsystem(enum, subsystem)
	C.free(unsafe.Pointer(subsystem))

	if rc := C.udev_enumerate_scan_devices(enum); rc < 0 {
		return nil, fmt.Errorf("gousbmon: udev_enumerate_scan_devices failed: %d", int(rc))
	}

	result := make(map[string]device.Info)
	for entry := C.udev_enumerate_get_list_entry(enum); entry != nil; entry = C.udev_list_entry_get_next(entry) {
		syspath := C.udev_list_entry_get_name(entry)
		dev := C.udev_device_new_from_syspath(ctx, syspath)
		if dev == nil {
			continue
		}
		if info, key, ok := udevDeviceInfo(dev); ok {
			result[key] = info
		}
		C.udev_device_unref(dev)
	}
	return result, nil
}

// udevDeviceInfo converts a udev device into a device.Info. It returns false for devices that should be skipped:
// non usb_device entries, uninitialised devices, and root hubs.
func udevDeviceInfo(dev *C.struct_udev_device) (device.Info, string, bool) {
	if goString(C.udev_device_get_devtype(dev)) != "usb_device" {
		return device.Info{}, "", false
	}

	if C.udev_device_get_is_initialized(dev) == 0 {
		return device.Info{}, "", false
	}
	if isRootHub(goString(C.udev_device_get_sysname(dev))) {
		return device.Info{}, "", false
	}

	props := udevProperties(dev)
	info := infoFromProperties(props)
	key := info.DevName
	if key == "" {
		key = goString(C.udev_device_get_syspath(dev))
	}
	return info, key, true
}

// udevProperties collects all udev properties of a device into a map.
func udevProperties(dev *C.struct_udev_device) map[string]string {
	props := make(map[string]string)
	for e := C.udev_device_get_properties_list_entry(dev); e != nil; e = C.udev_list_entry_get_next(e) {
		key := goString(C.udev_list_entry_get_name(e))
		if key == "" {
			continue
		}
		props[key] = goString(C.udev_list_entry_get_value(e))
	}
	return props
}

// goString converts a possibly-NULL C string to a string.
func goString(s *C.char) string {
	if s == nil {
		return ""
	}
	return C.GoString(s)
}
