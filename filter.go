package gousbmon

import "github.com/LemonSkin/gousbmon/device"

// Filter is a predicate that decides whether a device matches. A device matches when the function returns true.
type Filter func(device.Info) bool

// applyFilters keeps only the devices that match at least one of the filters. If no filters are provided, devices is
// returned unchanged.
func applyFilters(devices map[string]device.Info, filters []Filter) map[string]device.Info {
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
