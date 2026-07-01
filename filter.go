package gousbmon

// Filter builds a set of device matching criteria using a builder pattern.
type Filter struct {
	filters []filter
}

// filter is a function that determines if a device matches a specific criterion.
type filter func(DeviceInfo) bool

// NewFilter creates a new empty Filter.
func NewFilter() *Filter {
	return &Filter{filters: make([]filter, 0)}
}

// MatchVendorID adds a criterion to match devices with the given vendor ID.
func (f *Filter) MatchVendorID(id string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDVendorID == id }
	f.filters = append(f.filters, filter)
	return f
}

// MatchModelID adds a criterion to match devices with the given model ID.
func (f *Filter) MatchModelID(id string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDModelID == id }
	f.filters = append(f.filters, filter)
	return f
}

// MatchVendor adds a criterion to match devices with the given vendor name.
func (f *Filter) MatchVendor(name string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDVendor == name }
	f.filters = append(f.filters, filter)
	return f
}

// MatchModel adds a criterion to match devices with the given model name.
func (f *Filter) MatchModel(name string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDModel == name }
	f.filters = append(f.filters, filter)
	return f
}

// MatchSerial adds a criterion to match devices with the given serial number.
func (f *Filter) MatchSerial(serial string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDSerial == serial }
	f.filters = append(f.filters, filter)
	return f
}

// MatchUSBInterfaces adds a criterion to match devices with the given USB interfaces string.
func (f *Filter) MatchUSBInterfaces(interfaces string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDUSBInterfaces == interfaces }
	f.filters = append(f.filters, filter)
	return f
}

// MatchRevision adds a criterion to match devices with the given revision.
func (f *Filter) MatchRevision(revision string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDRevision == revision }
	f.filters = append(f.filters, filter)
	return f
}

// MatchUSBClassFromDatabase adds a criterion to match devices with the given USB class.
func (f *Filter) MatchUSBClassFromDatabase(class string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDUSBClassFromDatabase == class }
	f.filters = append(f.filters, filter)
	return f
}

// MatchVendorFromDatabase adds a criterion to match devices with the given vendor from database.
func (f *Filter) MatchVendorFromDatabase(vendor string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDVendorFromDatabase == vendor }
	f.filters = append(f.filters, filter)
	return f
}

// MatchModelFromDatabase adds a criterion to match devices with the given model from database.
func (f *Filter) MatchModelFromDatabase(model string) *Filter {
	filter := func(d DeviceInfo) bool { return d.IDModelFromDatabase == model }
	f.filters = append(f.filters, filter)
	return f
}

// MatchDevName adds a criterion to match devices with the given device name.
func (f *Filter) MatchDevName(name string) *Filter {
	filter := func(d DeviceInfo) bool { return d.DevName == name }
	f.filters = append(f.filters, filter)
	return f
}

// MatchDevType adds a criterion to match devices with the given device type.
func (f *Filter) MatchDevType(devType string) *Filter {
	filter := func(d DeviceInfo) bool { return d.DevType == devType }
	f.filters = append(f.filters, filter)
	return f
}
