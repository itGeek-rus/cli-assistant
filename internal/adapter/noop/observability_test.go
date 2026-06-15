package noop_test

import (
	"cli-assistant/internal/adapter/noop"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"testing"
)

func TestQueryMetrics_DemoSample(t *testing.T) {
	o := noop.NewObservability()
	res, err := o.QueryMetrics(context.Background(), domain.Scope{}, observe.QueryRequest{Expr: "up"})
	if err != nil {
		t.Fatalf("QueryMetrics: %v", err)
	}
	if len(res.Samples) != 1 {
		t.Fatalf("samples = %d, want 1", len(res.Samples))
	}
	if res.Samples[0].Value != 1 {
		t.Fatalf("value = %v, want 1", res.Samples[0].Value)
	}
}

func TestQueryLogs_RequiresExpr(t *testing.T) {
	o := noop.NewObservability()
	_, err := o.QueryLog(context.Background(), domain.Scope{}, observe.LogsRequest{})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}
