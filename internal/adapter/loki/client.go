package loki

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
		return nil, fmt.Errorf("loki: base URL is required")
	}
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if opts.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec: G402 -- dev only (logs_insecure)
	}
	return &Client{
		baseURL:    strings.TrimRight(opts.BaseURL, "/"),
		httpClient: &http.Client{Timeout: timeout, Transport: transport},
	}, nil
}

func NewFromScope(scope domain.Scope, insecure bool) (*Client, error) {
	return New(Options{BaseURL: scope.LogsURL, Insecure: insecure})
}

func (c *Client) QueryLog(
	ctx context.Context,
	_ domain.Scope,
	req observe.LogsRequest,
) (observe.LogResult, error) {
	if req.Query == "" {
		return observe.LogResult{}, fmt.Errorf("%w: log query is required", domain.ErrInvalidInput)
	}
	since := req.Since
	if since == 0 {
		since = time.Hour
	}
	limit := req.Limit
	if limit == 0 {
		limit = 100
	}

	q := url.Values{}
	q.Set("query", req.Query)
	q.Set("limit", strconv.Itoa(limit))
	q.Set("start", strconv.FormatInt(time.Now().Add(-since).UnixNano(), 10))
	q.Set("direction", "backward")

	var resp queryRangeResponse
	if err := c.getJSON(ctx, "/api/v1/query_range?"+q.Encode(), &resp); err != nil {
		return observe.LogResult{}, fmt.Errorf("query logs: %w", err)
	}
	if resp.Status != "success" {
		return observe.LogResult{}, fmt.Errorf("%w: loki: %s", domain.ErrUnavailable, resp.Error)
	}

	entries := parseLogStreams(resp.Data.Result)
	return observe.LogResult{Query: req.Query, Entries: entries}, nil
}

type queryRangeResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		Result []logStream `json:"result"`
	} `json:"data"`
}

type logStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func parseLogStreams(streams []logStream) []observe.LogEntry {
	out := make([]observe.LogEntry, 0)
	for _, s := range streams {
		for _, v := range s.Values {
			if len(v) < 2 {
				continue
			}
			ts, err := strconv.ParseInt(v[0], 10, 64)
			if err != nil {
				continue
			}
			out = append(out, observe.LogEntry{
				Timestamp: time.Unix(0, ts).UTC(),
				Line:      v[1],
				Labels:    s.Stream,
			})
		}
	}
	return out
}

func (c *Client) getJSON(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: loki GET %s: %s", domain.ErrUnavailable, path, strings.TrimSpace(string(body)))
	}
	return json.Unmarshal(body, dest)
}
