//go:build darwin

package platform

import (
	"github.com/LemonSkin/gousbmon/device"
	"github.com/LemonSkin/gousbmon/internal/errors"
)

// New returns the macOS USB detector.
func New() (device.Detector, error) {
	return nil, errors.ErrUnsupportedPlatform
}
