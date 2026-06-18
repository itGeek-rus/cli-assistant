package scope

import (
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain"
	"fmt"
)

func FromConfig(cfg config.Config) (domain.Scope, error) {
	profile, err := cfg.ActiveProfile()
	if err != nil {
		return domain.Scope{}, fmt.Errorf("active profile: %w", err)
	}
	return domain.Scope{
		ProfileName: cfg.Profile,
		KubeContext: profile.KubeContext,
		GitOpsURL:   profile.GitOpsURL,
		MetricsURL:  profile.MetricsURL,
		LogsURL:     profile.LogsURL,
		AlertsURL:   profile.AlertsURL,
	}, nil
}
