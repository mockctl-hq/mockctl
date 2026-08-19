package ports

import (
	"context"
	"errors"
)

// Sentinel Errors (Error Boundaries)
var (
	ErrNotFound       = errors.New("record not found")
	ErrLimitReached   = errors.New("collection size limit reached (OOM protection)")
	ErrInvalidPayload = errors.New("invalid payload structure")
	ErrDatabaseLocked = errors.New("database is locked by another process (read-only mode active)")
)

// StateStore defines the operations for the in-memory ephemeral database.
type StateStore interface {
	Insert(ctx context.Context, collection string, id string, payload map[string]any) error
	Get(ctx context.Context, collection string, id string) (map[string]any, error)
	List(ctx context.Context, collection string) ([]map[string]any, error)
	Update(ctx context.Context, collection string, id string, payload map[string]any) error
	Patch(ctx context.Context, collection string, id string, partial map[string]any) error
	Delete(ctx context.Context, collection string, id string) error
	Reset(ctx context.Context) error
}

// SystemStore defines the operations for the persistent bbolt database.
type SystemStore interface {
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key string, value string) error
	SaveAuthToken(ctx context.Context, token string) error
	GetAuthToken(ctx context.Context) (string, error)
	LogTelemetry(ctx context.Context, event string, data map[string]any) error
	Close(ctx context.Context) error
}
