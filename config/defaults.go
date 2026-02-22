package config

func applyDefaults(cfg *Config) {
	if cfg.Settings.SendPolicy == "" {
		cfg.Settings.SendPolicy = "only-on-updates"
	}
}

// NewDefault returns a config with sensible defaults.
func NewDefault() *Config {
	cfg := &Config{
		Settings: GlobalSettings{
			SendPolicy: "only-on-updates",
		},
	}
	return cfg
}
