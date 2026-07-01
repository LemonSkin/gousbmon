package gousbmon

import (
	"log/slog"

	"github.com/LemonSkin/gousbmon/device"
)

// config holds the options applied when constructing a Monitor.
type config struct {
	detector     device.Detector
	logger       *slog.Logger
	filterGroups [][]filter
}

// Option configures a Monitor at construction time. See WithDetector, WithLogger, WithHandler, and WithFilters.
type Option func(*config)

// WithDetector sets a custom Detector, bypassing platform detection. Mostly used for testing or for users who want
// to provide their own custom backend.
func WithDetector(d device.Detector) Option {
	return func(c *config) { c.detector = d }
}

// WithLogger sets the logger the Monitor and its platform detector use for diagnostic (debug) logging.
// WithLogger and WithHandler are mutually exclusive; if both are supplied the last one wins.
func WithLogger(l *slog.Logger) Option {
	return func(c *config) { c.logger = l }
}

// WithHandler sets the Monitor's logger from an slog.Handler, letting callers direct gousbmon's logs to a dedicated
// handler separate from their application logger. WithLogger and WithHandler are mutually exclusive; if both are
// supplied the last one wins.
func WithHandler(h slog.Handler) Option {
	return func(c *config) { c.logger = slog.New(h) }
}

// WithFilters restricts the devices that are reported and monitored to those matching the criteria
// defined in the Filter objects. Multiple Filters can be passed for OR logic between them.
func WithFilters(filters ...*Filter) Option {
	return func(c *config) {
		for _, f := range filters {
			c.filterGroups = append(c.filterGroups, f.filters)
		}
	}
}

// newConfig applies user-supplied options.
func newConfig(opts []Option) config {
	cfg := config{logger: slog.New(slog.DiscardHandler)}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
