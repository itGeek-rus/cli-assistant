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
