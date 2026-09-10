package observability

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

const Redacted = "[REDACTED]"

var sensitiveKeyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(^|[_.-])(authorization|access[_-]?token|refresh[_-]?token|id[_-]?token|client[_-]?secret|secret|password|passwd|pwd|cookie|credential|credentials|dsn|connection[_-]?string)([_.-]|$)`),
}

var sensitiveValuePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+/=-]+\b`),
	regexp.MustCompile(`(?i)\bBasic\s+[A-Za-z0-9+/=]+\b`),
	regexp.MustCompile(`(?i)(client[_-]?secret|access[_-]?token|refresh[_-]?token|password|passwd|pwd)\s*[:=]\s*[^\s,;]+`),
	regexp.MustCompile(`(?i)(postgres(?:ql)?|mysql|redis)://[^\s"']+`),
	regexp.MustCompile(`(?i)cookie\s*[:=]\s*[^\n]+`),
}

func IsSensitiveKey(key string) bool {
	key = strings.TrimSpace(key)
	if key == "" {
		return false
	}
	for _, pattern := range sensitiveKeyPatterns {
		if pattern.MatchString(key) {
			return true
		}
	}
	return false
}

func SanitizeString(key, value string) string {
	if IsSensitiveKey(key) {
		return Redacted
	}
	result := value
	for _, pattern := range sensitiveValuePatterns {
		result = pattern.ReplaceAllStringFunc(result, func(string) string { return Redacted })
	}
	return result
}

func SanitizeValue(key string, value any) any {
	if IsSensitiveKey(key) {
		return Redacted
	}
	switch typed := value.(type) {
	case string:
		return SanitizeString(key, typed)
	case error:
		if typed == nil {
			return nil
		}
		return errors.New(SanitizeString(key, typed.Error()))
	case fmt.Stringer:
		return SanitizeString(key, typed.String())
	default:
		return value
	}
}

type SanitizingHandler struct {
	next  slog.Handler
	group string
}

func NewSanitizingHandler(next slog.Handler) slog.Handler {
	if next == nil {
		return nil
	}
	return &SanitizingHandler{next: next}
}

func (h *SanitizingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *SanitizingHandler) Handle(ctx context.Context, record slog.Record) error {
	clean := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		clean.AddAttrs(sanitizeAttr(attr, h.group))
		return true
	})
	return h.next.Handle(ctx, clean)
}

func (h *SanitizingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cleaned := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		cleaned = append(cleaned, sanitizeAttr(attr, h.group))
	}
	return &SanitizingHandler{next: h.next.WithAttrs(cleaned), group: h.group}
}

func (h *SanitizingHandler) WithGroup(name string) slog.Handler {
	return &SanitizingHandler{next: h.next.WithGroup(name), group: joinGroup(h.group, name)}
}

func sanitizeAttr(attr slog.Attr, group string) slog.Attr {
	if attr.Equal(slog.Attr{}) {
		return attr
	}
	key := joinGroup(group, attr.Key)
	if attr.Value.Kind() == slog.KindGroup {
		children := attr.Value.Group()
		cleaned := make([]slog.Attr, 0, len(children))
		for _, child := range children {
			cleaned = append(cleaned, sanitizeAttr(child, key))
		}
		return slog.Attr{Key: attr.Key, Value: slog.GroupValue(cleaned...)}
	}
	return slog.Any(attr.Key, SanitizeValue(key, attr.Value.Any()))
}

func joinGroup(prefix, name string) string {
	if prefix == "" {
		return name
	}
	if name == "" {
		return prefix
	}
	return prefix + "." + name
}
