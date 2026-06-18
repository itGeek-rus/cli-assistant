package prometheus

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type Options struct {
	BaseURL  string
	Insecure bool
	Timeout  time.Duration
}

func New(opts Options) (*Client, error) {
	if opts.BaseURL == "" {
		return nil, fmt.Errorf("prometheus: base URL is required")
	}
	if _, err := url.Parse(opts.BaseURL); err != nil {
		return nil, fmt.Errorf("prometheus: invalid base URL: %w", err)
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if opts.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec: G402
	}
	return &Client{
		baseURL: strings.TrimRight(opts.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

func NewFromScope(scope domain.Scope, insecure bool) (*Client, error) {
	return New(Options{BaseURL: scope.MetricsURL, Insecure: insecure})
}

var _ observe.ObservabilityReader = (*Client)(nil)

func (c *Client) CheckStackHealth(ctx context.Context, scope domain.Scope) (observe.StackHealth, error) {
	components := []observe.ComponentHealth{
		c.checkEndpoint(ctx, "prometheus", scope.MetricsURL+"/-/healthy"),
	}
	if scope.LogsURL != "" {
		components = append(components, c.checkEndpoint(ctx, "loki", scope.LogsURL+"/ready"))
	}

	overall := observe.HealthHealthy
	for _, comp := range components {
		switch comp.Status {
		case observe.HealthUnavailable:
			overall = observe.HealthUnavailable
		case observe.HealthDegraded, observe.HealthUnknown:
			if overall == observe.HealthHealthy {
				overall = observe.HealthDegraded
			}
		}
	}
	return observe.StackHealth{
		Overall:    overall,
		Components: components,
		CheckedAt:  time.Now().UTC(),
	}, nil
}

func (c *Client) checkEndpoint(ctx context.Context, name, endpoint string) observe.ComponentHealth {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return observe.ComponentHealth{Name: name, Status: observe.HealthUnavailable, Message: err.Error(), URL: endpoint}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return observe.ComponentHealth{Name: name, Status: observe.HealthUnavailable, Message: err.Error(), URL: endpoint}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return observe.ComponentHealth{Name: name, Status: observe.HealthHealthy, URL: endpoint}
	}
	return observe.ComponentHealth{
		Name:    name,
		Status:  observe.HealthDegraded,
		Message: fmt.Sprintf("HTTP %d", resp.StatusCode),
		URL:     endpoint,
	}
}

func (c *Client) QueryLog(
	_ context.Context,
	_ domain.Scope,
	_ observe.LogsRequest,
) (observe.LogResult, error) {
	return observe.LogResult{}, fmt.Errorf("%w: logs not supported by prometheus adapter", domain.ErrUnavailable)
}

func (c *Client) ListAlerts(
	ctx context.Context,
	scope domain.Scope,
	filter observe.AlertFilter,
) ([]observe.Alert, error) {
	_ = ctx
	_ = scope
	_ = filter
	return nil, fmt.Errorf("%w: alerts not supported by prometheus adapter; use alerts_provider: alertmanager", domain.ErrUnavailable)
}

func (c *Client) QueryMetrics(
	ctx context.Context,
	scope domain.Scope,
	req observe.QueryRequest,
) (observe.QueryResult, error) {
	_ = scope
	if req.Expr == "" {
		return observe.QueryResult{}, fmt.Errorf("%w: query expression is required", domain.ErrInvalidInput)
	}

	q := url.Values{"query": {req.Expr}}
	if !req.Time.IsZero() {
		q.Set("time", strconv.FormatFloat(float64(req.Time.Unix()), 'f', 0, 64))
	}
	var resp queryAPIResponse
	if err := c.getJSON(ctx, "/api/v1/query?"+q.Encode(), &resp); err != nil {
		return observe.QueryResult{}, fmt.Errorf("query metrics: %w", err)
	}
	if resp.Status != "success" {
		return observe.QueryResult{}, fmt.Errorf("%w: prometheus: %s", domain.ErrUnavailable, resp.Error)
	}

	samples, err := parseQueryData(resp.Data)
	if err != nil {
		return observe.QueryResult{}, err
	}

	return observe.QueryResult{
		Expr:     req.Expr,
		Samples:  samples,
		Warnings: resp.Warnings,
	}, nil
}

type queryAPIResponse struct {
	Status   string   `json:"status"`
	Error    string   `json:"error"`
	Warnings []string `json:"warnings"`
	Data     struct {
		ResultType string          `json:"resultType"`
		Result     json.RawMessage `json:"result"`
	} `json:"data"`
}

func parseQueryData(data struct {
	ResultType string          `json:"resultType"`
	Result     json.RawMessage `json:"result"`
}) ([]observe.QuerySample, error) {
	if data.ResultType != "vector" {
		return nil, fmt.Errorf("prometheus: unsupported result type %q", data.ResultType)
	}

	var vector []struct {
		Metric map[string]string `json:"metric"`
		Value  []any             `json:"value"`
	}
	if err := json.Unmarshal(data.Result, &vector); err != nil {
		return nil, fmt.Errorf("prometheus: decode vector: %w", err)
	}

	out := make([]observe.QuerySample, 0, len(vector))
	for _, item := range vector {
		if len(item.Value) < 2 {
			continue
		}
		ts, err := parsePromTimestamp(item.Value[0])
		if err != nil {
			continue
		}
		val, err := parsePromValue(item.Value[1])
		if err != nil {
			continue
		}
		out = append(out, observe.QuerySample{
			Labels: item.Metric,
			Value:  val,
			Time:   ts,
		})
	}
	return out, nil
}

func parsePromTimestamp(v any) (time.Time, error) {
	switch t := v.(type) {
	case float64:
		return time.Unix(int64(t), 0).UTC(), nil
	case string:
		f, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(int64(f), 0).UTC(), nil
	default:
		return time.Time{}, fmt.Errorf("unexpected timestamp type %T", v)
	}
}

func parsePromValue(v any) (float64, error) {
	switch val := v.(type) {
	case string:
		return strconv.ParseFloat(val, 64)
	case float64:
		return val, nil
	default:
		return 0, fmt.Errorf("unexpected value type %T", v)
	}
}

func (c *Client) getJSON(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrNotFound
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: prometheus GET %s: %s", domain.ErrUnavailable, path, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("prometheus: decode response: %w", err)
	}
	return nil
}
