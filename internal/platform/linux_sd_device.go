//go:build linux && !udev && sd_device

package platform

/*
#cgo pkg-config: libsystemd
#include <systemd/sd-device.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/LemonSkin/gousbmon/device"
)

// udevPropertyKeys are the udev/hwdb property names read for each device. DEVNAME and DEVTYPE are read via dedicated
// getters instead.
var udevPropertyKeys = []string{
	"ID_VENDOR_ID",
	"ID_MODEL_ID",
	"ID_VENDOR",
	"ID_MODEL",
	"ID_SERIAL",
	"ID_SERIAL_SHORT",
	"ID_REVISION",
	"ID_USB_INTERFACES",
	"ID_USB_CLASS_FROM_DATABASE",
	"ID_VENDOR_FROM_DATABASE",
	"ID_MODEL_FROM_DATABASE",
}

// sdDeviceDetector implements device.Detector using libsystemd's sd-device.
type sdDeviceDetector struct{}

// New returns the Linux USB detector.
func New() (device.Detector, error) {
	fmt.Printf("Using sd_device detector\n")
	return &sdDeviceDetector{}, nil
}

func (d *sdDeviceDetector) GetAvailableDevices() (map[string]device.Info, error) {
	var enum *C.sd_device_enumerator
	if rc := C.sd_device_enumerator_new(&enum); rc < 0 {
		return nil, fmt.Errorf("gousbmon: sd_device_enumerator_new failed: %d", int(rc))
	}
	defer C.sd_device_enumerator_unref(enum)

	subsystem := C.CString("usb")
	C.sd_device_enumerator_add_match_subsystem(enum, subsystem, 1)
	C.free(unsafe.Pointer(subsystem))

	// The enumerator returns only udevd-initialised devices
	result := make(map[string]device.Info)
	for dev := C.sd_device_enumerator_get_device_first(enum); dev != nil; dev = C.sd_device_enumerator_get_device_next(enum) {
		if info, key, ok := sdDeviceInfo(dev); ok {
			result[key] = info
		}
	}
	return result, nil
}

// sdDeviceInfo converts an sd-device into a device.Info. It returns false for non usb_device entries and root hubs.
func sdDeviceInfo(dev *C.sd_device) (device.Info, string, bool) {
	devtype, ok := sdDevtype(dev)
	if !ok || devtype != "usb_device" {
		return device.Info{}, "", false
	}
	if sysname, ok := sdSysname(dev); ok && isRootHub(sysname) {
		return device.Info{}, "", false
	}

	props := map[string]string{"DEVTYPE": devtype}
	if devname, ok := sdDevname(dev); ok {
		props["DEVNAME"] = devname
	}
	for _, key := range udevPropertyKeys {
		if val, ok := sdGetProperty(dev, key); ok {
			props[key] = val
		}
	}

	info := infoFromProperties(props)
	mapKey := info.DevName
	if mapKey == "" {
		if syspath, ok := sdSyspath(dev); ok {
			mapKey = syspath
		}
	}
	return info, mapKey, true
}

func sdDevtype(dev *C.sd_device) (string, bool) {
	var val *C.char
	if C.sd_device_get_devtype(dev, &val) < 0 {
		return "", false
	}
	return goString(val), true
}

func sdSysname(dev *C.sd_device) (string, bool) {
	var val *C.char
	if C.sd_device_get_sysname(dev, &val) < 0 {
		return "", false
	}
	return goString(val), true
}

func sdDevname(dev *C.sd_device) (string, bool) {
	var val *C.char
	if C.sd_device_get_devname(dev, &val) < 0 {
		return "", false
	}
	return goString(val), true
}

func sdSyspath(dev *C.sd_device) (string, bool) {
	var val *C.char
	if C.sd_device_get_syspath(dev, &val) < 0 {
		return "", false
	}
	return goString(val), true
}

// sdGetProperty reads a single named property, returning false when the property is absent.
func sdGetProperty(dev *C.sd_device, key string) (string, bool) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var val *C.char
	if C.sd_device_get_property_value(dev, ckey, &val) < 0 {
		return "", false
	}
	return goString(val), true
}

// goString converts a possibly-NULL C string to a Go string.
func goString(s *C.char) string {
	if s == nil {
		return ""
	}
	return C.GoString(s)
}
