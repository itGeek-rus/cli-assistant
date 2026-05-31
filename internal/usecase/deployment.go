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
	gitops deploy.GitOpsReader
}

func NewDeployment(cfg config.Config, log *slog.Logger, gitops deploy.GitOpsReader) *Deployment {
	return &Deployment{cfg: cfg, log: log, gitops: gitops}
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
