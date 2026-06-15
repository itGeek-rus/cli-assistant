package prometheus_test

import (
	"cli-assistant/internal/adapter/prometheus"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryMetrics_Vector(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status":"success",
			"data":{
				"resultType":"vector",
				"result":[{
					"metric":{"__name__":"up","job":"demo"},
					"value":[1710000000,"1"]
				}]
			}
		}`))
	}))
	defer srv.Close()

	c, err := prometheus.New(prometheus.Options{BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	res, err := c.QueryMetrics(context.Background(), domain.Scope{}, observe.QueryRequest{Expr: "up"})
	if err != nil {
		t.Fatalf("QueryMetrics: %v", err)
	}
	if len(res.Samples) != 1 || res.Samples[0].Value != 1 {
		t.Fatalf("unexpected samples %+v", res.Samples)
	}
}
