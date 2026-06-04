package noop

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
	"context"
	"fmt"
	"time"
)

type GitOps struct{}

func NewGitOps() *GitOps {
	return &GitOps{}
}

var (
	_ deploy.GitOpsReader = (*GitOps)(nil)
	_ deploy.GitOpsClient = (*GitOps)(nil)
)

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

func (g *GitOps) SyncApplication(
	_ context.Context,
	_ domain.Scope,
	name string,
	opts deploy.SyncOptions,
) (deploy.SyncResult, error) {
	if name == "" {
		return deploy.SyncResult{}, fmt.Errorf("%w: application name is required", domain.ErrInvalidInput)
	}
	msg := "sync completed (noop)"
	if opts.DryRun {
		msg = "dry run: sync would be triggered (noop)"
	}
	return deploy.SyncResult{
		Application: name,
		DryRun:      opts.DryRun,
		Initiated:   !opts.DryRun,
		Message:     msg,
	}, nil
}

func (g *GitOps) DiffApplication(
	_ context.Context,
	_ domain.Scope,
	name string,
) (deploy.DiffResult, error) {
	if name == "" {
		return deploy.DiffResult{}, fmt.Errorf("%w: application name is required", domain.ErrInvalidInput)
	}
	return deploy.DiffResult{
		Application: name,
		OutOfSync:   true,
		Changes: []deploy.DiffChange{
			{Kind: "Deployment", Name: "demo", Namespace: "default", Summary: "1 modified"},
		},
		Raw: "--- noop diff\n+++ cluster\n demo-app would change",
	}, nil
}
