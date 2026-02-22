package cron

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const binaryName = "update-watcher"

// JobType identifies a cron job kind.
type JobType string

const (
	JobCheck      JobType = "check"
	JobSelfUpdate JobType = "self-update"
)

// allJobTypes lists all known job types for iteration.
var allJobTypes = []JobType{JobCheck, JobSelfUpdate}

// InstalledJob represents a cron job found in the crontab.
type InstalledJob struct {
	Type     JobType
	Schedule string
}

// commentMarker returns the crontab comment for a given job type.
func commentMarker(jt JobType) string {
	switch jt {
	case JobSelfUpdate:
		return "# update-watcher self-update"
	default:
		return "# update-watcher scheduled check"
	}
}

// cronCommand returns the command portion of the crontab line.
func cronCommand(jt JobType, binaryPath string) string {
	switch jt {
	case JobSelfUpdate:
		return fmt.Sprintf("%s self-update 2>&1 | logger -t update-watcher", binaryPath)
	default:
		return fmt.Sprintf("%s run --quiet 2>&1 | logger -t update-watcher", binaryPath)
	}
}

// JobTypeLabel returns a human-readable label for the job type.
func JobTypeLabel(jt JobType) string {
	switch jt {
	case JobSelfUpdate:
		return "Self-Update"
	default:
		return "Update Check"
	}
}

// FormatSchedule converts a cron expression to a human-readable string.
func FormatSchedule(expr string) string {
	parts := strings.Fields(expr)
	if len(parts) < 5 {
		return expr
	}

	// Every N minutes: "*/N * * * *"
	if strings.HasPrefix(parts[0], "*/") && parts[1] == "*" && parts[2] == "*" && parts[3] == "*" && parts[4] == "*" {
		return fmt.Sprintf("every %s minutes", strings.TrimPrefix(parts[0], "*/"))
	}

	if parts[2] == "*" && parts[3] == "*" && parts[4] == "*" {
		// Every N hours: "0 */N * * *"
		if parts[0] == "0" && strings.HasPrefix(parts[1], "*/") {
			return fmt.Sprintf("every %s hours", strings.TrimPrefix(parts[1], "*/"))
		}
		// Daily at specific time: "M H * * *"
		minute, mErr := strconv.Atoi(parts[0])
		hour, hErr := strconv.Atoi(parts[1])
		if mErr == nil && hErr == nil {
			return fmt.Sprintf("daily at %02d:%02d", hour, minute)
		}
	}

	// Weekly: "M H * * D"
	if parts[2] == "*" && parts[3] == "*" && parts[4] != "*" {
		minute, mErr := strconv.Atoi(parts[0])
		hour, hErr := strconv.Atoi(parts[1])
		day, dErr := strconv.Atoi(parts[4])
		dayNames := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
		if mErr == nil && hErr == nil && dErr == nil && day >= 0 && day <= 6 {
			return fmt.Sprintf("weekly on %s at %02d:%02d", dayNames[day], hour, minute)
		}
	}

	return expr
}

// IntervalToExpr converts a human-friendly interval to a cron expression.
func IntervalToExpr(value int, unit string) (string, error) {
	switch unit {
	case "hours":
		if value < 1 || value > 23 {
			return "", fmt.Errorf("hours must be between 1 and 23")
		}
		return fmt.Sprintf("0 */%d * * *", value), nil
	case "minutes":
		if value < 1 || value > 59 {
			return "", fmt.Errorf("minutes must be between 1 and 59")
		}
		return fmt.Sprintf("*/%d * * * *", value), nil
	default:
		return "", fmt.Errorf("unknown unit %q (use 'hours' or 'minutes')", unit)
	}
}

// ParseTime validates and parses an HH:MM time string.
func ParseTime(timeStr string) (hour, minute int, err error) {
	parts := strings.SplitN(timeStr, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected HH:MM format")
	}
	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}
	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}
	return hour, minute, nil
}

// --- Multi-job API ---

// InstallJob installs a cron entry for the given job type with an HH:MM time.
func InstallJob(jt JobType, timeStr string) error {
	hour, minute, err := ParseTime(timeStr)
	if err != nil {
		return fmt.Errorf("invalid time format %q: %w", timeStr, err)
	}
	cronExpr := fmt.Sprintf("%d %d * * *", minute, hour)
	return InstallJobWithExpr(jt, cronExpr)
}

// InstallJobWithExpr installs a cron entry for the given job type with a full cron expression.
func InstallJobWithExpr(jt JobType, cronExpr string) error {
	binaryPath, err := findBinary()
	if err != nil {
		return err
	}

	cronLine := fmt.Sprintf("%s %s", cronExpr, cronCommand(jt, binaryPath))
	return installCronLineForJob(jt, cronLine)
}

// UninstallJob removes the cron entry for the given job type.
func UninstallJob(jt JobType) error {
	existing, err := readCrontab()
	if err != nil {
		return err
	}
	cleaned := removeByMarker(existing, commentMarker(jt))
	return writeCrontab(cleaned)
}

// UninstallAll removes all update-watcher cron entries.
func UninstallAll() error {
	existing, err := readCrontab()
	if err != nil {
		return err
	}
	cleaned := existing
	for _, jt := range allJobTypes {
		cleaned = removeByMarker(cleaned, commentMarker(jt))
	}
	return writeCrontab(cleaned)
}

// IsJobInstalled returns whether a cron entry exists for the given type, and its schedule.
func IsJobInstalled(jt JobType) (bool, string) {
	existing, err := readCrontab()
	if err != nil {
		return false, ""
	}
	marker := commentMarker(jt)
	lines := strings.Split(existing, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == marker && i+1 < len(lines) {
			return true, extractSchedule(lines[i+1])
		}
	}
	return false, ""
}

// InstalledJobs returns all installed update-watcher cron jobs.
func InstalledJobs() []InstalledJob {
	var jobs []InstalledJob
	for _, jt := range allJobTypes {
		if installed, schedule := IsJobInstalled(jt); installed {
			jobs = append(jobs, InstalledJob{Type: jt, Schedule: schedule})
		}
	}
	return jobs
}

// --- Backward-compatible wrappers ---

// Install is a backward-compatible wrapper; installs a check job.
func Install(timeStr string) error {
	return InstallJob(JobCheck, timeStr)
}

// InstallWithExpr is a backward-compatible wrapper; installs a check job.
func InstallWithExpr(cronExpr string) error {
	return InstallJobWithExpr(JobCheck, cronExpr)
}

// Uninstall is a backward-compatible wrapper; removes the check job.
func Uninstall() error {
	return UninstallJob(JobCheck)
}

// IsInstalled is a backward-compatible wrapper; checks the check job.
func IsInstalled() (bool, string) {
	return IsJobInstalled(JobCheck)
}

// --- Internal helpers ---

func installCronLineForJob(jt JobType, cronLine string) error {
	existing, err := readCrontab()
	if err != nil {
		return err
	}

	marker := commentMarker(jt)
	cleaned := removeByMarker(existing, marker)
	newCrontab := cleaned
	if !strings.HasSuffix(newCrontab, "\n") && newCrontab != "" {
		newCrontab += "\n"
	}
	newCrontab += marker + "\n" + cronLine + "\n"

	return writeCrontab(newCrontab)
}

func removeByMarker(crontab string, marker string) string {
	lines := strings.Split(crontab, "\n")
	var result []string
	skip := false
	for _, line := range lines {
		if strings.TrimSpace(line) == marker {
			skip = true
			continue
		}
		if skip {
			skip = false
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

func extractSchedule(cronLine string) string {
	parts := strings.Fields(cronLine)
	if len(parts) >= 5 {
		return strings.Join(parts[:5], " ")
	}
	return cronLine
}

func readCrontab() (string, error) {
	out, err := exec.Command("crontab", "-l").CombinedOutput()
	if err != nil {
		outStr := string(out)
		if strings.Contains(outStr, "no crontab") {
			return "", nil
		}
		return "", fmt.Errorf("failed to read crontab: %w (output: %s)", err, outStr)
	}
	return string(out), nil
}

func writeCrontab(content string) error {
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to write crontab: %w (output: %s)", err, string(out))
	}
	return nil
}

func findBinary() (string, error) {
	path, err := exec.LookPath(binaryName)
	if err != nil {
		// Fallback to common install location
		return "/usr/local/bin/" + binaryName, nil
	}
	return path, nil
}
