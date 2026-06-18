package noop

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
)

type Alerts struct{}

func NewAlerts() *Alerts {
	return &Alerts{}
}

func (a *Alerts) ListAlerts(
	_ context.Context,
	_ domain.Scope,
	_ observe.AlertFilter,
) ([]observe.Alert, error) {
	return nil, nil // пока заглушки как у prometheus
}
