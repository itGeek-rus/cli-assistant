package factory_test

import (
	"testing"

	"cli-assistant/internal/adapter/composite"
	"cli-assistant/internal/adapter/factory"
	"cli-assistant/internal/adapter/noop"
	"cli-assistant/internal/config"
)

func TestNewObservabilityReader_NoopProvider(t *testing.T) {
	cfg := config.Default()
	reader, err := factory.NewObservabilityReader(cfg)
	if err != nil {
		t.Fatalf("NewObservabilityReader: %v", err)
	}
	if _, ok := reader.(*noop.Observability); !ok {
		t.Fatalf("noop provider: got %T, want *noop.Observability", reader)
	}
}

func TestNewObservabilityReader_PrometheusComposite(t *testing.T) {
	cfg := config.Default()
	cfg.Observability.Provider = "prometheus"
	cfg.Observability.LogsProvider = "noop"
	cfg.Observability.AlertsProvider = "noop"
	cfg.Profiles["default"] = config.Profile{
		MetricsURL: "http://localhost:9090",
	}

	reader, err := factory.NewObservabilityReader(cfg)
	if err != nil {
		t.Fatalf("NewObservabilityReader: %v", err)
	}
	if _, ok := reader.(*composite.Reader); !ok {
		t.Fatalf("prometheus provider: got %T, want *composite.Reader", reader)
	}
}

func TestNewObservabilityReader_AlertmanagerRequiresURL(t *testing.T) {
	cfg := config.Default()
	cfg.Observability.Provider = "prometheus"
	cfg.Observability.AlertsProvider = "alertmanager"
	cfg.Profiles["default"] = config.Profile{
		MetricsURL: "http://localhost:9090",
	}

	_, err := factory.NewObservabilityReader(cfg)
	if err == nil {
		t.Fatal("expected error when alerts_url is missing")
	}
}

func TestNewObservabilityReader_UnknownAlertsProvider(t *testing.T) {
	cfg := config.Default()
	cfg.Observability.Provider = "prometheus"
	cfg.Observability.AlertsProvider = "unknown"
	cfg.Profiles["default"] = config.Profile{
		MetricsURL: "http://localhost:9090",
		AlertsURL:  "http://localhost:9093",
	}

	_, err := factory.NewObservabilityReader(cfg)
	if err == nil {
		t.Fatal("expected error for unknown alerts provider")
	}
}
