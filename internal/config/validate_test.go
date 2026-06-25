package config_test

import (
	"cli-assistant/internal/config"
	"testing"
)

func TestValidate_noopDefault(t *testing.T) {
	cfg := config.Default()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("default noop config: %v", err)
	}
}

func TestValidate_argocdRequiresURLAndToken(t *testing.T) {
	cfg := config.Default()
	cfg.GitOps.Provider = "argocd"
	cfg.Profiles["default"] = config.Profile{
		GitOpsURL:      "https://argocd.example.com",
		GitOpsTokenEnv: "ARGOCD_AUTH_TOKEN",
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when token env is missing")
	}

	t.Setenv("ARGOCD_AUTH_TOKEN", "test-token")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config with token: %v", err)
	}
}

func TestValidate_observabilityProviders(t *testing.T) {
	cfg := config.Default()
	cfg.Observability.Provider = "prometheus"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when metrics_url is missing")
	}

	cfg.Profiles["default"] = config.Profile{MetricsURL: "http://localhost:9090"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("prometheus config: %v", err)
	}

	cfg.Observability.LogsProvider = "loki"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when logs_url is missing")
	}

	cfg.Profiles["default"] = config.Profile{
		MetricsURL: "http://localhost:9090",
		LogsURL:    "http://localhost:3100",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("loki config: %v", err)
	}

	cfg.Observability.AlertsProvider = "alertmanager"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when alerts_url is missing")
	}

	cfg.Profiles["default"] = config.Profile{
		MetricsURL: "http://localhost:9090",
		LogsURL:    "http://localhost:3100",
		AlertsURL:  "http://localhost:9093",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("alertmanager config: %v", err)
	}
}

func TestValidationErrorMessage(t *testing.T) {
	err := config.ValidationError{Field: "ARGOCD_AUTH_TOKEN", Message: "set token"}
	if got := err.Error(); got != "ARGOCD_AUTH_TOKEN: set token" {
		t.Fatalf("got %q", got)
	}
}
