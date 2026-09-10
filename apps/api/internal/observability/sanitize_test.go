package observability

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"authorization", "Authorization", "Bearer top-secret-token"},
		{"password", "db_password", "dont-log-this"},
		{"dsn", "endpoint", "postgres://village:pw@postgres:5432/village?sslmode=disable"},
		{"inline secret", "message", "client_secret=abc123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.key, tt.value)
			for _, secret := range []string{"top-secret-token", "dont-log-this", "postgres://", "abc123"} {
				if strings.Contains(got, secret) {
					t.Fatalf("secret leaked: %q", got)
				}
			}
			if !strings.Contains(got, Redacted) {
				t.Fatalf("expected redaction marker in %q", got)
			}
		})
	}
}

func TestSanitizingHandlerPreservesCorrelation(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(NewSanitizingHandler(slog.NewJSONHandler(&buf, nil)))
	logger.LogAttrs(
		context.Background(),
		slog.LevelError,
		"operation failed",
		slog.String("trace_id", "0123456789abcdef0123456789abcdef"),
		slog.String("span_id", "0123456789abcdef"),
		slog.String("authorization", "Bearer secret"),
		slog.Group("database", slog.String("password", "secret")),
		slog.Any("error", errors.New("postgres://village:pw@postgres:5432/village")),
	)
	out := buf.String()
	for _, forbidden := range []string{"Bearer secret", "postgres://village:pw", "\"password\":\"secret\""} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("sensitive value leaked: %s", forbidden)
		}
	}
	for _, required := range []string{"0123456789abcdef0123456789abcdef", "0123456789abcdef"} {
		if !strings.Contains(out, required) {
			t.Fatalf("correlation field missing: %s", required)
		}
	}
}
