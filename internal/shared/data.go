package shared

import "context"

// ValueProvider abstracts the generation of dynamic fake data (like UUIDs)
// ensuring the RuntimeEngine does not depend directly on third-party libraries (e.g., gofakeit).
type ValueProvider interface {
	// GenerateUUID returns a cryptographically secure random identifier.
	GenerateUUID(ctx context.Context) string

	// GenerateString generates a random string based on a format/schema.
	GenerateString(ctx context.Context, format string) string

	// GenerateInt generates a random integer within a range.
	GenerateInt(ctx context.Context, min, max int) int

	// GenerateBoolean generates a random boolean.
	GenerateBoolean(ctx context.Context) bool
}

// ChaosEvaluator determines if a request should intentionally fail or delay.
type ChaosEvaluator interface {
	// Evaluate takes the incoming context and applies potential latency (blocking)
	// or returns an error status code to simulate failure.
	// Returns an HTTP status code to inject (0 if no error is injected).
	Evaluate(ctx context.Context) (int, error)
	UpdateConfig(ctx context.Context, errorRate int, latencyMs int)
}
