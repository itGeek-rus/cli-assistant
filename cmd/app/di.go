package main

import (
	"cli-assistant/internal/adapter/factory"
	"cli-assistant/internal/adapter/postgres"
	"fmt"
	"os"

	"cli-assistant/internal/config"
	"cli-assistant/internal/delivery/cli"
	httpserver "cli-assistant/internal/delivery/http"
	"cli-assistant/internal/domain/history"
	"cli-assistant/internal/usecase"
	"cli-assistant/pkg/log"
	"cli-assistant/pkg/version"
)

type container struct {
	app *cli.App
}

func newContainer() (*container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	logger := log.New(cfg.LogLevel, os.Stderr)

	gitops, err := factory.NewGitOpsClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("gitops client: %w", err)
	}

	dep := usecase.NewDeployment(logger, cfg, gitops)

	obsReader, err := factory.NewObservabilityReader(cfg)
	if err != nil {
		return nil, fmt.Errorf("observability reader: %w", err)
	}

	inspector := usecase.NewInspector(logger, cfg, gitops, obsReader)

	obs := usecase.NewObservability(logger, cfg, obsReader)

	var hist history.Store
	if cfg.Database.URL != "" {
		store, err := postgres.New(cfg.Database.URL)
		if err != nil {
			logger.Warn("postgres unavailable, command history disabled", "error", err)
		} else {
			hist = store
		}
	}

	httpSrv := httpserver.New(dep, obs, inspector, hist, httpserver.Options{
		Addr:    cfg.API.Addr,
		Token:   cfg.API.Token,
		Profile: cfg.Profile,
	})

	return &container{
		app: cli.NewApp(cfg, obs, dep, inspector, httpSrv, hist, logger,
			version.Version, version.Commit, version.BuildDate),
	}, nil
}
