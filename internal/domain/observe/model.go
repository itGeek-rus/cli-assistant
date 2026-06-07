package observe

import "time"

type HealthStatus string

const (
	HealthUnknown     HealthStatus = "Unknown"
	HealthHealthy     HealthStatus = "Healthy"
	HealthDegraded    HealthStatus = "Degraded"
	HealthUnavailable HealthStatus = "Unavailable"
)

type ComponentHealth struct {
	Name    string
	Status  HealthStatus
	Message string
	URL     string
}

type StackHealth struct {
	Overall    HealthStatus
	Components []ComponentHealth
	CheckedAt  time.Time
}

type AlertSeverity string

const (
	AlertSeverityUnknown  AlertSeverity = "Unknown"
	AlertSeverityInfo     AlertSeverity = "Info"
	AlertSeverityWarning  AlertSeverity = "Warning"
	AlertSeverityCritical AlertSeverity = "Critical"
)

type AlertState string

const (
	AlertStateFiring   AlertState = "Firing"
	AlertStatePending  AlertState = "Pending"
	AlertStateInactive AlertState = "Inactive"
)

type Alert struct {
	Name     string
	Severity AlertSeverity
	State    AlertState
	Summary  string
	StartAt  time.Time
	Labels   map[string]string
}

type AlertFilter struct {
	State  AlertState
	Labels map[string]string
}

type QueryRequest struct {
	Expr string    // PromQL
	Time time.Time // zer0 == now (instant query)
}

type QuerySample struct {
	Labels map[string]string
	Value  float64
	Time   time.Time
}

type QueryResult struct {
	Expr     string
	Samples  []QuerySample
	Warnings []string
}
