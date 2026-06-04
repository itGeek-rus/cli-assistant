package output

import (
	"cli-assistant/internal/domain/deploy"
	"fmt"
	"time"
)

type ApplicationView struct {
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Project        string `json:"project"`
	Sync           string `json:"sync"`
	Health         string `json:"health"`
	Revision       string `json:"revision"`
	TargetRevision string `json:"target_revision"`
	LastSynced     string `json:"last_synced,omitempty"`
}

func ApplicationViews(apps []deploy.Application) []ApplicationView {
	out := make([]ApplicationView, 0, len(apps))
	for _, a := range apps {
		v := ApplicationView{
			Name:           a.Name,
			Namespace:      a.Namespace,
			Project:        a.Project,
			Sync:           string(a.SyncStatus),
			Health:         string(a.HealthStatus),
			Revision:       a.Revision,
			TargetRevision: a.TargetRevision,
		}
		if a.LastSyncedAt != nil {
			v.LastSynced = a.LastSyncedAt.UTC().Format(time.RFC3339)
		}
		out = append(out, v)
	}
	return out
}

func (p *Printer) PrintApplications(apps []deploy.Application) error {
	views := ApplicationViews(apps)
	if p.format == FormatJSON {
		return p.Print(views)
	}
	headers := []string{"NAME", "NAMESPACE", "SYNC", "HEALTH", "REVISION", "TARGET"}
	rows := make([][]string, 0, len(views))
	for _, v := range views {
		rows = append(rows, []string{
			v.Name, v.Namespace, v.Sync, v.Health, v.Revision, v.TargetRevision,
		})
	}
	return p.PrintTable(headers, rows)
}

func (p *Printer) PrintApplication(app deploy.Application) error {
	return p.PrintApplications([]deploy.Application{app})
}

func (p *Printer) PrintDiff(diff deploy.DiffResult) error {
	if p.format == FormatJSON {
		return p.Print(diff)
	}
	if diff.Raw != "" {
		_, err := fmt.Fprintln(p.w, diff.Raw)
		return err
	}
	for _, ch := range diff.Changes {
		fmt.Fprintf(p.w, "%s/%s: %s\n", ch.Kind, ch.Name, ch.Summary)
	}
	return nil
}

func (p *Printer) PrintSync(res deploy.SyncResult) error {
	if p.format == FormatJSON {
		return p.Print(res)
	}
	fmt.Fprintf(p.w, "%s: %s\n", res.Application, res.Message)
	return nil
}
