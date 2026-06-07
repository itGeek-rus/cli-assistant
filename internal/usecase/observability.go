package usecase

import (
	"cli-assistant/internal/domain"
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

type HealthResult struct {
	Health observe.StackHealth
}

type QueryResult struct {
	Result observe.QueryResult
}

type AlertResult struct {
	Alerts []observe.Alert
}

func (u *Observability) Health(ctx context.Context) (HealthResult, error) {
	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return HealthResult{}, err
	}

	health, err := u.reader.CheckStackHealth(ctx, scope)
	if err != nil {
		return HealthResult{}, fmt.Errorf("check stack health: %w", err)
	}

	u.log.DebugContext(ctx, "observability stack",
		"overall", health.Overall,
		"components", len(health.Components),
	)
	return HealthResult{Health: health}, nil
}

func (u *Observability) Query(ctx context.Context, expr string) (QueryResult, error) {
	if expr == "" {
		return QueryResult{}, fmt.Errorf("%w: query expression is required", domain.ErrInvalidInput)
	}
	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return QueryResult{}, err
	}
	res, err := u.reader.QueryMetrics(ctx, scope, observe.QueryRequest{Expr: expr})
	if err != nil {
		return QueryResult{}, fmt.Errorf("query metrics: %w", err)
	}
	return QueryResult{Result: res}, nil
}

func (u *Observability) Alert(ctx context.Context, filter observe.AlertFilter) (AlertResult, error) {
	scope, err := scopeFromConfig(u.cfg)
	if err != nil {
		return AlertResult{}, err
	}
	alerts, err := u.reader.ListAlerts(ctx, scope, filter)
	if err != nil {
		return AlertResult{}, fmt.Errorf("list alerts: %w", err)
	}
	return AlertResult{Alerts: alerts}, nil
}
