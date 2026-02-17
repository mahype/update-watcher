package notifier

import "github.com/mahype/update-watcher/checker"

// Notifier is the interface for all notification backends.
type Notifier interface {
	// Name returns a human-readable identifier.
	Name() string

	// Send delivers the aggregated check results.
	Send(hostname string, results []*checker.CheckResult) error
}

// SendPolicy controls when notifications are dispatched.
type SendPolicy string

const (
	SendAlways        SendPolicy = "always"
	SendOnUpdatesOnly SendPolicy = "only-on-updates"
)
