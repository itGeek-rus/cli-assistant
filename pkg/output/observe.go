package output

import (
	"cli-assistant/internal/domain/observe"
	"fmt"
	"time"
)

type StackHealthView struct {
	Overall    string                `json:"overall"`
	CheckedAt  string                `json:"checked_at"`
	Components []ComponentHealthView `json:"components"`
}

type ComponentHealthView struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	URL     string `json:"url,omitempty"`
}

func (p *Printer) PrintStackHealth(h observe.StackHealth) error {
	view := StackHealthView{
		Overall:   string(h.Overall),
		CheckedAt: h.CheckedAt.UTC().Format(time.RFC3339),
	}
	for _, c := range h.Components {
		view.Components = append(view.Components, ComponentHealthView{
			Name:    c.Name,
			Status:  string(c.Status),
			Message: c.Message,
			URL:     c.URL,
		})
	}
	if p.format == FormatJSON {
		return p.Print(view)
	}
	fmt.Fprintf(p.w, "OVERALL: %s (checked %s)\n", view.Overall, view.CheckedAt)
	for _, c := range h.Components {
		line := fmt.Sprintf(" %s: %s", c.Name, c.Status)
		if c.Message != "" {
			line += " - " + c.Message
		}
		fmt.Fprintln(p.w, line)
	}
	return nil
}

type QuerySampleView struct {
	Labels map[string]string `json:"labels"`
	Value  float64           `json:"value"`
	Time   string            `json:"time"`
}

type QueryResultView struct {
	Expr     string            `json:"expr"`
	Samples  []QuerySampleView `json:"samples"`
	Warnings []string          `json:"warnings,omitempty"`
}

func (p *Printer) PrintQueryResult(r observe.QueryResult) error {
	view := QueryResultView{Expr: r.Expr, Warnings: r.Warnings}
	for _, s := range r.Samples {
		view.Samples = append(view.Samples, QuerySampleView{
			Labels: s.Labels,
			Value:  s.Value,
			Time:   s.Time.UTC().Format(time.RFC3339),
		})
	}
	if p.format == FormatJSON {
		return p.Print(view)
	}
	for _, s := range view.Samples {
		fmt.Fprintf(p.w, "%s %v\n", s.Time, s.Labels)
		fmt.Fprintf(p.w, " => %.4f\n", s.Value)
	}
	for _, w := range r.Warnings {
		fmt.Fprintf(p.w, "warning: %s\n", w)
	}
	return nil
}

type AlertView struct {
	Name     string            `json:"name"`
	Severity string            `json:"severity"`
	State    string            `json:"state"`
	Summary  string            `json:"summary"`
	StartsAt string            `json:"starts_at,omitempty"`
	Labels   map[string]string `json:"labels,omitempty"`
}

func AlertViews(alerts []observe.Alert) []AlertView {
	out := make([]AlertView, 0, len(alerts))
	for _, a := range alerts {
		v := AlertView{
			Name:     a.Name,
			Severity: string(a.Severity),
			State:    string(a.State),
			Summary:  a.Summary,
			Labels:   a.Labels,
		}
		if !a.StartAt.IsZero() {
			v.StartsAt = a.StartAt.UTC().Format(time.RFC3339)
		}
		out = append(out, v)
	}
	return out
}

func (p *Printer) PrintAlerts(alerts []observe.Alert) error {
	views := AlertViews(alerts)
	if p.format == FormatJSON {
		return p.Print(views)
	}
	if len(views) == 0 {
		fmt.Fprintln(p.w, "no alerts")
		return nil
	}
	headers := []string{"NAME", "SEVERITY", "STATE", "SUMMARY"}
	rows := make([][]string, 0, len(views))
	for _, v := range views {
		rows = append(rows, []string{v.Name, v.Severity, v.State, v.Summary})
	}
	return p.PrintTable(headers, rows)
}

func (p *Printer) PrintLogs(r observe.LogResult) error {
	if p.format == FormatJSON {
		return p.Print(r)
	}
	if len(r.Entries) == 0 {
		fmt.Fprintln(p.w, "no log entries")
		return nil
	}
	for _, e := range r.Entries {
		fmt.Fprintf(p.w, "%s %s\n", e.Timestamp.UTC().Format(time.RFC3339), e.Line)
	}
	return nil
}

type InspectView struct {
	AppName      string         `json:"app_name"`
	Namespace    string         `json:"namespace"`
	SyncStatus   string         `json:"sync_status"`
	HealthStatus string         `json:"health_status"`
	MetricExpr   string         `json:"metric_expr"`
	MetricValue  *float64       `json:"metric_value,omitempty"`
	Alerts       []AlertView    `json:"alerts"`
	Logs         []LogEntryView `json:"logs"`
}

type LogEntryView struct {
	Timestamp string            `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels,omitempty"`
}

func (p *Printer) PrintInspect(v InspectView) error {
	if p.format == FormatJSON {
		return p.Print(v)
	}
	fmt.Fprintf(p.w, "APP: %s (%s) sync=%s health=%s\n",
		v.AppName, v.Namespace, v.SyncStatus, v.HealthStatus)
	fmt.Fprintf(p.w, "METRIC: %s\n", v.MetricExpr)
	if v.MetricValue != nil {
		fmt.Fprintf(p.w, "  => %.2f\n", *v.MetricValue)
	} else {
		fmt.Fprintln(p.w, "  => n/a")
	}
	fmt.Fprintf(p.w, "ALERTS (%d):\n", len(v.Alerts))
	for _, a := range v.Alerts {
		fmt.Fprintf(p.w, "  - [%s] %s: %s\n", a.Severity, a.Name, a.Summary)
	}
	fmt.Fprintf(p.w, "LOGS (%d):\n", len(v.Logs))
	for _, e := range v.Logs {
		fmt.Fprintf(p.w, "  %s %s\n", e.Timestamp, e.Line)
	}
	return nil
}
