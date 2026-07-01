//go:build darwin

package platform

import (
	"errors"
	"log/slog"

	"github.com/LemonSkin/gousbmon/device"
)

// New returns the macOS USB detector.
func New(_ *slog.Logger) (device.Detector, error) {
	return nil, errors.New("gousbmon: platform not supported")
}
