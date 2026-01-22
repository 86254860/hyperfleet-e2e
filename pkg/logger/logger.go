package logger

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/onsi/ginkgo/v2"

	"github.com/openshift-hyperfleet/hyperfleet-e2e/pkg/config"
)

var (
	// component is the component name for all logs
	component = "hyperfleet-e2e"

	// version is set at build time
	version = "dev"

	// hostname is the machine hostname
	hostname = ""
)

// GinkgoLogHandler wraps a standard handler to inject test context automatically
type GinkgoLogHandler struct {
	slog.Handler
}

// Handle adds Ginkgo test context to log records
func (h *GinkgoLogHandler) Handle(ctx context.Context, r slog.Record) error {
	report := ginkgo.CurrentSpecReport()

	if report.LeafNodeText != "" {
		r.AddAttrs(slog.String("test_case", report.LeafNodeText))
	}

	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a new GinkgoLogHandler that wraps the underlying handler's WithAttrs result
func (h *GinkgoLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &GinkgoLogHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new GinkgoLogHandler that wraps the underlying handler's WithGroup result
func (h *GinkgoLogHandler) WithGroup(name string) slog.Handler {
	return &GinkgoLogHandler{Handler: h.Handler.WithGroup(name)}
}

// Init initializes the global logger based on configuration
// Returns nil on success, error on failure
func Init(cfg *config.LogConfig, buildVersion string) error {
	if buildVersion != "" {
		version = buildVersion
	}

	var err error
	hostname, err = os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	var output io.Writer
	switch cfg.Output {
	case config.LogOutputStderr:
		output = os.Stderr
	default: // stdout or empty
		output = os.Stdout
	}

	level := parseLevel(cfg.Level)

	var baseHandler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	switch cfg.Format {
	case config.LogFormatJSON:
		baseHandler = slog.NewJSONHandler(output, opts)
	default: // text or empty
		baseHandler = slog.NewTextHandler(output, opts)
	}

	baseHandler = baseHandler.WithAttrs([]slog.Attr{
		slog.String("component", component),
		slog.String("version", version),
		slog.String("hostname", hostname),
	})

	handler := &GinkgoLogHandler{Handler: baseHandler}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}

// Info logs an info level message
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

// Debug logs a debug level message
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// Warn logs a warn level message
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
}

// Error logs an error level message
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

// parseLevel converts string level to slog.Level
func parseLevel(level string) slog.Level {
	switch level {
	case config.LogLevelDebug:
		return slog.LevelDebug
	case config.LogLevelInfo:
		return slog.LevelInfo
	case config.LogLevelWarn:
		return slog.LevelWarn
	case config.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// InfoWithCluster logs info with cluster context
func InfoWithCluster(clusterID, msg string, args ...any) {
	slog.Info(msg, append([]any{"cluster_id", clusterID}, args...)...)
}

// InfoWithResource logs info with resource context
func InfoWithResource(resourceType, resourceID, msg string, args ...any) {
	slog.Info(msg, append([]any{"resource_type", resourceType, "resource_id", resourceID}, args...)...)
}

// ErrorWithError logs error with error field
func ErrorWithError(msg string, err error, args ...any) {
	slog.Error(msg, append([]any{"error", err}, args...)...)
}
