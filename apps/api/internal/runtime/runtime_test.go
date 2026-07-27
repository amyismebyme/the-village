package runtime

import (
	"strings"
	"testing"
	"time"
)

func TestBuildVersionNotEmpty(t *testing.T) {

	if BuildVersion == "" {
		t.Fatal("BuildVersion should not be empty")
	}
}

func TestGitCommitNotEmpty(t *testing.T) {

	if GitCommit == "" {
		t.Fatal("GitCommit should not be empty")
	}
}

func TestStartedAtIsSet(t *testing.T) {

	if StartedAt.IsZero() {
		t.Fatal("StartedAt should be initialized")
	}
}

func TestBuildTimeIsSet(t *testing.T) {

	if BuildTime.IsZero() {
		t.Fatal("BuildTime should be initialized")
	}
}

func TestGoVersion(t *testing.T) {

	version := GoVersion()

	if !strings.HasPrefix(version, "go") {
		t.Fatalf("unexpected Go version: %s", version)
	}
}

func TestUptimeIncreases(t *testing.T) {

	first := Uptime()

	time.Sleep(20 * time.Millisecond)

	second := Uptime()

	if second <= first {
		t.Fatalf(
			"expected uptime to increase (%v <= %v)",
			second,
			first,
		)
	}
}
