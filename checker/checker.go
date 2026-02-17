package checker

import "time"

// Checker is the interface that all update checkers must implement.
type Checker interface {
	// Name returns a human-readable identifier for this checker.
	Name() string

	// Check performs the update check and returns results.
	Check() (*CheckResult, error)
}

// CheckResult holds the outcome of a single checker run.
type CheckResult struct {
	CheckerName string    `json:"checker_name"`
	Updates     []Update  `json:"updates"`
	Summary     string    `json:"summary"`
	CheckedAt   time.Time `json:"checked_at"`
	Error       string    `json:"error,omitempty"`
}

// HasUpdates returns true if this result contains at least one update.
func (cr *CheckResult) HasUpdates() bool {
	return len(cr.Updates) > 0
}

// HasSecurityUpdates returns true if any update is classified as security.
func (cr *CheckResult) HasSecurityUpdates() bool {
	for _, u := range cr.Updates {
		if u.Type == UpdateTypeSecurity {
			return true
		}
	}
	return false
}

// Update represents a single available update.
type Update struct {
	Name           string `json:"name"`
	CurrentVersion string `json:"current_version"`
	NewVersion     string `json:"new_version"`
	Type           string `json:"type"`
	Priority       string `json:"priority"`
	Source         string `json:"source,omitempty"`
	Phasing        string `json:"phasing,omitempty"`
}

// Update type constants.
const (
	UpdateTypeSecurity = "security"
	UpdateTypeRegular  = "regular"
	UpdateTypePlugin   = "plugin"
	UpdateTypeTheme    = "theme"
	UpdateTypeCore     = "core"
	UpdateTypeImage    = "image"
)

// Priority constants.
const (
	PriorityCritical = "critical"
	PriorityHigh     = "high"
	PriorityNormal   = "normal"
	PriorityLow      = "low"
)

// CheckError wraps errors from checker execution with context.
type CheckError struct {
	CheckerName string
	Err         error
	Retryable   bool
}

func (e *CheckError) Error() string {
	return e.CheckerName + ": " + e.Err.Error()
}

func (e *CheckError) Unwrap() error {
	return e.Err
}
