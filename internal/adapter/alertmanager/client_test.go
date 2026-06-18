package alertmanager_test

import (
	"cli-assistant/internal/adapter/alertmanager"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListAlerts_Active(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/alerts" {
			http.NotFound(w, r)
			return
		}
		if got := r.URL.Query().Get("filter"); got != `state="active"` {
			t.Errorf("filter query = %q, want state=\"active\"", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{
			"labels":{"alertname":"HighCPU","severity":"warning","app":"demo-app"},
			"annotations":{"summary":"CPU > 80%"},
			"startsAt":"2026-05-26T10:00:00Z",
			"status":{"state":"active"}
		}]`))
	}))
	defer srv.Close()
	c, err := alertmanager.New(alertmanager.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	alerts, err := c.ListAlerts(context.Background(), domain.Scope{}, observe.AlertFilter{
		State: observe.AlertStateFiring,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 || alerts[0].Name != "HighCPU" {
		t.Fatalf("unexpected: %+v", alerts)
	}
}

func TestListAlerts_LabelFilter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"labels":{"alertname":"A","app":"keep"},"status":{"state":"active"}},
			{"labels":{"alertname":"B","app":"drop"},"status":{"state":"active"}}
		]`))
	}))
	defer srv.Close()

	c, err := alertmanager.New(alertmanager.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	alerts, err := c.ListAlerts(context.Background(), domain.Scope{}, observe.AlertFilter{
		Labels: map[string]string{"app": "keep"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 1 || alerts[0].Name != "A" {
		t.Fatalf("unexpected: %+v", alerts)
	}
}
