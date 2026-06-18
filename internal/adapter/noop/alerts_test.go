package noop_test

import (
	"context"
	"testing"

	"cli-assistant/internal/adapter/noop"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
)

func TestAlerts_ListAlerts_Empty(t *testing.T) {
	alerts, err := noop.NewAlerts().ListAlerts(context.Background(), domain.Scope{}, observe.AlertFilter{})
	if err != nil {
		t.Fatalf("ListAlerts: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("alerts = %+v, want empty", alerts)
	}
}
