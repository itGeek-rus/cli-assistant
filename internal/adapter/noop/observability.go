package noop

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"time"
)

type Observability struct{}

func NewObservability() *Observability {
	return &Observability{}
}

var _ observe.ObservabilityReader = (*Observability)(nil)

func (o *Observability) CheckStackHealth(
	_ context.Context,
	scope domain.Scope,
) (observe.StackHealth, error) {
	components := []observe.ComponentHealth{
		{Name: "prometheus", Status: observe.HealthHealthy, URL: scope.MetricsURL},
		{Name: "loki", Status: observe.HealthUnknown, URL: scope.LogsURL},
	}
	return observe.StackHealth{
		Overall:    observe.HealthDegraded,
		Components: components,
		CheckedAt:  time.Now().UTC(),
	}, nil
}

func (o *Observability) ListAlerts(
	_ context.Context,
	_ domain.Scope,
	_ observe.AlertFilter,
) ([]observe.Alert, error) {
	return nil, nil
}
