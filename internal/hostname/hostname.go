package hostname

import "os"

// Get returns the system hostname, or "unknown" if detection fails.
func Get() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}
