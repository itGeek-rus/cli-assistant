package factory

import (
	"cli-assistant/internal/adapter/alertmanager"
	"cli-assistant/internal/adapter/composite"
	"cli-assistant/internal/adapter/loki"
	"cli-assistant/internal/adapter/noop"
	"cli-assistant/internal/adapter/prometheus"
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain/observe"
	"cli-assistant/internal/platform/scope"
	"fmt"
)

func NewObservabilityReader(cfg config.Config) (observe.ObservabilityReader, error) {
	s, err := scope.FromConfig(cfg)
	if err != nil {
		return nil, err
	}
	profile, err := cfg.ActiveProfile()
	if err != nil {
		return nil, err
	}

	var core observe.ObservabilityReader
	switch cfg.Observability.Provider {
	case "", "noop":
		return noop.NewObservability(), nil
	case "prometheus":
		if s.MetricsURL == "" {
			return nil, fmt.Errorf("prometheus: metrics_url is required in profile %q", cfg.Profile)
		}
		core, err = prometheus.NewFromScope(s, profile.MetricsInsecure)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported observability provider %q", cfg.Observability.Provider)
	}

	alertsProvider := cfg.Observability.AlertsProvider
	if alertsProvider == "" {
		alertsProvider = "noop"
	}
	var alerts composite.AlertsLister
	switch alertsProvider {
	case "noop":
		alerts = noop.NewAlerts()
	case "alertmanager":
		if s.AlertsURL == "" {
			return nil, fmt.Errorf("alertmanager: alerts_url is required in profile %q", cfg.Profile)
		}
		alerts, err = alertmanager.NewFromScope(s, profile.AlertsInsecure)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported alerts provider %q", alertsProvider)
	}

	logsProvider := cfg.Observability.LogsProvider
	if logsProvider == "" {
		logsProvider = "noop"
	}
	var logs composite.LogsQuerier
	switch logsProvider {
	case "noop":
		logs = noop.NewObservability()
	case "loki":
		if s.LogsURL == "" {
			return nil, fmt.Errorf("loki: logs_url is required in profile %q", cfg.Profile)
		}
		logs, err = loki.NewFromScope(s, profile.LogsInsecure)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported logs provider %q", logsProvider)
	}
	return composite.New(core, alerts, logs), nil
}
