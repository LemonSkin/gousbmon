//go:build linux

package platform

import (
	"github.com/LemonSkin/gousbmon/device"
	"github.com/LemonSkin/gousbmon/internal/errors"
)

// New returns the Linux USB detector.
func New() (device.Detector, error) {
	return nil, errors.ErrUnsupportedPlatform
}
