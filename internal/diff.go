package internal

import "github.com/LemonSkin/gousbmon/device"

// Diff returns the devices removed and added when moving from prev to current.
func Diff(prev, current map[string]device.Info) (removed, added map[string]device.Info) {
	removed = make(map[string]device.Info)
	added = make(map[string]device.Info)
	for id, info := range prev {
		if _, ok := current[id]; !ok {
			removed[id] = info
		}
	}
	for id, info := range current {
		if _, ok := prev[id]; !ok {
			added[id] = info
		}
	}
	return removed, added
}
