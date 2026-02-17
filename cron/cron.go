package cron

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const (
	cronComment = "# update-watcher scheduled check"
	binaryName  = "update-watcher"
)

// Install adds or replaces the update-watcher cron entry.
func Install(timeStr string) error {
	hour, minute, err := parseTime(timeStr)
	if err != nil {
		return fmt.Errorf("invalid time format %q: %w", timeStr, err)
	}

	binaryPath, err := findBinary()
	if err != nil {
		return err
	}

	cronLine := fmt.Sprintf("%d %d * * * %s run --quiet 2>&1 | logger -t update-watcher",
		minute, hour, binaryPath)

	return installCronLine(cronLine)
}

// InstallWithExpr adds or replaces the update-watcher cron entry with a custom expression.
func InstallWithExpr(cronExpr string) error {
	binaryPath, err := findBinary()
	if err != nil {
		return err
	}

	cronLine := fmt.Sprintf("%s %s run --quiet 2>&1 | logger -t update-watcher",
		cronExpr, binaryPath)

	return installCronLine(cronLine)
}

func installCronLine(cronLine string) error {
	existing, err := readCrontab()
	if err != nil {
		return err
	}

	cleaned := removeExisting(existing)
	newCrontab := cleaned
	if !strings.HasSuffix(newCrontab, "\n") && newCrontab != "" {
		newCrontab += "\n"
	}
	newCrontab += cronComment + "\n" + cronLine + "\n"

	return writeCrontab(newCrontab)
}

// Uninstall removes the update-watcher cron entry.
func Uninstall() error {
	existing, err := readCrontab()
	if err != nil {
		return err
	}

	cleaned := removeExisting(existing)
	return writeCrontab(cleaned)
}

// IsInstalled returns whether a cron entry exists, and the schedule if found.
func IsInstalled() (bool, string) {
	existing, err := readCrontab()
	if err != nil {
		return false, ""
	}

	lines := strings.Split(existing, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == cronComment && i+1 < len(lines) {
			return true, extractSchedule(lines[i+1])
		}
	}
	return false, ""
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

func removeExisting(crontab string) string {
	lines := strings.Split(crontab, "\n")
	var result []string
	skip := false
	for _, line := range lines {
		if strings.TrimSpace(line) == cronComment {
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

func parseTime(timeStr string) (hour, minute int, err error) {
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

func findBinary() (string, error) {
	path, err := exec.LookPath(binaryName)
	if err != nil {
		// Fallback to common install location
		return "/usr/local/bin/" + binaryName, nil
	}
	return path, nil
}
