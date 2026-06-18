package alertmanager

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/observe"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
		return nil, errors.New("alertmanager: base URL is required")
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
		baseURL:    strings.TrimRight(opts.BaseURL, "/"),
		httpClient: &http.Client{Timeout: timeout, Transport: transport},
	}, nil
}

func NewFromScope(scope domain.Scope, insecure bool) (*Client, error) {
	return New(Options{BaseURL: scope.AlertsURL, Insecure: insecure})
}

func (c *Client) ListAlerts(
	ctx context.Context,
	_ domain.Scope,
	filter observe.AlertFilter,
) ([]observe.Alert, error) {
	q := url.Values{}
	q.Set("silenced", "false")
	q.Set("inhibited", "false")
	if filter.State == observe.AlertStateFiring {
		q.Set("filter", `state="active"`)
	}
	if filter.State == observe.AlertStatePending {
		q.Set("filter", `state="suppressed"`)
	}

	var raw []amAlert
	if err := c.getJSON(ctx, "/api/v2/alerts?"+q.Encode(), &raw); err != nil {
		return nil, fmt.Errorf("list alerts: %w", err)
	}

	out := make([]observe.Alert, 0, len(raw))
	for _, a := range raw {
		alert := mapAlert(a)
		if !matchAlertFilter(alert, filter) {
			continue
		}
		out = append(out, alert)
	}
	return out, nil
}

type amAlert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
	Status      struct {
		State string `json:"state"` // active | suppressed | unmapped
	} `json:"status"`
}

func mapAlert(a amAlert) observe.Alert {
	name := a.Labels["alertname"]
	if name == "" {
		name = "unknown"
	}
	return observe.Alert{
		Name:     name,
		Severity: mapSeverity(a.Labels["severity"]),
		State:    mapState(a.Status.State),
		Summary:  firstNonEmpty(a.Annotations["summary"], a.Annotations["description"]),
		StartAt:  a.StartsAt.UTC(),
		Labels:   a.Labels,
	}
}

func mapSeverity(v string) observe.AlertSeverity {
	switch strings.ToLower(v) {
	case "info":
		return observe.AlertSeverityInfo
	case "warning", "warn":
		return observe.AlertSeverityWarning
	case "critical":
		return observe.AlertSeverityCritical
	default:
		return observe.AlertSeverityUnknown
	}
}

func mapState(v string) observe.AlertState {
	switch v {
	case "active":
		return observe.AlertStateFiring
	case "suppressed":
		return observe.AlertStatePending
	default:
		return observe.AlertStateInactive
	}
}

func matchAlertFilter(a observe.Alert, f observe.AlertFilter) bool {
	if f.State != "" && a.State != f.State {
		return false
	}
	for k, want := range f.Labels {
		if a.Labels[k] != want {
			return false
		}
	}
	return true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
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
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: alertmanager GET %s: %s", domain.ErrUnavailable, path, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, dest)
}
