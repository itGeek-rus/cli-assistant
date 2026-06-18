package main

import (
	"cli-assistant/internal/adapter/factory"
	"fmt"
	"os"

	"cli-assistant/internal/config"
	"cli-assistant/internal/delivery/cli"
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

	return &container{
		app: cli.NewApp(cfg, obs, dep, inspector, logger,
			version.Version, version.Commit, version.BuildDate),
	}, nil
}
