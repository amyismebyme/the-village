package observability

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var forbiddenTelemetryKey = regexp.MustCompile(`(?i)\b(access[_-]?token|refresh[_-]?token|client[_-]?secret|authorization|password|cookie|credential|dsn|connection[_-]?string)\b`)

func TestTelemetryDefinitionsDoNotUseSecretBearingKeys(t *testing.T) {
	root := moduleRoot(t)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)NewCounterVec\(`),
		regexp.MustCompile(`(?i)NewGaugeVec\(`),
		regexp.MustCompile(`(?i)NewHistogramVec\(`),
		regexp.MustCompile(`(?i)attribute\.(String|Int|Bool|Float64)\(`),
		regexp.MustCompile(`(?i)\.WithAttributes\(`),
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "vendor" || info.Name() == "api.exe" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			if err := file.Close(); err != nil {
				t.Errorf("failed to close %s: %v", path, err)
			}
		}()
		scanner := bufio.NewScanner(file)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()
			matched := false
			for _, pattern := range patterns {
				if pattern.MatchString(line) {
					matched = true
					break
				}
			}
			if matched && forbiddenTelemetryKey.MatchString(line) {
				t.Errorf("forbidden telemetry key at %s:%d: %s", path, lineNo, strings.TrimSpace(line))
			}
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/observability -> internal -> api
	return filepath.Clean(filepath.Join(filepath.Dir(current), "..", ".."))
}
