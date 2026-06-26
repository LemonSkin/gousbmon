// Package device defines the platform-independent representation of a USB device and the detector contract that
// platform backends implement.
package device

// Info holds the normalized attributes of a USB device.
type Info struct {
	IDVendorID             string
	IDVendor               string
	IDModel                string
	IDModelID              string
	IDSerial               string
	IDUSBInterfaces        string
	IDRevision             string
	IDUSBClassFromDatabase string
	IDVendorFromDatabase   string
	IDModelFromDatabase    string
	DevName                string
	DevType                string
}

// Detector is implemented by each platform-specific backend. It returns the raw set of currently connected USB devices.
type Detector interface {
	GetAvailableDevices() (map[string]Info, error)
}
