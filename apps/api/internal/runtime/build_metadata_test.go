package runtime

import (
	"os"
	"testing"
	"time"
)

func TestInjectedBuildMetadata(t *testing.T) {
	t.Helper()

	expectedVersion := os.Getenv(
		"EXPECTED_BUILD_VERSION",
	)
	expectedCommit := os.Getenv(
		"EXPECTED_GIT_COMMIT",
	)
	expectedTimestamp := os.Getenv(
		"EXPECTED_BUILD_TIMESTAMP",
	)
	expectedEnvironment := os.Getenv(
		"EXPECTED_ENVIRONMENT",
	)

	if expectedVersion == "" &&
		expectedCommit == "" &&
		expectedTimestamp == "" &&
		expectedEnvironment == "" {
		t.Skip(
			"build metadata expectations are not configured",
		)
	}

	if BuildVersion != expectedVersion {
		t.Fatalf(
			"expected BuildVersion=%q, got %q",
			expectedVersion,
			BuildVersion,
		)
	}

	if GitCommit != expectedCommit {
		t.Fatalf(
			"expected GitCommit=%q, got %q",
			expectedCommit,
			GitCommit,
		)
	}

	if BuildTimestamp != expectedTimestamp {
		t.Fatalf(
			"expected BuildTimestamp=%q, got %q",
			expectedTimestamp,
			BuildTimestamp,
		)
	}

	if Environment != expectedEnvironment {
		t.Fatalf(
			"expected Environment=%q, got %q",
			expectedEnvironment,
			Environment,
		)
	}

	if _, err := time.Parse(
		time.RFC3339,
		BuildTimestamp,
	); err != nil {
		t.Fatalf(
			"BuildTimestamp is not RFC3339: %v",
			err,
		)
	}
}
