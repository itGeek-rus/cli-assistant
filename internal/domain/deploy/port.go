package deploy

import (
	"cli-assistant/internal/domain"
	"context"
)

type GitOpsReader interface {
	ListApplications(ctx context.Context, scope domain.Scope, filter ListFilter) ([]Application, error)
	GetApplication(ctx context.Context, scope domain.Scope, name string) (Application, error)
}

type GitOpsWriter interface {
	SyncApplication(ctx context.Context, scope domain.Scope, name string, opts SyncOptions) (SyncResult, error)
	DiffApplication(ctx context.Context, scope domain.Scope, name string) (DiffResult, error)
}

// Client - для DI, когда один адаптер и читает, и пишет.
type GitOpsClient interface {
	GitOpsReader
	GitOpsWriter
}
