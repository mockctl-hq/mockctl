package db

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/mockctl-hq/mockctl/internal/core/ports"
)

func TestBBoltSystemStore_CRUD(t *testing.T) {
	ctx := context.Background()
	// PKS-029 Rule: Use t.TempDir() for isolated Integration Testing
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "config.db")

	store, err := NewBBoltSystemStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}
	// Ensure proper cleanup
	t.Cleanup(func() {
		_ = store.Close(ctx)
	})

	// 1. Settings Bucket
	err = store.SetSetting(ctx, "theme", "dark")
	if err != nil {
		t.Fatalf("Failed to set setting: %v", err)
	}

	val, err := store.GetSetting(ctx, "theme")
	if err != nil {
		t.Fatalf("Failed to get setting: %v", err)
	}
	if val != "dark" {
		t.Errorf("Expected 'dark', got %s", val)
	}

	// 2. Auth Bucket (JWT Token)
	err = store.SaveAuthToken(ctx, "fake.jwt.token")
	if err != nil {
		t.Fatalf("Failed to save auth token: %v", err)
	}

	token, err := store.GetAuthToken(ctx)
	if err != nil {
		t.Fatalf("Failed to get auth token: %v", err)
	}
	if token != "fake.jwt.token" {
		t.Errorf("Expected 'fake.jwt.token', got %s", token)
	}

	// 3. Telemetry Bucket
	err = store.LogTelemetry(ctx, "app_start", map[string]any{"os": "linux", "arch": "arm64"})
	if err != nil {
		t.Fatalf("Failed to log telemetry: %v", err)
	}
}

func TestBBoltSystemStore_ReadOnlyFallback(t *testing.T) {
	ctx := context.Background()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "config.db")

	// Instance 1 (Primary - holds the OS file lock)
	store1, err := NewBBoltSystemStore(dbPath)
	if err != nil {
		t.Fatalf("Instance 1 failed: %v", err)
	}

	// Write baseline data
	_ = store1.SetSetting(ctx, "port", "8080")
	_ = store1.Close(ctx) // Explicitly close to release flock so store2 doesn't hang in test.

	// Instance 2 (Secondary)
	// Use the internal constructor to force read-only mode.
	store2, err := newBBoltSystemStore(dbPath, true)
	if err != nil {
		t.Fatalf("Instance 2 crashed instead of falling back to Read-Only: %v", err)
	}
	defer func() { _ = store2.Close(ctx) }()

	// 1. Verify Read operation works in Read-Only mode
	val, err := store2.GetSetting(ctx, "port")
	if err != nil || val != "8080" {
		t.Errorf("Read-Only instance failed to read data. err=%v, val=%s", err, val)
	}

	// 2. Verify Write operation is blocked and returns the correct Sentinel Error
	err = store2.SetSetting(ctx, "port", "9090")
	if err != ports.ErrDatabaseLocked {
		t.Errorf("Expected ErrDatabaseLocked on write in Read-Only mode, got: %v", err)
	}
}
