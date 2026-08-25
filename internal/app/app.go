package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	adapterHTTP "github.com/mockctl-hq/mockctl/internal/adapter/http"
	"github.com/mockctl-hq/mockctl/internal/core/ports"
	"github.com/mockctl-hq/mockctl/internal/generator"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/shared"
)

// App acts as the Composition Root orchestrating all components.
type App struct {
	logger      shared.Logger
	systemStore ports.SystemStore
	stateStore  ports.StateStore
}

// StartServer orchestrates the instantiation and execution of the HTTP Server pipeline.
func (a *App) StartServer(ctx context.Context, port string) error {
	// 1. Generate Admin Token (Secure 32-byte hex)
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return err
	}
	adminToken := hex.EncodeToString(tokenBytes)
	
	// Save to SystemStore
	if err := a.systemStore.SaveAuthToken(ctx, adminToken); err != nil {
		return err
	}
	a.logger.Info("Admin Access Token Generated", "token", adminToken)

	// 2. Setup Stubs for dependencies (ValueProvider, ChaosEvaluator, Clock, etc.)
	// TODO: Inject real concrete implementations
	var def *generator.RuntimeDefinition
	var vp ports.ValueProvider
	var chaos ports.ChaosEvaluator
	clock := shared.NewRealClock()

	// 3. Wire RuntimeEngine
	engine := runtime.NewRuntimeEngine(a.logger, def, a.stateStore, chaos, vp, clock)

	// 4. Wire HTTPServer
	server := adapterHTTP.NewHTTPServer(a.logger, a.systemStore, engine)

	// 5. Start Server in a Goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		if err := server.Start(port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// 6. OS Signal Graceful Shutdown Hook
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		return err
	case <-sigChan:
		a.logger.Info("OS Signal received, beginning graceful shutdown...")
		
		// 5 second timeout for draining connections
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error("HTTP Server forced to shutdown", err)
		}

		// TODO: Trigger StateStore dump to temp-state.json
		
		a.logger.Info("Graceful shutdown completed successfully.")
		return nil
	}
}
