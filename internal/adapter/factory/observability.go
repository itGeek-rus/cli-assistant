package factory

import (
	"cli-assistant/internal/adapter/noop"
	"cli-assistant/internal/adapter/prometheus"
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain/observe"
	"fmt"
)

func NewObservabilityReader(cfg config.Config) (observe.ObservabilityReader, error) {
	scope, err := scopeFromConfig(cfg)
	if err != nil {
		return nil, err
	}

	switch cfg.Observability.Provider {
	case "", "noop":
		return noop.NewObservability(), nil
	case "prometheus":
		if scope.MetricsURL == "" {
			return nil, fmt.Errorf("prometheus: metrics_url is required in profile %q", cfg.Profile)
		}
		profile, err := cfg.ActiveProfile()
		if err != nil {
			return nil, err
		}
		return prometheus.NewFromScope(scope, profile.GitOpsInsecure)
	default:
		return nil, fmt.Errorf("unsupported observability provider: %q", cfg.Observability.Provider)
	}
}
