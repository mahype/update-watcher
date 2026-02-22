package checker

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Checker is the interface that all update checkers must implement.
type Checker interface {
	// Name returns a human-readable identifier for this checker.
	Name() string

	// Check performs the update check and returns results.
	Check(ctx context.Context) (*CheckResult, error)
}

// CheckResult holds the outcome of a single checker run.
type CheckResult struct {
	CheckerName string    `json:"checker_name"`
	Updates     []Update  `json:"updates"`
	Summary     string    `json:"summary"`
	CheckedAt   time.Time `json:"checked_at"`
	Error       string    `json:"error,omitempty"`
	Notes       []string  `json:"notes,omitempty"`
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
	UpdateTypeDistro   = "distro"
)

// Priority constants.
const (
	PriorityCritical = "critical"
	PriorityHigh     = "high"
	PriorityNormal   = "normal"
	PriorityLow      = "low"
)

// priorityLevels maps priority strings to numeric levels for comparison.
var priorityLevels = map[string]int{
	PriorityLow:      1,
	PriorityNormal:   2,
	PriorityHigh:     3,
	PriorityCritical: 4,
}

// PriorityLevel returns the numeric level for a priority string.
// Unknown or empty priorities are treated as normal.
func PriorityLevel(priority string) int {
	if lvl, ok := priorityLevels[priority]; ok {
		return lvl
	}
	return priorityLevels[PriorityNormal]
}

// ValidPriority reports whether the given string is a known priority value.
func ValidPriority(p string) bool {
	_, ok := priorityLevels[p]
	return ok
}

// FilterByMinPriority returns only those updates whose priority is at or above
// the given minimum. If minPriority is empty, all updates are returned.
func FilterByMinPriority(updates []Update, minPriority string) []Update {
	if minPriority == "" {
		return updates
	}
	minLevel := PriorityLevel(minPriority)
	filtered := make([]Update, 0, len(updates))
	for _, u := range updates {
		if PriorityLevel(u.Priority) >= minLevel {
			filtered = append(filtered, u)
		}
	}
	return filtered
}

// FilterResultsByPriority creates filtered copies of the given check results,
// keeping only updates at or above the minimum priority. Original results are
// never mutated. Returns the filtered results and the total number of remaining
// updates across all results.
func FilterResultsByPriority(results []*CheckResult, minPriority string) ([]*CheckResult, int) {
	if minPriority == "" {
		total := 0
		for _, cr := range results {
			total += len(cr.Updates)
		}
		return results, total
	}

	filtered := make([]*CheckResult, 0, len(results))
	total := 0
	for _, cr := range results {
		kept := FilterByMinPriority(cr.Updates, minPriority)
		total += len(kept)
		cp := *cr
		cp.Updates = kept
		filtered = append(filtered, &cp)
	}
	return filtered, total
}

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

// BuildSummary generates a standard summary string for a list of updates.
// The unit parameter describes what is being updated (e.g. "packages", "snaps",
// "containers"). When no updates are present the summary reads "all <unit> are
// up to date"; otherwise it is "<count> <unit>" with an optional security count.
func BuildSummary(updates []Update, unit string) string {
	total := len(updates)
	if total == 0 {
		return fmt.Sprintf("all %s are up to date", unit)
	}
	sec := 0
	phased := 0
	for _, u := range updates {
		if u.Type == UpdateTypeSecurity {
			sec++
		}
		if u.Phasing != "" {
			phased++
		}
	}

	var details []string
	if phased > 0 {
		details = append(details, fmt.Sprintf("%d phased", phased))
	}
	if sec > 0 {
		details = append(details, fmt.Sprintf("%d security", sec))
	}
	if len(details) > 0 {
		return fmt.Sprintf("%d %s (%s)", total, unit, strings.Join(details, ", "))
	}
	return fmt.Sprintf("%d %s", total, unit)
}
