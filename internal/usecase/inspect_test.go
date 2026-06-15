package usecase_test

import (
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
	"cli-assistant/internal/domain/observe"
	"cli-assistant/internal/usecase"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

type fakeGitOps struct{}
type fakeObs struct{}

func (f fakeGitOps) ListApplications(context.Context, domain.Scope, deploy.ListFilter) ([]deploy.Application, error) {
	return nil, nil
}

func (f fakeGitOps) GetApplication(context.Context, domain.Scope, string) (deploy.Application, error) {
	return deploy.Application{
		Name: "demo-app", Namespace: "default",
		SyncStatus: deploy.SyncStatusSynced, HealthStatus: deploy.HealthHealthy,
	}, nil
}

func (f fakeGitOps) SyncApplication(context.Context, domain.Scope, string, deploy.SyncOptions) (deploy.SyncResult, error) {
	return deploy.SyncResult{}, nil
}

func (f fakeGitOps) DiffApplication(context.Context, domain.Scope, string) (deploy.DiffResult, error) {
	return deploy.DiffResult{}, nil
}

func (f fakeObs) CheckStackHealth(context.Context, domain.Scope) (observe.StackHealth, error) {
	return observe.StackHealth{}, nil
}

func (f fakeObs) ListAlerts(context.Context, domain.Scope, observe.AlertFilter) ([]observe.Alert, error) {
	return []observe.Alert{{Name: "HighCPU", Severity: observe.AlertSeverityWarning, State: observe.AlertStateFiring}}, nil
}

func (f fakeObs) QueryMetrics(context.Context, domain.Scope, observe.QueryRequest) (observe.QueryResult, error) {
	return observe.QueryResult{
		Samples: []observe.QuerySample{{Value: 1, Time: time.Now().UTC()}},
	}, nil
}

func (f fakeObs) QueryLog(context.Context, domain.Scope, observe.LogsRequest) (observe.LogResult, error) {
	return observe.LogResult{Entries: []observe.LogEntry{{Line: "ok"}}}, nil
}

func TestInspector_Inspect(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	ins := usecase.NewInspector(log, config.Default(), fakeGitOps{}, fakeObs{})
	res, err := ins.Inspect(context.Background(), "demo-app")
	if err != nil {
		t.Fatalf("Inspect: %v", err)
	}
	if res.Application.Name != "demo-app" {
		t.Fatalf("app = %q", res.Application.Name)
	}
	if len(res.Alerts) != 1 {
		t.Fatalf("alerts = %d", len(res.Alerts))
	}
	if res.MetricUp == nil || res.MetricUp.Value != 1 {
		t.Fatalf("metric up missing")
	}
}
