package argocd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
)

type syncRequest struct {
	Name   string `json:"name"`
	DryRun bool   `json:"dryRun,omitempty"`
	Prune  bool   `json:"prune,omitempty"`
	Force  bool   `json:"force,omitempty"`
}

func (c *Client) SyncApplication(
	ctx context.Context,
	_ domain.Scope,
	name string,
	opts deploy.SyncOptions,
) (deploy.SyncResult, error) {
	if name == "" {
		return deploy.SyncResult{}, fmt.Errorf("%w: application name is required", domain.ErrInvalidInput)
	}

	reqBody, err := json.Marshal(syncRequest{
		Name:   name,
		DryRun: opts.DryRun,
		Prune:  opts.Prune,
		Force:  opts.Force,
	})
	if err != nil {
		return deploy.SyncResult{}, err
	}

	path := fmt.Sprintf("/api/v1/applications/%s/sync", url.PathEscape(name))
	if err := c.postJSON(ctx, path, reqBody); err != nil {
		return deploy.SyncResult{}, fmt.Errorf("sync application: %w", err)
	}

	msg := "sync initiated"
	if opts.DryRun {
		msg = "dry-run sync completed"
	}
	return deploy.SyncResult{
		Application: name,
		DryRun:      opts.DryRun,
		Initiated:   true,
		Message:     msg,
	}, nil
}

func (c *Client) postJSON(ctx context.Context, path string, reqBody []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode == http.StatusNotFound {
		return domain.ErrNotFound
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: argocd POST %s: %s", domain.ErrUnavailable, path, strings.TrimSpace(string(respBody)))
	}
	return nil
}
