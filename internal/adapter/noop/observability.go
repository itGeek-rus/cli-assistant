package noop

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"fmt"
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

func (o *Observability) QueryMetrics(
	_ context.Context,
	_ domain.Scope,
	req observe.QueryRequest,
) (observe.QueryResult, error) {
	if req.Expr == "" {
		return observe.QueryResult{}, fmt.Errorf("%w: query expression is required", domain.ErrInvalidInput)
	}
	now := time.Now().UTC()
	return observe.QueryResult{
		Expr: req.Expr,
		Samples: []observe.QuerySample{
			{
				Labels: map[string]string{"job": "demo", "__name__": "up"},
				Value:  1,
				Time:   now,
			},
		},
	}, nil
}

func (o *Observability) ListAlerts(
	_ context.Context,
	_ domain.Scope,
	filter observe.AlertFilter,
) ([]observe.Alert, error) {
	alerts := []observe.Alert{
		{
			Name:     "DemoHighCPU",
			Severity: observe.AlertSeverityWarning,
			State:    observe.AlertStateFiring,
			Summary:  "CPU > 80% on demo-app",
			StartAt:  time.Now().UTC().Add(-15 * time.Minute),
			Labels:   map[string]string{"app": "demo-app"},
		},
	}
	if filter.State != "" {
		out := make([]observe.Alert, 0)
		for _, a := range alerts {
			if a.State == filter.State {
				out = append(out, a)
			}
		}
		return out, nil
	}
	return alerts, nil
}
