package usecase

import (
	"context"
	"log/slog"
)

type Observability struct {
	log *slog.Logger
}

func NewObservability(log *slog.Logger) *Observability {
	return &Observability{log: log}
}

// TODO: Временная загрушка
func (u *Observability) Health(ctx context.Context) error {
	u.log.InfoContext(ctx, "observability health (not implemented yet)")
	return nil
}
