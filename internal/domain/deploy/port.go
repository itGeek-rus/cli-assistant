package deploy

import (
	"cli-assistant/internal/domain"
	"context"
)

type GitOpsReader interface {
	ListApplications(ctx context.Context, scope domain.Scope, filter ListFilter) ([]Application, error)
	GetApplication(ctx context.Context, scope domain.Scope, name string) (Application, error)
}
