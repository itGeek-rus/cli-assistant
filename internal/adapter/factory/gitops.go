package factory

import (
	"cli-assistant/internal/adapter/argocd"
	"cli-assistant/internal/adapter/noop"
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
	"fmt"
	"os"
)

func NewGitOpsReader(cfg config.Config) (deploy.GitOpsReader, error) {
	scope, err := scopeFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	profile, err := cfg.ActiveProfile()
	if err != nil {
		return nil, err
	}

	switch cfg.GitOps.Provider {
	case "", "noop":
		return noop.NewGitOps(), nil
	case "argocd":
		token := ""
		if profile.GitOpsTokenEnv != "" {
			token = os.Getenv(profile.GitOpsTokenEnv)
		}
		if token == "" {
			return nil, fmt.Errorf("argocd: set token in %s", profile.GitOpsTokenEnv)
		}
		return argocd.NewFromScope(scope, token, profile.GitOpsInsecure)
	default:
		return nil, fmt.Errorf("unsupported gitops provider %q", cfg.GitOps.Provider)
	}
}

func scopeFromConfig(cfg config.Config) (domain.Scope, error) {
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
	}, nil
}
