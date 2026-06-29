package errors

import "errors"

// TODO: Should probably rename this package to prevent shadowing
var ErrUnsupportedPlatform = errors.New("gousbmon: platform not supported")
var ErrAlreadyMonitoring = errors.New("gousbmon: monitoring is already running")
var ErrInvalidInterval = errors.New("gousbmon: invalid interval")
