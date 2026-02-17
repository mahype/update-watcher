package config

func applyDefaults(cfg *Config) {
	if cfg.Settings.SendPolicy == "" {
		cfg.Settings.SendPolicy = "only-on-updates"
	}
	if cfg.Settings.Schedule == "" {
		cfg.Settings.Schedule = "0 7 * * *"
	}
}

// NewDefault returns a config with sensible defaults.
func NewDefault() *Config {
	cfg := &Config{
		Settings: GlobalSettings{
			SendPolicy: "only-on-updates",
			Schedule:   "0 7 * * *",
		},
	}
	return cfg
}
