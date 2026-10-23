package runtime

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mockctl-hq/mockctl/internal/shared"
	"golang.org/x/time/rate"
)

// DomainError represents a standardized error within the system.
type DomainError struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
}

// HTTPServer acts as the presentation boundary, wrapping the Chi router
// and the underlying net/http Server.
type HTTPServer struct {
	router      *chi.Mux
	systemStore shared.SystemStore
	engine      http.Handler
	manager     ProjectManager
	logger      shared.Logger
	httpServer  *http.Server
	rateLimiter *rate.Limiter
	broker      EventSubscriber
	corsEnabled atomic.Bool
}

// NewHTTPServer initializes a new HTTPServer with the required middleware pipeline.
func NewHTTPServer(logger shared.Logger, store shared.SystemStore, manager ProjectManager, engine http.Handler, broker EventSubscriber) *HTTPServer {
	r := chi.NewRouter()

	s := &HTTPServer{
		router:      r,
		systemStore: store,
		engine:      engine,
		manager:     manager,
		logger:      logger,
		// EDL-054: Rate limiter (100 req/sec, burst of 10)
		rateLimiter: rate.NewLimiter(rate.Limit(100), 10),
		broker:      broker,
	}
	s.corsEnabled.Store(true)

	// Strict Middleware Chain (EDL-025, PKS-024)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(s.loggerMiddleware) // Custom Logger to redact Auth header
	// TODO: Metrics Middleware

	s.setupRoutes()

	return s
}

// SetCORS dynamically toggles global CORS on mock endpoints.
func (s *HTTPServer) SetCORS(enabled bool) {
	s.corsEnabled.Store(enabled)
}

// CORSEnabled reports whether global CORS is active for mock endpoints.
func (s *HTTPServer) CORSEnabled() bool {
	return s.corsEnabled.Load()
}

// mockCorsMiddleware conditionally injects CORS headers and handles OPTIONS preflight
// requests based on the Global CORS setting.
func (s *HTTPServer) mockCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.corsEnabled.Load() {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")

		reqHeaders := r.Header.Get("Access-Control-Request-Headers")
		if reqHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "*")
		}
		w.Header().Set("Access-Control-Expose-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// loggerMiddleware integrates with shared.Logger and scrubs Authorization headers (PKS-028)
func (s *HTTPServer) loggerMiddleware(next http.Handler) http.Handler {
	// Initialize a dedicated logger for HTTP requests
	var reqLogger *log.Logger
	if userHome, err := os.UserHomeDir(); err == nil && userHome != "" {
		reqLogDir := filepath.Join(userHome, ".mockctl")
		_ = os.MkdirAll(reqLogDir, 0755)
		reqLogPath := filepath.Join(reqLogDir, "requests.log")
		// #nosec G302 G304 -- reqLogPath is deterministic in ~/.mockctl and uses secure 0600 permissions
		logFile, err := os.OpenFile(reqLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err == nil {
			reqLogger = log.New(logFile, "", log.Ldate|log.Ltime)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a sanitized clone of headers for logging
		sanitizedHeaders := r.Header.Clone()
		if sanitizedHeaders.Get("Authorization") != "" {
			sanitizedHeaders.Set("Authorization", "[REDACTED]")
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		if reqLogger != nil {
			logStr := fmt.Sprintf("HTTP %s %s | status: %d | bytes: %d | %dms",
				r.Method, r.URL.Path, ww.Status(), ww.BytesWritten(), time.Since(start).Milliseconds())
			reqLogger.Println(logStr)
		}
	})
}

func (s *HTTPServer) setupRoutes() {
	// Mount Admin Routes under /__mockctl/*
	s.setupAdminRoutes()

	// Mount RuntimeEngine as the catch-all wildcard for mock execution with optional CORS
	s.router.Mount("/", s.mockCorsMiddleware(s.engine))
}

// Start binds the server to the specified port and begins listening.
func (s *HTTPServer) Start(port string) error {
	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	s.logger.Info("Starting HTTP Server on port " + port)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the server, rejecting new requests and draining active ones.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		s.logger.Info("Gracefully shutting down HTTP Server...")
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
