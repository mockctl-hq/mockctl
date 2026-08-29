package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/shared"
	"github.com/mockctl-hq/mockctl/internal/storage"
)

// App acts as the Composition Root orchestrating all components.
type App struct {
	logger      shared.Logger
	systemStore shared.SystemStore
	stateStore  shared.StateStore
	gateway     *runtime.ProjectGateway
}

// Ensure App implements ProjectManager
var _ runtime.ProjectManager = (*App)(nil)



// StartDaemon initializes and starts the Mock:ctl master daemon.
func StartDaemon(ctx context.Context, port int) error {
	// Task 1.4 (TUI-Safe & Docker-Safe Logs)
	logger := &consoleLogger{}

	// Task 1.2 (App Bootstrapping & Cross-Platform Paths)
	userHome, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}
	mockctlDir := filepath.Join(userHome, ".mockctl")
	if err := os.MkdirAll(mockctlDir, 0750); err != nil {
		return fmt.Errorf("failed to create .mockctl dir: %w", err)
	}

	dbPath := filepath.Join(mockctlDir, "system.db")
	systemStore, err := storage.NewBBoltSystemStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open system database: %w", err)
	}
	defer systemStore.Close(ctx)

	// 1. Generate Admin Token (Secure 32-byte hex)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	adminToken := hex.EncodeToString(tokenBytes)
	if err := systemStore.SaveAuthToken(ctx, adminToken); err != nil {
		return fmt.Errorf("failed to save admin token: %w", err)
	}
	fmt.Printf("\n======================================================\n")
	fmt.Printf("ADMIN TOKEN: %s\n", adminToken)
	fmt.Printf("======================================================\n\n")
	logger.Info("Admin Access Token Generated and Saved")
	
	tokenFile := filepath.Join(mockctlDir, "admin.token")
	if err := os.WriteFile(tokenFile, []byte(adminToken), 0600); err != nil {
		return fmt.Errorf("failed to write admin.token: %w", err)
	}
	defer os.Remove(tokenFile) // cleanup on exit

	// Task 2.2: ProjectGateway Router
	gateway := runtime.NewProjectGateway()

	a := &App{
		logger:      logger,
		systemStore: systemStore,
		stateStore:  storage.NewMemoryStateStore(), // Temporary
		gateway:     gateway,
	}

	// Task 1.3: Non-Blocking Re-hydration
	if err := a.HydrateProjects(ctx); err != nil {
		logger.Error("Failed to hydrate projects", err)
	}

	// Task 2.1: Downward Dependency Flow - Inject App as ProjectManager
	server := runtime.NewHTTPServer(logger, systemStore, a, gateway)

	serverErrChan := make(chan error, 1)
	go func() {
		if err := server.Start(strconv.Itoa(port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// Task 2.4 (Graceful Storage Shutdown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		return err
	case <-sigChan:
		logger.Info("OS Signal received, beginning graceful shutdown...")

		// Graceful Shutdown Hang Fix: strict 5 second timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP Server forced to shutdown", err)
		}
		
		// Flush all active crud data if needed here

		logger.Info("Graceful shutdown completed successfully.")
		return nil
	}
}

// consoleLogger is a simple stdout logger
type consoleLogger struct{}
func (l *consoleLogger) Info(msg string, args ...any) { fmt.Printf("[INFO] %s %v\n", msg, args) }
func (l *consoleLogger) Warn(msg string, args ...any) { fmt.Printf("[WARN] %s %v\n", msg, args) }
func (l *consoleLogger) Error(msg string, err error, args ...any) { fmt.Printf("[ERROR] %s: %v %v\n", msg, err, args) }
func (l *consoleLogger) Debug(msg string, args ...any) { fmt.Printf("[DEBUG] %s %v\n", msg, args) }
