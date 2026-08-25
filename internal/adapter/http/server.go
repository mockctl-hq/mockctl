package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mockctl-hq/mockctl/internal/core/ports"
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
	systemStore ports.SystemStore
	engine      http.Handler
	logger      shared.Logger
	httpServer  *http.Server
	rateLimiter *rate.Limiter
}

// NewHTTPServer initializes a new HTTPServer with the required middleware pipeline.
func NewHTTPServer(logger shared.Logger, store ports.SystemStore, engine http.Handler) *HTTPServer {
	r := chi.NewRouter()

	s := &HTTPServer{
		router:      r,
		systemStore: store,
		engine:      engine,
		logger:      logger,
		// EDL-054: Rate limiter (100 req/sec, burst of 10)
		rateLimiter: rate.NewLimiter(rate.Limit(100), 10),
	}

	// Strict Middleware Chain (EDL-025, PKS-024)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(s.loggerMiddleware) // Custom Logger to redact Auth header
	// TODO: Metrics Middleware

	s.setupRoutes()

	return s
}

// loggerMiddleware integrates with shared.Logger and scrubs Authorization headers (PKS-028)
func (s *HTTPServer) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a sanitized clone of headers for logging
		sanitizedHeaders := r.Header.Clone()
		if sanitizedHeaders.Get("Authorization") != "" {
			sanitizedHeaders.Set("Authorization", "[REDACTED]")
		}

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		s.logger.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration_ms", time.Since(start).Milliseconds(),
			"headers", sanitizedHeaders,
		)
	})
}

func (s *HTTPServer) setupRoutes() {
	// Mount Admin Routes under /__mockctl/*
	s.setupAdminRoutes()

	// Mount RuntimeEngine as the catch-all wildcard for mock execution
	s.router.Mount("/", s.engine)
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
