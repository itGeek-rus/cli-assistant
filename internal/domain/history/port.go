package history

import (
	"context"
	"time"
)

type CommandRun struct {
	ID        int64
	Command   string
	Profile   string
	Status    string
	Output    string
	ErrorText string
	CreatedAt time.Time
}

type InspectSnapshot struct {
	ID          int64
	AppName     string
	Profile     string
	PayloadJSON []byte
	CreatedAt   time.Time
}

type Store interface {
	SaveCommandRun(ctx context.Context, run CommandRun) error
	ListCommandRuns(ctx context.Context, limit int) ([]CommandRun, error)
	SaveInspectSnapshot(ctx context.Context, snap InspectSnapshot) error
	LatestSnapshot(ctx context.Context, appName string) (InspectSnapshot, error)
}
