package wordpress

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Environment represents a WordPress development environment type.
type Environment string

const (
	EnvAuto          Environment = "auto"
	EnvNative        Environment = "native"
	EnvDdev          Environment = "ddev"
	EnvLando         Environment = "lando"
	EnvWpEnv         Environment = "wp-env"
	EnvDockerCompose Environment = "docker-compose"
	EnvBedrock       Environment = "bedrock"
	EnvLocalWP       Environment = "local"
	EnvMAMP          Environment = "mamp"
	EnvXAMPP         Environment = "xampp"
	EnvLaragon       Environment = "laragon"
	EnvValet         Environment = "valet"
)

// AllEnvironments returns all known environment types (excluding auto).
var AllEnvironments = []Environment{
	EnvDdev, EnvLando, EnvWpEnv, EnvDockerCompose, EnvBedrock,
	EnvLocalWP, EnvMAMP, EnvXAMPP, EnvLaragon, EnvValet, EnvNative,
}

// Label returns a human-readable label for the environment.
func (e Environment) Label() string {
	switch e {
	case EnvDdev:
		return "ddev"
	case EnvLando:
		return "Lando"
	case EnvWpEnv:
		return "wp-env"
	case EnvDockerCompose:
		return "Docker Compose"
	case EnvBedrock:
		return "Bedrock"
	case EnvLocalWP:
		return "Local (LocalWP)"
	case EnvMAMP:
		return "MAMP"
	case EnvXAMPP:
		return "XAMPP"
	case EnvLaragon:
		return "Laragon"
	case EnvValet:
		return "Laravel Valet"
	case EnvNative:
		return "Native (wp-cli)"
	case EnvAuto:
		return "Auto-detect"
	default:
		return string(e)
	}
}

// IsContainerBased returns true if WP-CLI runs inside a container (no --path needed).
func (e Environment) IsContainerBased() bool {
	switch e {
	case EnvDdev, EnvLando, EnvWpEnv, EnvDockerCompose:
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
// checks for environment markers. Returns the detected environment.
func DetectEnvironment(sitePath string) Environment {
	absPath, err := filepath.Abs(sitePath)
	if err != nil {
		return EnvNative
	}

	// Walk up the directory tree
	dir := absPath
	for {
		env := checkDirectoryMarkers(dir)
		if env != "" {
			slog.Debug("detected WordPress environment", "env", env, "dir", dir)
			return env
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}

	// Check path-based heuristics that don't need walking
	env := checkPathHeuristics(absPath)
	if env != "" {
		return env
	}

	// Check if wp-config.php exists at the given path
	if fileExists(filepath.Join(absPath, "wp-config.php")) {
		return EnvNative
	}

	return EnvNative
}

// checkDirectoryMarkers checks a single directory for environment markers.
func checkDirectoryMarkers(dir string) Environment {
	// 1. ddev: .ddev/config.yaml
	ddevConfig := filepath.Join(dir, ".ddev", "config.yaml")
	if fileExists(ddevConfig) {
		return EnvDdev
	}

	// 2. Lando: .lando.yml
	if fileExists(filepath.Join(dir, ".lando.yml")) {
		return EnvLando
	}

	// 3. wp-env: .wp-env.json
	if fileExists(filepath.Join(dir, ".wp-env.json")) {
		return EnvWpEnv
	}

	// 4. Bedrock: composer.json with roots/bedrock
	composerFile := filepath.Join(dir, "composer.json")
	if fileExists(composerFile) && isBedrockProject(composerFile) {
		return EnvBedrock
	}

	// 5. Docker Compose: docker-compose.yml with wordpress image
	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		composeFile := filepath.Join(dir, name)
		if fileExists(composeFile) && hasWordPressService(composeFile) {
			return EnvDockerCompose
		}
	}

	return ""
}

// checkPathHeuristics checks path-based patterns that don't require file content parsing.
func checkPathHeuristics(absPath string) Environment {
	// LocalWP: path contains "/Local Sites/"
	if strings.Contains(absPath, string(filepath.Separator)+"Local Sites"+string(filepath.Separator)) {
		return EnvLocalWP
	}

	switch runtime.GOOS {
	case "darwin":
		// MAMP
		if strings.HasPrefix(absPath, "/Applications/MAMP/") {
			return EnvMAMP
		}
		// XAMPP
		if strings.HasPrefix(absPath, "/Applications/XAMPP/") {
			return EnvXAMPP
		}
	case "linux":
		// XAMPP on Linux
		if strings.HasPrefix(absPath, "/opt/lampp/") {
			return EnvXAMPP
		}
	case "windows":
		lower := strings.ToLower(absPath)
		if strings.HasPrefix(lower, "c:\\mamp\\") {
			return EnvMAMP
		}
		if strings.HasPrefix(lower, "c:\\xampp\\") {
			return EnvXAMPP
		}
		if strings.HasPrefix(lower, "c:\\laragon\\") {
			return EnvLaragon
		}
	}

	// Valet: check if ~/.config/valet/ exists
	if homeDir, err := os.UserHomeDir(); err == nil {
		valetDir := filepath.Join(homeDir, ".config", "valet")
		if dirExists(valetDir) {
			return EnvValet
		}
	}

	return ""
}

// CommandSpec describes how to invoke WP-CLI for a given environment.
type CommandSpec struct {
	Command    string   // The binary to execute (e.g., "ddev", "lando", "wp")
	Args       []string // Arguments before the WP-CLI subcommand args
	WorkDir    string   // Working directory (project root for container-based envs)
	NeedsPath  bool     // Whether to append --path=<path>
	NeedsRunAs bool     // Whether to use sudo -u
	Env        []string // Additional environment variables
}

// BuildCommand returns the command specification for executing WP-CLI
// in the given environment.
func BuildCommand(env Environment, sitePath string, projectDir string) CommandSpec {
	switch env {
	case EnvDdev:
		return CommandSpec{
			Command: "ddev",
			Args:    []string{"wp"},
			WorkDir: projectDir,
		}
	case EnvLando:
		return CommandSpec{
			Command: "lando",
			Args:    []string{"wp"},
			WorkDir: projectDir,
		}
	case EnvWpEnv:
		return CommandSpec{
			Command: "npx",
			Args:    []string{"wp-env", "run", "cli", "wp"},
			WorkDir: projectDir,
		}
	case EnvDockerCompose:
		// Default service name; could be configurable
		return CommandSpec{
			Command: "docker",
			Args:    []string{"compose", "exec", "-T", "wordpress", "wp"},
			WorkDir: projectDir,
		}
	case EnvBedrock:
		// Bedrock: WordPress core is in web/wp/
		wpPath := filepath.Join(sitePath, "web", "wp")
		return CommandSpec{
			Command:   "wp",
			NeedsPath: true,
			Env:       phpEnv,
			Args:      []string{"--path=" + wpPath},
		}
	case EnvLocalWP:
		// Local: WordPress is in app/public/
		wpPath := sitePath
		if dirExists(filepath.Join(sitePath, "app", "public")) {
			wpPath = filepath.Join(sitePath, "app", "public")
		}
		return CommandSpec{
			Command:   "wp",
			NeedsPath: true,
			Env:       phpEnv,
			Args:      []string{"--path=" + wpPath},
		}
	case EnvMAMP:
		spec := CommandSpec{
			Command:   "wp",
			NeedsPath: true,
			Env:       phpEnv,
			Args:      []string{"--path=" + sitePath},
		}
		// Try to use MAMP's PHP
		if phpPath := findMAMPPhp(); phpPath != "" {
			spec.Env = append(spec.Env, "WP_CLI_PHP="+phpPath)
		}
		return spec
	case EnvXAMPP:
		spec := CommandSpec{
			Command:   "wp",
			NeedsPath: true,
			Env:       phpEnv,
			Args:      []string{"--path=" + sitePath},
		}
		if phpPath := findXAMPPPhp(); phpPath != "" {
			spec.Env = append(spec.Env, "WP_CLI_PHP="+phpPath)
		}
		return spec
	case EnvLaragon:
		return CommandSpec{
			Command:   "wp",
			NeedsPath: true,
			Env:       phpEnv,
			Args:      []string{"--path=" + sitePath},
		}
	case EnvValet:
		return CommandSpec{
			Command:   "wp",
			NeedsPath: true,
			Env:       phpEnv,
			Args:      []string{"--path=" + sitePath},
		}
	default: // EnvNative
		return CommandSpec{
			Command:    "wp",
			NeedsPath:  true,
			NeedsRunAs: true,
			Env:        phpEnv,
			Args:       []string{"--path=" + sitePath},
		}
	}
}

// FindProjectDir walks up from sitePath to find the project root for
// container-based environments. Returns the directory containing the
// environment marker file.
func FindProjectDir(sitePath string, env Environment) string {
	absPath, err := filepath.Abs(sitePath)
	if err != nil {
		return sitePath
	}

	dir := absPath
	for {
		switch env {
		case EnvDdev:
			if fileExists(filepath.Join(dir, ".ddev", "config.yaml")) {
				return dir
			}
		case EnvLando:
			if fileExists(filepath.Join(dir, ".lando.yml")) {
				return dir
			}
		case EnvWpEnv:
			if fileExists(filepath.Join(dir, ".wp-env.json")) {
				return dir
			}
		case EnvDockerCompose:
			for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
				if fileExists(filepath.Join(dir, name)) {
					return dir
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return sitePath
}

// --- helper functions ---

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isBedrockProject(composerFile string) bool {
	data, err := os.ReadFile(composerFile)
	if err != nil {
		return false
	}
	var composer struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(data, &composer); err != nil {
		return false
	}
	_, hasBedrock := composer.Require["roots/bedrock-autoloader"]
	_, hasWP := composer.Require["roots/wordpress"]
	return hasBedrock || hasWP
}

func hasWordPressService(composeFile string) bool {
	data, err := os.ReadFile(composeFile)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "image: wordpress") ||
		strings.Contains(content, "image: 'wordpress") ||
		strings.Contains(content, "image: \"wordpress")
}

func findMAMPPhp() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	phpDir := "/Applications/MAMP/bin/php/"
	entries, err := os.ReadDir(phpDir)
	if err != nil {
		return ""
	}
	// Find the latest PHP version
	var latest string
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "php") {
			latest = e.Name()
		}
	}
	if latest != "" {
		phpBin := filepath.Join(phpDir, latest, "bin", "php")
		if fileExists(phpBin) {
			return phpBin
		}
	}
	return ""
}

func findXAMPPPhp() string {
	candidates := []string{
		"/Applications/XAMPP/bin/php",          // macOS
		"/opt/lampp/bin/php",                   // Linux
		"C:\\xampp\\php\\php.exe",              // Windows
	}
	for _, p := range candidates {
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// EnvironmentDescription returns a short description of the environment
// for display in the wizard and status output.
func EnvironmentDescription(env Environment) string {
	switch env {
	case EnvDdev:
		return "ddev project (uses 'ddev wp')"
	case EnvLando:
		return "Lando project (uses 'lando wp')"
	case EnvWpEnv:
		return "wp-env project (uses 'npx wp-env run cli wp')"
	case EnvDockerCompose:
		return "Docker Compose (uses 'docker compose exec wordpress wp')"
	case EnvBedrock:
		return "Bedrock project (WordPress in web/wp/)"
	case EnvLocalWP:
		return "Local by Flywheel (WordPress in app/public/)"
	case EnvMAMP:
		return "MAMP (uses MAMP's PHP)"
	case EnvXAMPP:
		return "XAMPP (uses XAMPP's PHP)"
	case EnvLaragon:
		return "Laragon"
	case EnvValet:
		return "Laravel Valet"
	case EnvNative:
		return "Native wp-cli (direct execution)"
	default:
		return string(env)
	}
}
