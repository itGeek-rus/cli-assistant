package factory

import (
	"cli-assistant/internal/adapter/argocd"
	"cli-assistant/internal/adapter/noop"
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain/deploy"
	"cli-assistant/internal/platform/scope"
	"fmt"
	"os"
)

func NewGitOpsReader(cfg config.Config) (deploy.GitOpsReader, error) {
	s, err := scope.FromConfig(cfg)
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
		return argocd.NewFromScope(s, token, profile.GitOpsInsecure)
	default:
		return nil, fmt.Errorf("unsupported gitops provider %q", cfg.GitOps.Provider)
	}
}

func NewGitOpsClient(cfg config.Config) (deploy.GitOpsClient, error) {
	reader, err := NewGitOpsReader(cfg)
	if err != nil {
		return nil, err
	}
	client, ok := reader.(deploy.GitOpsClient)
	if !ok {
		return nil, fmt.Errorf("gitops provider %q does not support write operations", cfg.GitOps.Provider)
	}
	return client, nil
}
