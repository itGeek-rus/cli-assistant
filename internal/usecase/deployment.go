package usecase

import (
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type Deployment struct {
	log    *slog.Logger
	cfg    config.Config
	gitops deploy.GitOpsClient
}

func NewDeployment(log *slog.Logger, cfg config.Config, gitops deploy.GitOpsClient) *Deployment {
	return &Deployment{log: log, cfg: cfg, gitops: gitops}
}

type ListResult struct {
	Applications []deploy.Application
}

type StatusResult struct {
	Application deploy.Application
}

func (u *Deployment) List(ctx context.Context, filter deploy.ListFilter) (ListResult, error) {
	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return ListResult{}, err
	}
	apps, err := u.gitops.ListApplications(ctx, scope, filter)
	if err != nil {
		return ListResult{}, fmt.Errorf("list applications: %w", err)
	}
	return ListResult{Applications: apps}, nil
}

func (u *Deployment) Status(ctx context.Context, name string) (StatusResult, error) {
	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return StatusResult{}, err
	}

	if name == "" {
		apps, err := u.gitops.ListApplications(ctx, scope, deploy.ListFilter{
			Namespace: scope.Namespace,
		})
		if err != nil {
			return StatusResult{}, fmt.Errorf("list applications: %w", err)
		}
		if len(apps) == 0 {
			return StatusResult{}, domain.ErrNotFound
		}
		return StatusResult{}, fmt.Errorf("%w: specify application name or use deploy list", domain.ErrInvalidInput)
	}

	app, err := u.gitops.GetApplication(ctx, scope, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return StatusResult{}, err
		}
		return StatusResult{}, fmt.Errorf("get application: %w", err)
	}

	return StatusResult{Application: app}, nil
}

func (u *Deployment) Sync(ctx context.Context, name string, opts deploy.SyncOptions, confirmed bool) (deploy.SyncResult, error) {
	if name == "" {
		return deploy.SyncResult{}, fmt.Errorf("%w: application name is required", domain.ErrInvalidInput)
	}

	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return deploy.SyncResult{}, err
	}

	if opts.DryRun {
		diff, err := u.gitops.DiffApplication(ctx, scope, name)
		if err != nil {
			return deploy.SyncResult{}, fmt.Errorf("diff application: %w", err)
		}
		msg := syncDryRunMessage(diff)
		return deploy.SyncResult{
			Application: name,
			DryRun:      true,
			Initiated:   false,
			Message:     msg,
		}, nil
	}

	if !confirmed {
		return deploy.SyncResult{}, fmt.Errorf("%w: confirmation required (use --yes)", domain.ErrInvalidInput)
	}

	res, err := u.gitops.SyncApplication(ctx, scope, name, opts)
	if err != nil {
		return deploy.SyncResult{}, fmt.Errorf("sync application: %w", err)
	}
	return res, nil
}

func (u *Deployment) Diff(ctx context.Context, name string) (deploy.DiffResult, error) {
	if name == "" {
		return deploy.DiffResult{}, fmt.Errorf("%w: application name is required", domain.ErrInvalidInput)
	}
	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return deploy.DiffResult{}, err
	}
	diff, err := u.gitops.DiffApplication(ctx, scope, name)
	if err != nil {
		return deploy.DiffResult{}, fmt.Errorf("diff application: %w", err)
	}
	return diff, nil
}

func syncDryRunMessage(diff deploy.DiffResult) string {
	if diff.Raw != "" {
		return "dry-run: " + diff.Raw
	}
	if diff.OutOfSync {
		return fmt.Sprintf("dry-run: application %q is out of sync", diff.Application)
	}
	return fmt.Sprintf("dry-run: application %q is in sync", diff.Application)
}
