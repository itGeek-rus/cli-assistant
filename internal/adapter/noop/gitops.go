package noop

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
	"context"
	"time"
)

type GitOps struct{}

func NewGitOps() *GitOps {
	return &GitOps{}
}

var _ deploy.GitOpsReader = (*GitOps)(nil)

func (g *GitOps) ListApplications(
	_ context.Context,
	_ domain.Scope,
	_ deploy.ListFilter,
) ([]deploy.Application, error) {
	now := time.Now().UTC()
	return []deploy.Application{
		{
			Name:           "demo-app",
			Namespace:      "argocd",
			Project:        "default",
			SyncStatus:     deploy.SyncStatusSynced,
			HealthStatus:   deploy.HealthHealthy,
			Revision:       "abc123",
			TargetRevision: "main",
			LastSyncedAt:   &now,
		},
	}, nil
}

func (g *GitOps) GetApplication(
	ctx context.Context,
	scope domain.Scope,
	name string,
) (deploy.Application, error) {
	apps, err := g.ListApplications(ctx, scope, deploy.ListFilter{})
	if err != nil {
		return deploy.Application{}, err
	}
	for _, app := range apps {
		if app.Name == name {
			return app, nil
		}
	}
	return deploy.Application{}, domain.ErrNotFound
}
