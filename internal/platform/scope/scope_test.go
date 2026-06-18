package scope_test

import (
	"cli-assistant/internal/config"
	"cli-assistant/internal/platform/scope"
	"testing"
)

func TestFromConfig_DefaultProfile(t *testing.T) {
	cfg := config.Default()
	s, err := scope.FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if s.ProfileName != "default" {
		t.Fatalf("profile = %q, want default", s.ProfileName)
	}
}

func TestFromConfig_ProfileURLs(t *testing.T) {
	cfg := config.Default()
	cfg.Profiles["default"] = config.Profile{
		MetricsURL: "http://prom:9090",
		LogsURL:    "http://loki:3100",
		AlertsURL:  "http://alertmanager:9093",
	}

	s, err := scope.FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if s.MetricsURL != "http://prom:9090" {
		t.Fatalf("metrics_url = %q", s.MetricsURL)
	}
	if s.LogsURL != "http://loki:3100" {
		t.Fatalf("logs_url = %q", s.LogsURL)
	}
	if s.AlertsURL != "http://alertmanager:9093" {
		t.Fatalf("alerts_url = %q", s.AlertsURL)
	}
}

func TestFromConfig_UnknownProfile(t *testing.T) {
	cfg := config.Default()
	cfg.Profile = "missing"
	if _, err := scope.FromConfig(cfg); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}
