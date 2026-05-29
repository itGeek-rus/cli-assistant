package usecase

import (
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain/deploy"
	"context"
	"fmt"
	"log/slog"
)

type Deployment struct {
	log    *slog.Logger
	cfg    config.Config
	gitops deploy.GitOpsReader
}

func NewDeployment(cfg config.Config, log *slog.Logger, gitops deploy.GitOpsReader) *Deployment {
	return &Deployment{cfg: cfg, log: log, gitops: gitops}
}

func (u *Deployment) Status(ctx context.Context) error {
	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return err
	}

	apps, err := u.gitops.ListApplications(ctx, scope, deploy.ListFilter{
		Namespace: scope.Namespace,
	})
	if err != nil {
		return fmt.Errorf("list application: %w", err)
	}

	for _, app := range apps {
		u.log.InfoContext(ctx, "application",
			"name", app.Name,
			"namespace", app.Namespace,
			"sync", app.SyncStatus,
			"health", app.HealthStatus,
			"revision", app.Revision,
		)
	}
	return nil
}
