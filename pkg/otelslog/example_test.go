package otelslog

import (
	"go.opentelemetry.io/otel/log/noop"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func Example() {
	// Use a working LoggerProvider implementation instead e.g. using go.opentelemetry.io/otel/sdk/log.
	provider := noop.NewLoggerProvider()

	// Create an *slog.Logger and use it in your application.
	otelslog.NewLogger("my/pkg/name", otelslog.WithLoggerProvider(provider))
}
