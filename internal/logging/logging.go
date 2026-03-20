package logging

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	logger *slog.Logger
	once   sync.Once
	mu     sync.RWMutex
)

func Init(level string) {
	level = resolveLevel(level)

	mu.Lock()
	logger = defaultLogger(level)
	mu.Unlock()

	once.Do(func() {})
}

func L() *slog.Logger {
	ensureInitialized()

	mu.RLock()
	l := logger
	mu.RUnlock()

	return l
}

func Info(msg string, args ...any) {
	L().Info(msg, args...)
}

func Debug(msg string, args ...any) {
	L().Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	L().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	L().Error(msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	L().InfoContext(ctx, msg, args...)
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	L().DebugContext(ctx, msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	L().WarnContext(ctx, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	L().ErrorContext(ctx, msg, args...)
}

func InfoAttrs(ctx context.Context, msg string, attrs ...slog.Attr) {
	L().LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

func DebugAttrs(ctx context.Context, msg string, attrs ...slog.Attr) {
	L().LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

func WarnAttrs(ctx context.Context, msg string, attrs ...slog.Attr) {
	L().LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

func ErrorAttrs(ctx context.Context, msg string, attrs ...slog.Attr) {
	L().LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

func With(args ...any) *slog.Logger {
	return L().With(args...)
}

func Logger() *slog.Logger {
	return L()
}

func defaultLogger(level string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(level),
		AddSource: false,
	}

	format := strings.ToLower(strings.TrimSpace(os.Getenv("REPOMOVER_LOG_FORMAT")))
	if format == "json" {
		return slog.New(slog.NewJSONHandler(os.Stderr, opts))
	}

	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func ensureInitialized() {
	once.Do(func() {
		mu.Lock()
		if logger == nil {
			logger = defaultLogger(resolveLevel(""))
		}
		mu.Unlock()
	})
}

func resolveLevel(level string) string {
	if strings.TrimSpace(level) == "" {
		level = os.Getenv("REPOMOVER_LOG_LEVEL")
	}
	if strings.TrimSpace(level) == "" {
		level = "info"
	}
	return level
}
