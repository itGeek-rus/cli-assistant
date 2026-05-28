package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	EnvPrefix      = "CLI_ASSISTANT"
	DefaultAppName = "cli_assistant"
)

type Config struct {
	LogLevel string             `yaml:"log_level"`
	Output   string             `yaml:"output"`
	Profile  string             `yaml:"profile"`
	Profiles map[string]Profile `yaml:"profiles"`
}

type Profile struct {
	KubeContext string `yaml:"kube_context"`
	GitOpsURL   string `yaml:"gitops_url"`  // Argo CD
	MetricsURL  string `yaml:"metrics_url"` // Prometheus
	LogsURL     string `yaml:"logs_url"`    // Loki
}

func Default() Config {
	return Config{
		LogLevel: "info",
		Output:   "human",
		Profile:  "default",
		Profiles: map[string]Profile{
			"default": {},
		},
	}
}

func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, DefaultAppName), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (Config, error) {
	cfg := Default()

	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			applyEnv(&cfg)
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}

	applyEnv(&cfg)
	return cfg, nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv(EnvPrefix + "_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv(EnvPrefix + "_OUTPUT"); v != "" {
		cfg.Output = v
	}
	if v := os.Getenv(EnvPrefix + "_PROFILE"); v != "" {
		cfg.Profile = v
	}
}

func (c Config) ActiveProfile() (Profile, error) {
	p, ok := c.Profiles[c.Profile]
	if !ok {
		return Profile{}, fmt.Errorf("profile %q not found", c.Profile)
	}
	return p, nil
}
