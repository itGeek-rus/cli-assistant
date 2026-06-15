package usecase

import (
	"cli-assistant/internal/config"
	"cli-assistant/internal/domain"
	"cli-assistant/internal/domain/deploy"
	"cli-assistant/internal/domain/observe"
	"cli-assistant/internal/platform/scope"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Inspector struct {
	log    *slog.Logger
	cfg    config.Config
	gitops deploy.GitOpsReader
	obs    observe.ObservabilityReader
}

func NewInspector(log *slog.Logger, cfg config.Config, gitops deploy.GitOpsReader, obs observe.ObservabilityReader) *Inspector {
	return &Inspector{log: log, cfg: cfg, gitops: gitops, obs: obs}
}

type InspectResult struct {
	Application deploy.Application
	Alerts      []observe.Alert
	MetricExpr  string
	MetricUp    *observe.QuerySample
	Logs        observe.LogResult
}

func (i *Inspector) Inspect(ctx context.Context, appName string) (InspectResult, error) {
	if appName == "" {
		return InspectResult{}, fmt.Errorf("%w: application name is required", domain.ErrInvalidInput)
	}

	s, err := scope.FromConfig(i.cfg)
	if err != nil {
		return InspectResult{}, err
	}

	app, err := i.gitops.GetApplication(ctx, s, appName)
	if err != nil {
		return InspectResult{}, fmt.Errorf("get application: %w", err)
	}

	alerts, err := i.obs.ListAlerts(ctx, s, observe.AlertFilter{
		Labels: map[string]string{"app": appName},
	})
	if err != nil {
		return InspectResult{}, fmt.Errorf("list alerts: %w", err)
	}

	metricExpr := fmt.Sprintf(`up{app="%s"}`, appName)
	metrics, err := i.obs.QueryMetrics(ctx, s, observe.QueryRequest{Expr: metricExpr})
	if err != nil {
		return InspectResult{}, fmt.Errorf("query metrics: %w", err)
	}

	var metricsUp *observe.QuerySample
	if len(metrics.Samples) > 0 {
		metricsUp = &metrics.Samples[0]
	}

	logQL := fmt.Sprintf(`{app="%s"}`, appName)
	logs, err := i.obs.QueryLog(ctx, s, observe.LogsRequest{
		Query: logQL,
		Since: 30 * time.Minute,
		Limit: 20,
	})
	if err != nil {
		i.log.DebugContext(ctx, "logs unavailable for inspect", "app", appName, "err", err)
		// логи не блокируют inspect
	}

	return InspectResult{
		Application: app,
		Alerts:      alerts,
		MetricExpr:  metricExpr,
		MetricUp:    metricsUp,
		Logs:        logs,
	}, nil
}
