package usecase

import (
	"context"
	"log/slog"
)

type Deployment struct {
	log *slog.Logger
}

func NewDeployment(log *slog.Logger) *Deployment {
	return &Deployment{log: log}
}

// TODO: Времменная заглушка
func (u *Deployment) Status(ctx context.Context) error {
	u.log.InfoContext(ctx, "deployment status (not deployment yet)")
	return nil
}
