package loki_test

import (
	"cli-assistant/internal/adapter/loki"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestQueryLogs_Range(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query_range" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status":"success",
			"data":{"result":[{
				"stream":{"app":"demo"},
				"values":[["1710000000000000000","hello world"]]
			}]}
		}`))
	}))
	defer srv.Close()
	c, err := loki.New(loki.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	res, err := c.QueryLog(context.Background(), domain.Scope{}, observe.LogsRequest{
		Query: `{app="demo"}`,
		Since: time.Hour,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("QueryLogs: %v", err)
	}
	if len(res.Entries) != 1 || res.Entries[0].Line != "hello world" {
		t.Fatalf("unexpected entries: %+v", res.Entries)
	}
}
