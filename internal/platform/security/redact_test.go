package security_test

import (
	"testing"

	"cli-assistant/internal/platform/security"
)

func TestRedactToken(t *testing.T) {
	if got := security.RedactToken("short"); got != "***" {
		t.Fatalf("short token = %q", got)
	}
	if got := security.RedactToken("abcdefghij"); got != "abcd...ghij" {
		t.Fatalf("long token = %q", got)
	}
}

func TestRedactURL(t *testing.T) {
	got := security.RedactURL("https://user:secret@argocd.example.com")
	if got != "***@argocd.example.com" {
		t.Fatalf("url = %q", got)
	}
}
