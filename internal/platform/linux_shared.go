//go:build linux

package platform

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/LemonSkin/gousbmon/device"
)

// rootHubNameRe matches USB root-hub kernel sysfs names such as "usb1" or "usb2".
var rootHubNameRe = regexp.MustCompile(`^usb[0-9]+$`)

// isRootHub reports whether the given kernel sysfs name denotes a USB root hub.
func isRootHub(sysname string) bool {
	return rootHubNameRe.MatchString(sysname)
}

// buildDevName returns the udev-style DEVNAME (/dev/bus/usb/BBB/DDD) for the given bus and device numbers, or an empty
// string when either number is missing.
func buildDevName(busnum, devnum int) string {
	if busnum <= 0 || devnum <= 0 {
		return ""
	}
	return fmt.Sprintf("/dev/bus/usb/%03d/%03d", busnum, devnum)
}

// infoFromProperties maps a udev / sd-device property set (ID_*, DEVNAME, DEVTYPE) to a normalised device.Info.
func infoFromProperties(props map[string]string) device.DeviceInfo {
	serial := props["ID_SERIAL_SHORT"]
	if serial == "" {
		serial = props["ID_SERIAL"]
	}
	return device.DeviceInfo{
		IDVendorID:             strings.ToLower(props["ID_VENDOR_ID"]),
		IDVendor:               props["ID_VENDOR"],
		IDModel:                props["ID_MODEL"],
		IDModelID:              strings.ToLower(props["ID_MODEL_ID"]),
		IDSerial:               serial,
		IDUSBInterfaces:        props["ID_USB_INTERFACES"],
		IDRevision:             props["ID_REVISION"],
		IDUSBClassFromDatabase: props["ID_USB_CLASS_FROM_DATABASE"],
		IDVendorFromDatabase:   props["ID_VENDOR_FROM_DATABASE"],
		IDModelFromDatabase:    props["ID_MODEL_FROM_DATABASE"],
		DevName:                props["DEVNAME"],
		DevType:                props["DEVTYPE"],
	}
}
