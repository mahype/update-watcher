package version

import (
	"os/exec"
	"strings"
)

// These variables are set at build time via ldflags.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func init() {
	if Version != "dev" {
		return
	}
	out, err := exec.Command("git", "describe", "--tags", "--always").Output()
	if err != nil {
		return
	}
	v := strings.TrimSpace(string(out))
	if v != "" {
		Version = strings.TrimPrefix(v, "v")
	}
}
