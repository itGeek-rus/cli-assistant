package domain

// Scope - контекст выполнения: профиль, кластер, опциональные фильтры.
type Scope struct {
	ProfileName string
	KubeContext string
	GitOpsURL   string
	MetricsURL  string
	LogsURL     string
	AlertsURL   string
	Namespace   string
}
