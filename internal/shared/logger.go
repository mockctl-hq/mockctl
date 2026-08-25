package shared

// Logger abstracts the logging framework (e.g., log/slog) to prevent
// tight coupling across the system architecture.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, err error, args ...any)
	Debug(msg string, args ...any)
}
