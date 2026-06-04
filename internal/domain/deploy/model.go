package deploy

import "time"

type SyncStatus string

const (
	SyncStatusUnknown   SyncStatus = "Unknown"
	SyncStatusSynced    SyncStatus = "Synced"
	SyncStatusOutOfSync SyncStatus = "OutOfSync"
)

type HealthStatus string

const (
	HealthUnknown     HealthStatus = "Unknown"
	HealthHealthy     HealthStatus = "Healthy"
	HealthDegraded    HealthStatus = "Degraded"
	HealthProgressing HealthStatus = "Progressing"
	HealthMissing     HealthStatus = "Missing"
)

type Application struct {
	Name           string
	Namespace      string
	Project        string
	SyncStatus     SyncStatus
	HealthStatus   HealthStatus
	Revision       string
	TargetRevision string
	LastSyncedAt   *time.Time
}

type ListFilter struct {
	Namespace  string
	NamePrefix string
}

type SyncOptions struct {
	DryRun bool
	Prune  bool
	Force  bool
}

type SyncResult struct {
	Application string
	DryRun      bool
	Initiated   bool
	Message     string
}

type DiffChange struct {
	Kind      string // Deployment, Service
	Name      string
	Namespace string
	Summary   string
}

type DiffResult struct {
	Application string
	OutOfSync   bool
	Changes     []DiffChange
	Raw         string // полный diff текст, если API отдал
}
