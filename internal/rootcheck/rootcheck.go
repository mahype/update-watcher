package rootcheck

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"syscall"
)

const serviceUser = "update-watcher"

// IsRoot returns true if the current process runs as UID 0.
func IsRoot() bool {
	return os.Getuid() == 0
}

// IsServiceUser returns true if the current process runs as the dedicated service user.
func IsServiceUser() bool {
	current, err := user.Current()
	if err != nil {
		return false
	}
	return current.Username == serviceUser
}

// ServiceUserName returns the name of the dedicated service user.
func ServiceUserName() string {
	return serviceUser
}

// ServiceUserExists checks if the dedicated "update-watcher" system user exists.
func ServiceUserExists() bool {
	_, err := user.Lookup(serviceUser)
	return err == nil
}

// ReExecAsServiceUser re-executes the current command as the service user
// via "sudo -u update-watcher <binary> <original-args...>".
// This replaces the current process and does not return on success.
func ReExecAsServiceUser() error {
	binary, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		return fmt.Errorf("sudo not found: %w", err)
	}

	// Build args: sudo -u update-watcher /path/to/update-watcher <original-args>
	// Filter out --as-service-user to avoid infinite loop
	var filteredArgs []string
	for _, arg := range os.Args[1:] {
		if arg != "--as-service-user" {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	args := append([]string{"sudo", "-u", serviceUser, binary}, filteredArgs...)

	fmt.Fprintf(os.Stderr, "Re-running as '%s' user...\n", serviceUser)

	// Replace current process
	return syscall.Exec(sudoPath, args, os.Environ())
}

// WarnOrReExec checks if running as root and handles it appropriately.
//
// If not root, returns immediately (caller should continue normally).
// If root and the service user exists:
//   - force=true: re-execs as service user without asking
//   - interactive (TTY): asks user whether to re-exec
//   - non-interactive: prints warning and continues
//
// If root but service user doesn't exist, prints a warning.
//
// Returns true if the caller should continue execution.
// Returns false if the process was replaced (should not happen, but as safety).
func WarnOrReExec(force bool) bool {
	// Already running as the service user → nothing to do.
	if IsServiceUser() {
		return true
	}

	if !IsRoot() {
		// Normal user (not root, not service user): offer re-exec if service user exists.
		if ServiceUserExists() {
			if force {
				if err := ReExecAsServiceUser(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to re-run as '%s': %v\n", serviceUser, err)
					fmt.Fprintf(os.Stderr, "Continuing as current user...\n\n")
				}
				return true
			}
			if isInteractive() {
				fmt.Fprintf(os.Stderr, "Note: A '%s' system user exists for server setup.\n", serviceUser)
				fmt.Fprint(os.Stderr, "Re-run as '"+serviceUser+"' user? [Y/n] ")
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer == "" || answer == "y" || answer == "yes" {
					if err := ReExecAsServiceUser(); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: Failed to re-run as '%s': %v\n", serviceUser, err)
						fmt.Fprintf(os.Stderr, "Continuing as current user...\n\n")
					}
					return true
				}
				fmt.Fprintf(os.Stderr, "Continuing as current user. Config will be saved to home directory.\n\n")
			}
		}
		return true
	}

	if !ServiceUserExists() {
		fmt.Fprintf(os.Stderr, "Warning: Running as root is not recommended. "+
			"Consider creating a dedicated '%s' user.\n"+
			"See: update-watcher documentation for Linux server setup.\n\n", serviceUser)
		return true
	}

	// Service user exists
	if force {
		if err := ReExecAsServiceUser(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to re-run as '%s': %v\n", serviceUser, err)
			fmt.Fprintf(os.Stderr, "Continuing as root...\n\n")
		}
		return true
	}

	// Interactive: ask
	if isInteractive() {
		fmt.Fprintf(os.Stderr, "Warning: Running as root. A dedicated '%s' user exists.\n", serviceUser)
		fmt.Fprint(os.Stderr, "Re-run as '"+serviceUser+"' user? [Y/n] ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer == "" || answer == "y" || answer == "yes" {
			if err := ReExecAsServiceUser(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to re-run as '%s': %v\n", serviceUser, err)
				fmt.Fprintf(os.Stderr, "Continuing as root...\n\n")
			}
			return true
		}

		fmt.Fprintf(os.Stderr, "Continuing as root. Config and cron will be owned by root.\n\n")
		return true
	}

	// Non-interactive: warn
	fmt.Fprintf(os.Stderr, "Warning: Running as root. Consider using: sudo -u %s update-watcher ...\n\n", serviceUser)
	return true
}

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
