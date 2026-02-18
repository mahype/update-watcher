package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/mahype/update-watcher/internal/hostname"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

// Config is the top-level configuration structure.
type Config struct {
	Hostname  string           `yaml:"hostname" mapstructure:"hostname"`
	Watchers  []WatcherConfig  `yaml:"watchers" mapstructure:"watchers"`
	Notifiers []NotifierConfig `yaml:"notifiers" mapstructure:"notifiers"`
	Settings  GlobalSettings   `yaml:"settings" mapstructure:"settings"`
}

// WatcherConfig represents a single watcher entry.
type WatcherConfig struct {
	Type    string                 `yaml:"type" mapstructure:"type"`
	Enabled bool                   `yaml:"enabled" mapstructure:"enabled"`
	Options map[string]interface{} `yaml:"options,omitempty" mapstructure:"options"`
}

// NotifierConfig represents a single notifier entry.
type NotifierConfig struct {
	Type    string                 `yaml:"type" mapstructure:"type"`
	Enabled bool                   `yaml:"enabled" mapstructure:"enabled"`
	Options map[string]interface{} `yaml:"options,omitempty" mapstructure:"options"`
}

// GlobalSettings holds cross-cutting settings.
type GlobalSettings struct {
	SendPolicy string `yaml:"send_policy" mapstructure:"send_policy"`
	LogFile    string `yaml:"log_file,omitempty" mapstructure:"log_file"`
	Schedule   string `yaml:"schedule,omitempty" mapstructure:"schedule"`
	Quiet      bool   `yaml:"quiet,omitempty" mapstructure:"quiet"`
}

// GetBool reads a boolean from Options with a default fallback.
func (w WatcherConfig) GetBool(key string, defaultVal bool) bool {
	v, ok := w.Options[key]
	if !ok {
		return defaultVal
	}
	b, ok := v.(bool)
	if !ok {
		return defaultVal
	}
	return b
}

// GetString reads a string from Options with a default fallback.
func (w WatcherConfig) GetString(key string, defaultVal string) string {
	v, ok := w.Options[key]
	if !ok {
		return defaultVal
	}
	s, ok := v.(string)
	if !ok {
		return defaultVal
	}
	return s
}

// GetStringSlice reads a string slice from Options with a default fallback.
func (w WatcherConfig) GetStringSlice(key string, defaultVal []string) []string {
	v, ok := w.Options[key]
	if !ok {
		return defaultVal
	}
	switch val := v.(type) {
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return val
	default:
		return defaultVal
	}
}

// GetMapSlice reads a slice of maps from Options.
func (w WatcherConfig) GetMapSlice(key string) []map[string]interface{} {
	v, ok := w.Options[key]
	if !ok {
		return nil
	}
	switch val := v.(type) {
	case []interface{}:
		result := make([]map[string]interface{}, 0, len(val))
		for _, item := range val {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result
	default:
		return nil
	}
}

// envVarPattern matches ${VAR} and ${VAR:-default} patterns.
var envVarPattern = regexp.MustCompile(`\$\{([^}:]+)(?::-([^}]*))?\}`)

// resolveEnvVars replaces ${VAR} and ${VAR:-default} in a string with
// environment variable values. Unset variables resolve to empty string
// unless a default is provided.
func resolveEnvVars(input string) string {
	return envVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		groups := envVarPattern.FindStringSubmatch(match)
		varName := groups[1]
		defaultVal := groups[2]

		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		return defaultVal
	})
}

// resolveOptionsEnvVars walks an options map and resolves ${ENV_VAR} references
// in all string values, including those nested in slices and maps.
func resolveOptionsEnvVars(options map[string]interface{}) {
	for key, val := range options {
		switch v := val.(type) {
		case string:
			options[key] = resolveEnvVars(v)
		case []interface{}:
			for i, item := range v {
				if s, ok := item.(string); ok {
					v[i] = resolveEnvVars(s)
				} else if m, ok := item.(map[string]interface{}); ok {
					resolveOptionsEnvVars(m)
				}
			}
		case map[string]interface{}:
			resolveOptionsEnvVars(v)
		}
	}
}

// resolveConfigEnvVars resolves ${ENV_VAR} references in all watcher and notifier options.
func resolveConfigEnvVars(cfg *Config) {
	for i := range cfg.Watchers {
		if cfg.Watchers[i].Options != nil {
			resolveOptionsEnvVars(cfg.Watchers[i].Options)
		}
	}
	for i := range cfg.Notifiers {
		if cfg.Notifiers[i].Options != nil {
			resolveOptionsEnvVars(cfg.Notifiers[i].Options)
		}
	}
}

// Load reads the configuration from Viper.
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Hostname == "" {
		cfg.Hostname = hostname.Get()
	}

	applyDefaults(&cfg)
	resolveConfigEnvVars(&cfg)

	return &cfg, nil
}

// Save writes the configuration to the given path.
func Save(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfigDir returns the platform-appropriate config directory.
// Linux: /etc/update-watcher (system-wide)
// macOS/other: ~/.config/update-watcher (user-level)
func DefaultConfigDir() string {
	if runtime.GOOS == "linux" {
		return "/etc/update-watcher"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".config/update-watcher"
	}
	return filepath.Join(home, ".config", "update-watcher")
}

// ConfigSearchPaths returns all config directories to search, in priority order.
func ConfigSearchPaths() []string {
	var paths []string
	if runtime.GOOS == "linux" {
		paths = append(paths, "/etc/update-watcher")
	}
	home, err := os.UserHomeDir()
	if err == nil {
		paths = append(paths, filepath.Join(home, ".config", "update-watcher"))
	}
	return paths
}

// ConfigPath returns the path where the config file is loaded from, or the default path.
func ConfigPath() string {
	if f := viper.ConfigFileUsed(); f != "" {
		return f
	}
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

// AddWatcher adds a watcher to the config. If a watcher of the same type
// already exists, it will be replaced (except for WordPress, which supports multiple).
func (c *Config) AddWatcher(watcher WatcherConfig) {
	if watcher.Type != "wordpress" && watcher.Type != "webproject" {
		for i, w := range c.Watchers {
			if w.Type == watcher.Type {
				c.Watchers[i] = watcher
				return
			}
		}
	}
	c.Watchers = append(c.Watchers, watcher)
}

// RemoveWatcher removes a watcher by type. For WordPress and webproject, an
// optional name parameter can target a specific site/project (removing only
// that entry, not the entire watcher). If name is empty, the entire watcher
// is removed.
func (c *Config) RemoveWatcher(watcherType string, name string) bool {
	for i, w := range c.Watchers {
		if w.Type != watcherType {
			continue
		}
		if name != "" {
			var listKey string
			switch watcherType {
			case "wordpress":
				listKey = "sites"
			case "webproject":
				listKey = "projects"
			}
			if listKey != "" {
				items := w.GetMapSlice(listKey)
				for j, item := range items {
					if itemName, ok := item["name"].(string); ok && itemName == name {
						remaining := make([]interface{}, 0, len(items)-1)
						for k, s := range items {
							if k != j {
								remaining = append(remaining, s)
							}
						}
						if len(remaining) == 0 {
							c.Watchers = append(c.Watchers[:i], c.Watchers[i+1:]...)
						} else {
							c.Watchers[i].Options[listKey] = remaining
						}
						return true
					}
				}
				continue
			}
		}
		c.Watchers = append(c.Watchers[:i], c.Watchers[i+1:]...)
		return true
	}
	return false
}

// HasWatcher checks if a watcher of the given type exists.
func (c *Config) HasWatcher(watcherType string) bool {
	for _, w := range c.Watchers {
		if w.Type == watcherType {
			return true
		}
	}
	return false
}
