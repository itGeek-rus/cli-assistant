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
	LogLevel      string              `yaml:"log_level"`
	Output        string              `yaml:"output"`
	Profile       string              `yaml:"profile"`
	GitOps        GitOpsConfig        `yaml:"gitops"`
	Observability ObservabilityConfig `yaml:"observability"`
	Profiles      map[string]Profile  `yaml:"profiles"`
}

type GitOpsConfig struct {
	Provider string `yaml:"provider"`
}

type ObservabilityConfig struct {
	Provider     string `yaml:"provider"`      // noop | prometheus
	LogsProvider string `yaml:"logs_provider"` // noop | loki (пусто = noop)
}

type Profile struct {
	KubeContext     string `yaml:"kube_context"`
	GitOpsURL       string `yaml:"gitops_url"` // Argo CD
	GitOpsTokenEnv  string `yaml:"gitops_token_env"`
	GitOpsInsecure  bool   `yaml:"gitops_insecure"`
	MetricsURL      string `yaml:"metrics_url"` // Prometheus
	MetricsInsecure bool   `yaml:"metrics_insecure"`
	LogsURL         string `yaml:"logs_url"` // Loki
	LogsInsecure    bool   `yaml:"logs_insecure"`
}

func Default() Config {
	return Config{
		LogLevel: "info",
		Output:   "human",
		Profile:  "default",
		GitOps: GitOpsConfig{
			Provider: "noop",
		},
		Observability: ObservabilityConfig{
			Provider: "noop",
		},
		Profiles: map[string]Profile{
			"default": { // #nosec: G101
				GitOpsTokenEnv: "ARGOCD_AUTH_TOKEN",
			},
		},
	}
}

func Dir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config dir: %w", err)
	}
	return filepath.Join(dir, DefaultAppName), nil
}

func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (Config, error) {
	cfg := Default()

	path, err := Path()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path) // #nosec:G304
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
	if v := os.Getenv(EnvPrefix + "_GITOPS_PROVIDER"); v != "" {
		cfg.GitOps.Provider = v
	}
	if v := os.Getenv(EnvPrefix + "_OBSERVABILITY_PROVIDER"); v != "" {
		cfg.Observability.Provider = v
	}
	if v := os.Getenv(EnvPrefix + "_LOGS_PROVIDER"); v != "" {
		cfg.Observability.LogsProvider = v
	}
}

func (c Config) ActiveProfile() (Profile, error) {
	p, ok := c.Profiles[c.Profile]
	if !ok {
		return Profile{}, fmt.Errorf("profile %q not found", c.Profile)
	}
	return p, nil
}
