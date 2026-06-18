package version_test

import (
	"strings"
	"testing"

	"cli-assistant/pkg/version"
)

func TestString_ContainsBuildMetadata(t *testing.T) {
	got := version.String()
	for _, part := range []string{version.Version, version.Commit, version.BuildDate} {
		if !strings.Contains(got, part) {
			t.Fatalf("String() = %q, missing %q", got, part)
		}
	}
}
