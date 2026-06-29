//go:build linux && !udev && !sd_device

package platform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const usbIDsFixture = `# comment line
046d  Logitech, Inc.
	c077  M105 Optical Mouse
	c52b  Unifying Receiver
1d6b  Linux Foundation
	0002  2.0 root hub
C 03  Human Interface Device
	01  Boot Interface Subclass
		01  Keyboard
C 09  Hub
`

func newFixtureIDs(t *testing.T) *usbIDs {
	t.Helper()
	return parseUSBIDs(strings.NewReader(usbIDsFixture))
}

func TestSplitIDName(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		wantID   string
		wantName string
		wantOK   bool
	}{
		{"vendor", "046d  Logitech, Inc.", "046d", "Logitech, Inc.", true},
		{"uppercase normalised", "046D  Logitech", "046d", "Logitech", true},
		{"surrounding whitespace trimmed", "046d  Logitech  ", "046d", "Logitech", true},
		{"single space separator rejected", "046d Logitech", "", "", false},
		{"non-hex id", "zzzz  Nope", "", "", false},
		{"empty id", "  Name only", "", "", false},
		{"empty name", "046d  ", "", "", false},
		{"empty", "", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, name, ok := splitIDName(tt.in)
			if id != tt.wantID || name != tt.wantName || ok != tt.wantOK {
				t.Errorf("splitIDName(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.in, id, name, ok, tt.wantID, tt.wantName, tt.wantOK)
			}
		})
	}
}

func TestIsHex(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"046d", true},
		{"0002", true},
		{"abcdef0123456789", true},
		{"", true}, // vacuously true: no non-hex runes
		{"046D", false},
		{"04g6", false},
		{"04 6", false},
		{"0x46", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := isHex(tt.in); got != tt.want {
				t.Errorf("isHex(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseUSBIDs(t *testing.T) {
	ids := newFixtureIDs(t)

	if got := ids.vendor("046d"); got != "Logitech, Inc." {
		t.Errorf("vendor(046d) = %q", got)
	}
	if got := ids.product("046d", "c077"); got != "M105 Optical Mouse" {
		t.Errorf("product(046d, c077) = %q", got)
	}
	if got := ids.product("046d", "c52b"); got != "Unifying Receiver" {
		t.Errorf("product(046d, c52b) = %q", got)
	}
	if got := ids.vendor("1d6b"); got != "Linux Foundation" {
		t.Errorf("vendor(1d6b) = %q", got)
	}
	if got := ids.product("1d6b", "0002"); got != "2.0 root hub" {
		t.Errorf("product(1d6b, 0002) = %q", got)
	}
	if got := ids.class("03"); got != "Human Interface Device" {
		t.Errorf("class(03) = %q", got)
	}
	if got := ids.class("09"); got != "Hub" {
		t.Errorf("class(09) = %q", got)
	}
}

func TestParseUSBIDs_ClassEntriesDoNotLeakIntoProducts(t *testing.T) {
	ids := newFixtureIDs(t)

	// The "01 Boot Interface Subclass" line lives under "C 03"; it must not be attributed to the preceding vendor
	// (1d6b) as a product, nor recorded as a vendor.
	if got := ids.product("1d6b", "01"); got != "" {
		t.Errorf("product(1d6b, 01) = %q, want empty", got)
	}
	if got := ids.vendor("01"); got != "" {
		t.Errorf("vendor(01) = %q, want empty", got)
	}
	// Deeply indented protocol lines ("\t\t01 Keyboard") must be ignored entirely.
	if got := ids.product("03", "01"); got != "" {
		t.Errorf("product(03, 01) = %q, want empty", got)
	}
}

func TestParseUSBIDs_IgnoresUnrelatedSections(t *testing.T) {
	// usb.ids contains other top-level sections (HID, AT, L, ...) whose headers are not bare 4-hex ids. They must be
	// ignored and must not capture the indented lines that follow as products.
	const db = `046d  Logitech, Inc.
	c077  M105 Optical Mouse
HID 00  None
	01  Pointer
L 0409  English (US)
`
	ids := parseUSBIDs(strings.NewReader(db))

	if got := ids.product("046d", "c077"); got != "M105 Optical Mouse" {
		t.Errorf("product(046d, c077) = %q", got)
	}
	if got := ids.vendor("hid"); got != "" {
		t.Errorf("vendor(hid) = %q, want empty", got)
	}
	// The "\t01 Pointer" line follows a non-vendor section header, so it must not be recorded under 046d.
	if got := ids.product("046d", "01"); got != "" {
		t.Errorf("product(046d, 01) = %q, want empty", got)
	}
}

func TestParseUSBIDs_Empty(t *testing.T) {
	ids := parseUSBIDs(strings.NewReader(""))
	if got := ids.vendor("046d"); got != "" {
		t.Errorf("vendor on empty db = %q, want empty", got)
	}
	if got := ids.product("046d", "c077"); got != "" {
		t.Errorf("product on empty db = %q, want empty", got)
	}
	if got := ids.class("03"); got != "" {
		t.Errorf("class on empty db = %q, want empty", got)
	}
}

func TestUSBIDsLookupsCaseInsensitive(t *testing.T) {
	ids := newFixtureIDs(t)

	if got := ids.vendor("046D"); got != "Logitech, Inc." {
		t.Errorf("vendor(046D) = %q", got)
	}
	if got := ids.product("046D", "C077"); got != "M105 Optical Mouse" {
		t.Errorf("product(046D, C077) = %q", got)
	}
	if got := ids.class("03"); got != "Human Interface Device" {
		t.Errorf("class(03) = %q", got)
	}
}

func TestNilUSBIDsLookups(t *testing.T) {
	var ids *usbIDs
	if ids.vendor("046d") != "" || ids.product("046d", "c077") != "" || ids.class("03") != "" {
		t.Error("nil *usbIDs lookups should return empty strings")
	}
}

func TestLoadUSBIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "usb.ids")
	if err := os.WriteFile(path, []byte(usbIDsFixture), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	// loadUSBIDs walks usbIDsPaths in order; point it at our fixture (preceded by a missing path to exercise the
	// "skip on error" branch) and restore afterwards.
	orig := usbIDsPaths
	t.Cleanup(func() { usbIDsPaths = orig })
	usbIDsPaths = []string{filepath.Join(dir, "missing.ids"), path}

	ids := loadUSBIDs()
	if ids == nil {
		t.Fatal("loadUSBIDs returned nil")
	}
	if got := ids.vendor("046d"); got != "Logitech, Inc." {
		t.Errorf("vendor(046d) = %q, want %q", got, "Logitech, Inc.")
	}
}

func TestLoadUSBIDs_NotFound(t *testing.T) {
	orig := usbIDsPaths
	t.Cleanup(func() { usbIDsPaths = orig })
	usbIDsPaths = []string{filepath.Join(t.TempDir(), "does-not-exist.ids")}

	ids := loadUSBIDs()
	if ids == nil {
		t.Fatal("loadUSBIDs should return a non-nil resolver even when no database is found")
	}
	if got := ids.vendor("046d"); got != "" {
		t.Errorf("vendor lookup on empty resolver = %q, want empty", got)
	}
}
