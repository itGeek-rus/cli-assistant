package config

import (
	"fmt"
	"os"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (c Config) Validate() error {
	p, err := c.ActiveProfile()
	if err != nil {
		return err
	}

	switch c.GitOps.Provider {
	case "noop":
		// ok
	case "argocd":
		if strings.TrimSpace(p.GitOpsURL) == "" {
			return ValidationError{Field: "profiles.*.gitops_url", Message: "required when gitops.provider=argocd"}
		}
		envName := p.GitOpsTokenEnv
		if envName == "" {
			envName = "ARGOCD_AUTH_TOKEN"
		}
		if os.Getenv(envName) == "" {
			return ValidationError{
				Field:   envName,
				Message: fmt.Sprintf("set %s (Argo CD UI -> User Info -> Generate New Token)", envName),
			}
		}
	default:
		return ValidationError{Field: "gitops.provider", Message: fmt.Sprintf("unsupported: %q", c.GitOps.Provider)}
	}
	if c.Observability.Provider == "prometheus" && strings.TrimSpace(p.MetricsURL) == "" {
		return ValidationError{Field: "profiles.*.metrics_url", Message: "required when observability.provider=prometheus"}
	}
	if c.Observability.LogsProvider == "loki" && strings.TrimSpace(p.LogsURL) == "" {
		return ValidationError{Field: "profiles.*.logs_url", Message: "required when logs_provider=loki"}
	}
	if c.Observability.AlertsProvider == "alertmanager" && strings.TrimSpace(p.AlertsURL) == "" {
		return ValidationError{Field: "profiles.*.alerts_url", Message: "required when alerts_provider=alertmanager"}
	}
	return nil
}
