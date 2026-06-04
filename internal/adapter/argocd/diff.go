package argocd

import (
	"context"
	"fmt"

	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
)

func (c *Client) DiffApplication(
	ctx context.Context,
	scope domain.Scope,
	name string,
) (deploy.DiffResult, error) {
	app, err := c.GetApplication(ctx, scope, name)
	if err != nil {
		return deploy.DiffResult{}, err
	}

	result := deploy.DiffResult{
		Application: name,
		OutOfSync:   app.SyncStatus == deploy.SyncStatusOutOfSync,
		Raw: fmt.Sprintf("application %q is %s (revision %s, target %s)",
			app.Name, app.SyncStatus, app.Revision, app.TargetRevision),
	}

	if app.SyncStatus == deploy.SyncStatusOutOfSync {
		result.Changes = []deploy.DiffChange{{
			Kind:    "Application",
			Name:    app.Name,
			Summary: "out of sync with target revision " + app.TargetRevision,
		}}
	}
	return result, nil
}
