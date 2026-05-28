package main

import (
	"fmt"
	"os"

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

	dep := usecase.NewDeployment(logger)
	obs := usecase.NewObservability(logger)

	return &container{
		app: cli.NewApp(cfg, *obs, dep, logger),
	}, nil
}
