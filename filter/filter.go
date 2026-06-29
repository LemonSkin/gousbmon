package filter

import "github.com/LemonSkin/gousbmon/device"

type Filter func(device.Info) bool

// MatchAll returns a Filter that matches only when all of the provided filters match.
func MatchAll(filters ...Filter) Filter {
	return func(d device.Info) bool {
		for _, f := range filters {
			if !f(d) {
				return false
			}
		}
		return true
	}
}

// MatchVendorID returns a Filter that matches devices with the given vendor ID.
func MatchVendorID(id string) Filter {
	return func(d device.Info) bool { return d.IDVendorID == id }
}

// MatchModelID returns a Filter that matches devices with the given model ID.
func MatchModelID(id string) Filter {
	return func(d device.Info) bool { return d.IDModelID == id }
}

// MatchVendor returns a Filter that matches devices with the given vendor name.
func MatchVendor(name string) Filter {
	return func(d device.Info) bool { return d.IDVendor == name }
}

// MatchModel returns a Filter that matches devices with the given model name.
func MatchModel(name string) Filter {
	return func(d device.Info) bool { return d.IDModel == name }
}

// MatchSerial returns a Filter that matches devices with the given serial number.
func MatchSerial(serial string) Filter {
	return func(d device.Info) bool { return d.IDSerial == serial }
}

// MatchUSBInterfaces returns a Filter that matches devices with the given USB interfaces string.
func MatchUSBInterfaces(interfaces string) Filter {
	return func(d device.Info) bool { return d.IDUSBInterfaces == interfaces }
}

// MatchRevision returns a Filter that matches devices with the given revision.
func MatchRevision(revision string) Filter {
	return func(d device.Info) bool { return d.IDRevision == revision }
}

// MatchUSBClassFromDatabase returns a Filter that matches devices with the given USB class.
func MatchUSBClassFromDatabase(class string) Filter {
	return func(d device.Info) bool { return d.IDUSBClassFromDatabase == class }
}

// MatchVendorFromDatabase returns a Filter that matches devices with the given vendor from database.
func MatchVendorFromDatabase(vendor string) Filter {
	return func(d device.Info) bool { return d.IDVendorFromDatabase == vendor }
}

// MatchModelFromDatabase returns a Filter that matches devices with the given model from database.
func MatchModelFromDatabase(model string) Filter {
	return func(d device.Info) bool { return d.IDModelFromDatabase == model }
}

// MatchDevName returns a Filter that matches devices with the given device name.
func MatchDevName(name string) Filter {
	return func(d device.Info) bool { return d.DevName == name }
}

// MatchDevType returns a Filter that matches devices with the given device type.
func MatchDevType(devType string) Filter {
	return func(d device.Info) bool { return d.DevType == devType }
}
