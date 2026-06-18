package composite

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"fmt"
)

type LogsQuerier interface {
	QueryLog(ctx context.Context, scope domain.Scope, req observe.LogsRequest) (observe.LogResult, error)
}

type AlertsLister interface {
	ListAlerts(ctx context.Context, scope domain.Scope, filter observe.AlertFilter) ([]observe.Alert, error)
}

type Reader struct {
	core   observe.ObservabilityReader // metrics + health + alert
	alerts AlertsLister
	logs   LogsQuerier
}

func New(core observe.ObservabilityReader, alerts AlertsLister, logs LogsQuerier) *Reader {
	return &Reader{core: core, alerts: alerts, logs: logs}
}

var _ observe.ObservabilityReader = (*Reader)(nil)

func (r *Reader) CheckStackHealth(ctx context.Context, scope domain.Scope) (observe.StackHealth, error) {
	return r.core.CheckStackHealth(ctx, scope)
}

func (r *Reader) ListAlerts(ctx context.Context, scope domain.Scope, filter observe.AlertFilter) ([]observe.Alert, error) {
	if r.alerts != nil {
		return r.alerts.ListAlerts(ctx, scope, filter)
	}
	return r.core.ListAlerts(ctx, scope, filter)
}

func (r *Reader) QueryMetrics(ctx context.Context, scope domain.Scope, req observe.QueryRequest) (observe.QueryResult, error) {
	return r.core.QueryMetrics(ctx, scope, req)
}

func (r *Reader) QueryLog(ctx context.Context, scope domain.Scope, req observe.LogsRequest) (observe.LogResult, error) {
	if r.logs == nil {
		return observe.LogResult{}, fmt.Errorf("%w: logs provider is not configured", domain.ErrUnavailable)
	}
	return r.logs.QueryLog(ctx, scope, req)
}
