package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/shared"
	"github.com/mockctl-hq/mockctl/internal/storage"
	"go.etcd.io/bbolt"
)

// App acts as the Composition Root orchestrating all components.
type App struct {
	logger           shared.Logger
	fileSystem       shared.FileSystem
	systemStore      shared.SystemStore
	stateStore       shared.StateStore
	gateway          *runtime.ProjectGateway
	broker           *runtime.EventBroker
	server           *runtime.HTTPServer
	metricsCollector *runtime.MetricsCollector
	persister        *storage.TelemetryPersister
	corsEnabled      atomic.Bool
}

// EventBroker exposes the internal EventBroker to the presentation layer.
func (a *App) EventBroker() *runtime.EventBroker { return a.broker }

// Gateway exposes the internal ProjectGateway to the presentation layer.
func (a *App) Gateway() *runtime.ProjectGateway { return a.gateway }

// StateStore exposes the internal StateStore to the presentation layer for the State Inspector.
func (a *App) StateStore() shared.StateStore { return a.stateStore }

// GetTelemetryHistory retrieves past telemetry events from the system database.
func (a *App) GetTelemetryHistory(ctx context.Context, limit int) ([]map[string]any, error) {
	return a.systemStore.GetTelemetryHistory(ctx, limit)
}

// GetProjectModTime checks the modification time of an external file (e.g. api.yaml).
func (a *App) GetProjectModTime(filePath string) (time.Time, error) {
	return a.fileSystem.ModTime(filePath)
}

// SetGlobalCORS updates the global CORS policy on the running HTTP server and internal state.
func (a *App) SetGlobalCORS(enabled bool) {
	a.corsEnabled.Store(enabled)
	if a.server != nil {
		a.server.SetCORS(enabled)
	}
}

// GlobalCORSEnabled returns whether global CORS is currently enabled.
func (a *App) GlobalCORSEnabled() bool {
	return a.corsEnabled.Load()
}

// SetStrictValidation updates the OpenAPI validation mode across all running mock engines.
func (a *App) SetStrictValidation(enabled bool) {
	if a.gateway != nil {
		a.gateway.SetStrictValidation(enabled)
	}
}

// StrictValidationEnabled reports whether strict OpenAPI schema validation is enabled.
func (a *App) StrictValidationEnabled() bool {
	if a.gateway != nil {
		return a.gateway.StrictValidationEnabled()
	}
	return true
}

// SaveSettingsConfig persists the serialized TUI settings into the system database
// and syncs runtime policies (like Global CORS and Strict OpenAPI Validation).
func (a *App) SaveSettingsConfig(ctx context.Context, cfgJSON []byte) error {
	var cfg struct {
		GlobalCORS              *bool `json:"global_cors"`
		StrictOpenAPIValidation *bool `json:"strict_openapi_validation"`
	}
	if json.Unmarshal(cfgJSON, &cfg) == nil {
		if cfg.GlobalCORS != nil {
			a.SetGlobalCORS(*cfg.GlobalCORS)
		}
		if cfg.StrictOpenAPIValidation != nil {
			a.SetStrictValidation(*cfg.StrictOpenAPIValidation)
		}
	}
	return a.systemStore.SetSetting(ctx, "tui_settings", string(cfgJSON))
}

// LoadSettingsConfig loads the persisted TUI settings from the system database.
func (a *App) LoadSettingsConfig(ctx context.Context) ([]byte, error) {
	val, err := a.systemStore.GetSetting(ctx, "tui_settings")
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

// GetSetting retrieves a persistent configuration value from the system database.
func (a *App) GetSetting(ctx context.Context, key string) (string, error) {
	return a.systemStore.GetSetting(ctx, key)
}

// SetSetting saves a persistent configuration value into the system database.
func (a *App) SetSetting(ctx context.Context, key string, value string) error {
	return a.systemStore.SetSetting(ctx, key, value)
}

// Ensure App implements ProjectManager
var _ runtime.ProjectManager = (*App)(nil)

// NewApp initializes the Mock:ctl master daemon without starting the server.
func NewApp(ctx context.Context) (*App, error) {
	// Task 1.4 (TUI-Safe & Docker-Safe Logs)
	logger := &slogLogger{}

	// Task 1.2 (App Bootstrapping & Cross-Platform Paths)
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}
	mockctlDir := filepath.Join(userHome, ".mockctl")
	
	fs := storage.NewLocalFileSystem()
	
	if err := fs.MkdirAll(mockctlDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create .mockctl dir: %w", err)
	}

	dbPath := filepath.Join(mockctlDir, "system.db")
	systemStore, err := storage.NewBBoltSystemStore(dbPath)
	if err != nil {
		if errors.Is(err, bbolt.ErrTimeout) {
			fmt.Println("🚫 Error: Mock:ctl is already running in another terminal. Please close it first.")
			os.Exit(1)
		}
		return nil, fmt.Errorf("failed to open system database: %w", err)
	}

	// Check for existing Admin Token, generate if missing
	tokenFile := filepath.Join(mockctlDir, "admin.token")
	// #nosec G304 -- tokenFile path is deterministically built from user home directory
	tokenData, err := fs.ReadFile(tokenFile)
	if err != nil {
		// Generate Admin Token (Secure 32-byte hex)
		tokenBytes := make([]byte, 32)
		if _, err := rand.Read(tokenBytes); err != nil {
			return nil, err
		}
		tokenData = []byte(hex.EncodeToString(tokenBytes))
		if err := fs.WriteFile(tokenFile, tokenData, 0600); err != nil {
			return nil, fmt.Errorf("failed to write admin.token: %w", err)
		}
		logger.Info("Admin Access Token Generated and Saved")
	}

	// Always sync the token to the DB on startup
	if err := systemStore.SaveAuthToken(ctx, string(tokenData)); err != nil {
		return nil, fmt.Errorf("failed to sync admin token to db: %w", err)
	}

	gateway := runtime.NewProjectGateway()
	broker := runtime.NewEventBroker()
	
	a := &App{
		logger:      logger,
		fileSystem:  fs,
		systemStore: systemStore,
		stateStore:  storage.NewMemoryStateStore(), // Temporary
		gateway:     gateway,
		broker:      broker,
	}
	a.corsEnabled.Store(true) // Default to true

	// Restore persisted runtime settings (CORS & Strict Validation) if present
	if settingsData, err := a.LoadSettingsConfig(ctx); err == nil && len(settingsData) > 0 {
		var cfg struct {
			GlobalCORS              *bool `json:"global_cors"`
			StrictOpenAPIValidation *bool `json:"strict_openapi_validation"`
		}
		if json.Unmarshal(settingsData, &cfg) == nil {
			if cfg.GlobalCORS != nil {
				a.corsEnabled.Store(*cfg.GlobalCORS)
			}
			if cfg.StrictOpenAPIValidation != nil {
				a.SetStrictValidation(*cfg.StrictOpenAPIValidation)
			}
		}
	}

	// Non-Blocking Re-hydration
	if err := a.HydrateProjects(ctx); err != nil {
		logger.Error("Failed to hydrate projects", err)
	}

	return a, nil
}

// ServeInBackground starts the background telemetry and HTTP server without blocking.
func (a *App) ServeInBackground(ctx context.Context, port int) chan error {
	a.persister = storage.NewTelemetryPersister(a.broker, a.systemStore.(*storage.BBoltSystemStore).DB())
	a.persister.Start(ctx)

	a.metricsCollector = runtime.NewMetricsCollector(a.broker)
	a.metricsCollector.Start(ctx)

	a.server = runtime.NewHTTPServer(a.logger, a.systemStore, a, a.gateway, a.broker)
	a.server.SetCORS(a.corsEnabled.Load())

	serverErrChan := make(chan error, 1)
	go func() {
		if err := a.server.Start(strconv.Itoa(port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()
	return serverErrChan
}

// Shutdown stops all background components and releases the database lock gracefully.
func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("Beginning graceful shutdown...")

	if a.metricsCollector != nil {
		a.metricsCollector.Stop()
	}
	if a.persister != nil {
		a.persister.Stop()
	}
	if a.broker != nil {
		a.broker.Stop()
	}

	var err error
	if a.server != nil {
		err = a.server.Shutdown(ctx)
		if err != nil {
			a.logger.Error("HTTP Server forced to shutdown", err)
		}
	}
	
	if a.systemStore != nil {
		if dbErr := a.systemStore.Close(ctx); dbErr != nil && err == nil {
			err = dbErr
		}
	}

	a.logger.Info("Graceful shutdown completed successfully.")
	return err
}

// StartDaemon initializes and starts the Mock:ctl master daemon in blocking mode (Legacy CLI support).
func StartDaemon(ctx context.Context, port int) error {
	a, err := NewApp(ctx)
	if err != nil {
		return err
	}
	
	serverErrChan := a.ServeInBackground(ctx, port)
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		_ = a.Shutdown(ctx)
		return err
	case <-sigChan:
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return a.Shutdown(shutdownCtx)
	}
}

// slogLogger proxies logs to the globally configured slog.Logger (which is TUI-safe).
type slogLogger struct{}

func (l *slogLogger) Info(msg string, args ...any) { slog.Info(msg, args...) }
func (l *slogLogger) Warn(msg string, args ...any) { slog.Warn(msg, args...) }
func (l *slogLogger) Error(msg string, err error, args ...any) {
	slog.Error(msg, append(args, "error", err)...)
}
func (l *slogLogger) Debug(msg string, args ...any) { slog.Debug(msg, args...) }
