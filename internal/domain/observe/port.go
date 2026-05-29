package observe

import (
	"cli-assistant/internal/domain"
	"context"
)

type ObservabilityReader interface {
	CheckStackHealth(ctx context.Context, scope domain.Scope) (StackHealth, error)
	ListAlerts(ctx context.Context, scope domain.Scope, filter AlertFilter) ([]Alert, error)
}
