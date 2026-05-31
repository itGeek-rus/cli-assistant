package main

import (
	"cli-assistant/internal/adapter/factory"
	"fmt"
	"os"

	"cli-assistant/internal/adapter/noop"
	"cli-assistant/internal/config"
	"cli-assistant/internal/delivery/cli"
	"cli-assistant/internal/usecase"
	"cli-assistant/pkg/log"
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

	gitops, err := factory.NewGitOpsReader(cfg)
	if err != nil {
		return nil, fmt.Errorf("gitops reader: %w", err)
	}

	dep := usecase.NewDeployment(cfg, logger, gitops)
	obs := usecase.NewObservability(logger, cfg, noop.NewObservability())

	return &container{
		app: cli.NewApp(cfg, obs, dep, logger),
	}, nil
}
