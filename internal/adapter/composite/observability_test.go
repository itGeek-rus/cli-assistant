package composite_test

import (
	"context"
	"testing"

	"cli-assistant/internal/adapter/composite"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
)

type fakeCore struct {
	health observe.StackHealth
	alerts []observe.Alert
	metric observe.QueryResult
}

func (f fakeCore) CheckStackHealth(context.Context, domain.Scope) (observe.StackHealth, error) {
	return f.health, nil
}

func (f fakeCore) ListAlerts(context.Context, domain.Scope, observe.AlertFilter) ([]observe.Alert, error) {
	return f.alerts, nil
}

func (f fakeCore) QueryMetrics(context.Context, domain.Scope, observe.QueryRequest) (observe.QueryResult, error) {
	return f.metric, nil
}

func (f fakeCore) QueryLog(context.Context, domain.Scope, observe.LogsRequest) (observe.LogResult, error) {
	return observe.LogResult{}, nil
}

type fakeAlerts struct {
	alerts []observe.Alert
}

func (f fakeAlerts) ListAlerts(context.Context, domain.Scope, observe.AlertFilter) ([]observe.Alert, error) {
	return f.alerts, nil
}

type fakeLogs struct {
	result observe.LogResult
}

func (f fakeLogs) QueryLog(context.Context, domain.Scope, observe.LogsRequest) (observe.LogResult, error) {
	return f.result, nil
}

func TestReader_ListAlerts_DelegatesToAlertsAdapter(t *testing.T) {
	core := fakeCore{alerts: []observe.Alert{{Name: "from-core"}}}
	alerts := fakeAlerts{alerts: []observe.Alert{{Name: "from-alertmanager"}}}
	r := composite.New(core, alerts, fakeLogs{})

	got, err := r.ListAlerts(context.Background(), domain.Scope{}, observe.AlertFilter{})
	if err != nil {
		t.Fatalf("ListAlerts: %v", err)
	}
	if len(got) != 1 || got[0].Name != "from-alertmanager" {
		t.Fatalf("alerts = %+v, want alertmanager adapter", got)
	}
}

func TestReader_QueryLog_DelegatesToLogsAdapter(t *testing.T) {
	want := observe.LogResult{Query: `{app="demo"}`, Entries: []observe.LogEntry{{Line: "ok"}}}
	r := composite.New(fakeCore{}, fakeAlerts{}, fakeLogs{result: want})

	got, err := r.QueryLog(context.Background(), domain.Scope{}, observe.LogsRequest{Query: `{app="demo"}`})
	if err != nil {
		t.Fatalf("QueryLog: %v", err)
	}
	if got.Query != want.Query || len(got.Entries) != 1 {
		t.Fatalf("logs = %+v, want %+v", got, want)
	}
}

func TestReader_CheckStackHealth_DelegatesToCore(t *testing.T) {
	want := observe.StackHealth{Overall: observe.HealthHealthy}
	r := composite.New(fakeCore{health: want}, fakeAlerts{}, fakeLogs{})

	got, err := r.CheckStackHealth(context.Background(), domain.Scope{})
	if err != nil {
		t.Fatalf("CheckStackHealth: %v", err)
	}
	if got.Overall != want.Overall {
		t.Fatalf("health = %s, want %s", got.Overall, want.Overall)
	}
}
