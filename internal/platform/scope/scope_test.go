package scope_test

import (
	"cli-assistant/internal/config"
	"cli-assistant/internal/platform/scope"
	"testing"
)

func TestFromConfig_DefaultProfile(t *testing.T) {
	cfg := config.Default()
	s, err := scope.FromConfig(cfg)
	if err != nil {
		t.Fatalf("FromConfig: %v", err)
	}
	if s.ProfileName != "default" {
		t.Fatalf("profile = %q, want default", s.ProfileName)
	}
}

func TestFromConfig_UnknownProfile(t *testing.T) {
	cfg := config.Default()
	cfg.Profile = "missing"
	if _, err := scope.FromConfig(cfg); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}
