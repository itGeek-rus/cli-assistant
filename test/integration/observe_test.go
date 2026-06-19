package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestObserveQuery_Integration(t *testing.T) {
	if os.Getenv("INTEGRATION") == "" {
		t.Skip("set INTEGRATION=1 and run docker-compose-integration.yml")
	}

	bin := resolveAssistantBin(t)

	env := integrationEnv(t)
	cmd := exec.Command(bin, "observe", "query", "up")
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("query failed: %v\n%s", err, out)
	}
}

func integrationEnv(t *testing.T) []string {
	t.Helper()

	cfgPath := filepath.Join(t.TempDir(), "config.yaml")
	cfg := `log_level: error
output: human
profile: integration

observability:
  provider: prometheus
  logs_provider: loki
  alerts_provider: alertmanager

profiles:
  integration:
    metrics_url: http://localhost:19090
    logs_url: http://localhost:13100
    alerts_url: http://localhost:19093
`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	return []string{
		"CLI_ASSISTANT_CONFIG=" + cfgPath,
		"CLI_ASSISTANT_PROFILE=integration",
	}
}

func resolveAssistantBin(t *testing.T) string {
	t.Helper()

	bin := os.Getenv("ASSISTANT_BIN")
	root := repoRoot(t)

	if bin == "" {
		return filepath.Join(root, "bin", "assistant")
	}
	if filepath.IsAbs(bin) {
		return bin
	}
	return filepath.Join(root, bin)
}

func repoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}
