//go:build linux && !udev && !sd_device

package platform

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// usbIDsPaths lists the well-known locations of the usb.ids database, in search order.
var usbIDsPaths = []string{
	"/usr/share/hwdata/usb.ids",
	"/usr/share/usb.ids",
	"/usr/share/misc/usb.ids",
	"/var/lib/usbutils/usb.ids",
}

// usbIDs holds the vendor, product and class name lookups parsed from a usb.ids database.
type usbIDs struct {
	vendors  map[string]string // key: vid (lowercase hex)
	products map[string]string // key: vid+pid (lowercase hex)
	classes  map[string]string // key: class code (lowercase hex)
}

// loadUSBIDs loads the first usb.ids database found in usbIDsPaths. A missing database is not an error: an empty (but
// usable) resolver is returned so that the *FromDatabase fields are simply left blank.
func loadUSBIDs() *usbIDs {
	for _, path := range usbIDsPaths {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		ids := parseUSBIDs(f)
		f.Close()
		return ids
	}
	return &usbIDs{}
}

// parseUSBIDs parses a usb.ids database. Only the vendor/product and class (C) sections are retained, since those are
// the ones that map onto device.Info's *FromDatabase fields.
func parseUSBIDs(r io.Reader) *usbIDs {
	ids := &usbIDs{
		vendors:  map[string]string{},
		products: map[string]string{},
		classes:  map[string]string{},
	}

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)

	var curVendor string
	inClass := false
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Class section header, e.g. "C 09  Hub"
		if strings.HasPrefix(line, "C ") {
			inClass = true
			curVendor = ""
			if id, name, ok := splitIDName(line[2:]); ok && len(id) == 2 {
				ids.classes[id] = name
			}
			continue
		}

		// Top-level (non-indented) line: either a vendor or the start of an unrelated section
		if !strings.HasPrefix(line, "\t") {
			inClass = false
			if id, name, ok := splitIDName(line); ok && len(id) == 4 {
				ids.vendors[id] = name
				curVendor = id
			} else {
				curVendor = ""
			}
			continue
		}

		// Deeper indentation (subclass/protocol or interface detail) is not needed
		if strings.HasPrefix(line, "\t\t") {
			continue
		}

		// Single-tab line: a product under the current vendor (class subclasses are ignored)
		if inClass || curVendor == "" {
			continue
		}
		if id, name, ok := splitIDName(line[1:]); ok && len(id) == 4 {
			ids.products[curVendor+id] = name
		}
	}
	return ids
}

// splitIDName splits a "id  Name" entry (id and name separated by two spaces) into its lowercase hex id and name. It
// returns false when the line is not in that form or the id is not hexadecimal.
func splitIDName(s string) (id, name string, ok bool) {
	parts := strings.SplitN(s, "  ", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	id = strings.ToLower(strings.TrimSpace(parts[0]))
	name = strings.TrimSpace(parts[1])
	if id == "" || name == "" || !isHex(id) {
		return "", "", false
	}
	return id, name, true
}

func isHex(s string) bool {
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		default:
			return false
		}
	}
	return true
}

func (u *usbIDs) vendor(vid string) string {
	if u == nil {
		return ""
	}
	return u.vendors[strings.ToLower(vid)]
}

func (u *usbIDs) product(vid, pid string) string {
	if u == nil {
		return ""
	}
	return u.products[strings.ToLower(vid)+strings.ToLower(pid)]
}

func (u *usbIDs) class(code string) string {
	if u == nil {
		return ""
	}
	return u.classes[strings.ToLower(code)]
}
