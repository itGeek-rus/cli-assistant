package argocd

import (
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
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
	token      string
	httpClient *http.Client
}

type Options struct {
	BaseURL  string
	Token    string
	Insecure bool
	Timeout  time.Duration
}

func New(opt Options) (*Client, error) {
	if opt.BaseURL == "" {
		return nil, errors.New("argocd: base URL is required")
	}
	if _, err := url.Parse(opt.BaseURL); err != nil {
		return nil, fmt.Errorf("argocd: invalid base URL: %w", err)
	}
	timeout := opt.Timeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if opt.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // nolint:gosec
	}
	return &Client{
		baseURL: strings.TrimRight(opt.BaseURL, "/"),
		token:   opt.Token,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

func NewFromScope(scope domain.Scope, token string, insecure bool) (*Client, error) {
	return New(Options{
		BaseURL:  scope.GitOpsURL,
		Token:    token,
		Insecure: insecure,
	})
}

var _ deploy.GitOpsReader = (*Client)(nil)

func (c *Client) ListApplications(
	ctx context.Context,
	scope domain.Scope,
	filter deploy.ListFilter,
) ([]deploy.Application, error) {
	_ = scope

	q := url.Values{}
	if filter.Namespace != "" {
		q.Set("appNamespace", filter.Namespace)
	}
	path := "/api/v1/applications"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}

	var list applicationList
	if err := c.getJSON(ctx, path, &list); err != nil {
		return nil, err
	}

	apps := make([]deploy.Application, 0, len(list.Items))
	for _, item := range list.Items {
		app := mapApplication(item)
		if filter.NamePrefix != "" && !strings.HasPrefix(app.Name, filter.NamePrefix) {
			continue
		}
		apps = append(apps, app)
	}
	return apps, nil
}

func (c *Client) GetApplication(
	ctx context.Context,
	_ domain.Scope,
	name string,
) (deploy.Application, error) {
	if name == "" {
		return deploy.Application{}, fmt.Errorf("%w: application name is required", domain.ErrInvalidInput)
	}
	path := fmt.Sprintf("/api/v1/applications/%s", url.PathEscape(name))
	var item application
	if err := c.getJSON(ctx, path, &item); err != nil {
		return deploy.Application{}, err
	}
	return mapApplication(item), nil
}

func (c *Client) getJSON(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

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
		return fmt.Errorf("%w: argocd GET %s: %s", domain.ErrUnavailable, path, strings.TrimSpace(string(body)))
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("argocd: decode response: %w", err)
	}
	return nil
}

func mapApplication(item application) deploy.Application {
	app := deploy.Application{
		Name:           item.Metadata.Name,
		Namespace:      item.Metadata.Namespace,
		Project:        item.Spec.Project,
		SyncStatus:     mapSyncStatus(item.Status.Sync.Status),
		HealthStatus:   mapHealthStatus(item.Status.Health.Status),
		Revision:       item.Status.Sync.Revision,
		TargetRevision: item.Spec.Source.TargetRevision,
	}
	if item.Status.OperationState != nil && item.Status.OperationState.FinishedAt != "" {
		if t, err := time.Parse(time.RFC3339, item.Status.OperationState.FinishedAt); err == nil {
			app.LastSyncedAt = &t
		}
	}
	return app
}

func mapSyncStatus(s string) deploy.SyncStatus {
	switch strings.TrimSpace(s) {
	case "Synced":
		return deploy.SyncStatusSynced
	case "OutOfSync":
		return deploy.SyncStatusOutOfSync
	default:
		return deploy.SyncStatusUnknown
	}
}

func mapHealthStatus(s string) deploy.HealthStatus {
	switch strings.TrimSpace(s) {
	case "Healthy":
		return deploy.HealthHealthy
	case "Degraded":
		return deploy.HealthDegraded
	case "Progressing":
		return deploy.HealthProgressing
	case "Missing":
		return deploy.HealthMissing
	default:
		return deploy.HealthUnknown
	}
}
