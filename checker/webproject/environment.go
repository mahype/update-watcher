package webproject

import (
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/mahype/update-watcher/internal/executil"
	"github.com/mahype/update-watcher/internal/fsutil"
)

// Environment represents a web project environment type.
type Environment string

const (
	EnvAuto          Environment = "auto"
	EnvNative        Environment = "native"
	EnvDockerCompose Environment = "docker-compose"
	EnvDdev          Environment = "ddev"
	EnvLando         Environment = "lando"
)

// AllEnvironments returns all known environment types (excluding auto).
var AllEnvironments = []Environment{
	EnvDdev, EnvLando, EnvDockerCompose, EnvNative,
}

// Label returns a human-readable label for the environment.
func (e Environment) Label() string {
	switch e {
	case EnvDdev:
		return "ddev"
	case EnvLando:
		return "Lando"
	case EnvDockerCompose:
		return "Docker Compose"
	case EnvNative:
		return "Native"
	case EnvAuto:
		return "Auto-detect"
	default:
		return string(e)
	}
}

// IsContainerBased returns true if commands run inside a container.
func (e Environment) IsContainerBased() bool {
	switch e {
	case EnvDdev, EnvLando, EnvDockerCompose:
		return true
	default:
		return false
	}
}

// NeedsRunAs returns true if the environment supports the run_as (sudo -u) option.
func (e Environment) NeedsRunAs() bool {
	return e == EnvNative
}

// DetectEnvironment walks up the directory tree from the given path and
// checks for environment markers.
func DetectEnvironment(projectPath string) Environment {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return EnvNative
	}

	dir := absPath
	for {
		if fsutil.FileExists(filepath.Join(dir, ".ddev", "config.yaml")) {
			slog.Debug("detected webproject environment", "env", EnvDdev, "dir", dir)
			return EnvDdev
		}
		if fsutil.FileExists(filepath.Join(dir, ".lando.yml")) {
			slog.Debug("detected webproject environment", "env", EnvLando, "dir", dir)
			return EnvLando
		}
		for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
			if fsutil.FileExists(filepath.Join(dir, name)) {
				slog.Debug("detected webproject environment", "env", EnvDockerCompose, "dir", dir)
				return EnvDockerCompose
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return EnvNative
}

// EnvironmentDescription returns a short description of the environment.
func EnvironmentDescription(env Environment) string {
	switch env {
	case EnvDdev:
		return "ddev project (uses 'ddev exec')"
	case EnvLando:
		return "Lando project (uses 'lando ssh')"
	case EnvDockerCompose:
		return "Docker Compose (uses 'docker compose exec')"
	case EnvNative:
		return "Native (direct execution)"
	default:
		return string(env)
	}
}

// CommandSpec describes how to invoke a package manager command.
type CommandSpec struct {
	Command string
	Args    []string
	WorkDir string
	Env     []string
	RunAs   string
}

// BuildManagerCommand creates a CommandSpec for running a package manager
// command in the project's environment.
func BuildManagerCommand(project ProjectConfig, tool string, args ...string) CommandSpec {
	switch project.Environment {
	case EnvDdev:
		allArgs := append([]string{"exec", tool}, args...)
		return CommandSpec{
			Command: "ddev",
			Args:    allArgs,
			WorkDir: findProjectDir(project.Path, project.Environment),
		}
	case EnvLando:
		cmdStr := tool
		if len(args) > 0 {
			cmdStr += " " + strings.Join(args, " ")
		}
		return CommandSpec{
			Command: "lando",
			Args:    []string{"ssh", "-c", cmdStr},
			WorkDir: findProjectDir(project.Path, project.Environment),
		}
	case EnvDockerCompose:
		allArgs := append([]string{"compose", "exec", "-T", "app", tool}, args...)
		return CommandSpec{
			Command: "docker",
			Args:    allArgs,
			WorkDir: findProjectDir(project.Path, project.Environment),
		}
	default: // EnvNative
		return CommandSpec{
			Command: tool,
			Args:    args,
			WorkDir: project.Path,
			RunAs:   project.RunAs,
		}
	}
}

// ExecuteCommand runs a CommandSpec using executil.
func ExecuteCommand(spec CommandSpec) (*executil.Result, error) {
	if spec.RunAs != "" {
		return executil.RunAsUserWithEnv(spec.Env, spec.RunAs, spec.Command, spec.Args...)
	}
	return executil.RunInDirWithEnv(spec.WorkDir, spec.Env, spec.Command, spec.Args...)
}

// findProjectDir walks up from projectPath to find the directory containing
// the environment marker file.
func findProjectDir(projectPath string, env Environment) string {
	return fsutil.FindParentDir(projectPath, func(dir string) bool {
		switch env {
		case EnvDdev:
			return fsutil.FileExists(filepath.Join(dir, ".ddev", "config.yaml"))
		case EnvLando:
			return fsutil.FileExists(filepath.Join(dir, ".lando.yml"))
		case EnvDockerCompose:
			for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
				if fsutil.FileExists(filepath.Join(dir, name)) {
					return true
				}
			}
		}
		return false
	})
}
