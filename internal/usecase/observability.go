package usecase

import (
	"context"
	"fmt"
	"log/slog"

	"cli-assistant/internal/config"
	"cli-assistant/internal/domain/observe"
)

type Observability struct {
	log    *slog.Logger
	cfg    config.Config
	reader observe.ObservabilityReader
}

func NewObservability(log *slog.Logger, cfg config.Config, reader observe.ObservabilityReader) *Observability {
	return &Observability{log: log, cfg: cfg, reader: reader}
}

func (u *Observability) Health(ctx context.Context) error {
	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return err
	}

	health, err := u.reader.CheckStackHealth(ctx, scope)
	if err != nil {
		return fmt.Errorf("check stack health: %w", err)
	}

	u.log.InfoContext(ctx, "observability stack",
		"overall", health.Overall,
		"checked_at", health.CheckedAt,
	)
	for _, c := range health.Components {
		u.log.InfoContext(ctx, "component",
			"name", c.Name,
			"status", c.Status,
			"message", c.Message,
			"url", c.URL,
		)
	}
	return nil
}
