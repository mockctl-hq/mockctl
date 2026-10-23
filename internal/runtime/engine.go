package runtime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/mockctl-hq/mockctl/internal/generator"
	"github.com/mockctl-hq/mockctl/internal/shared"
)

// RuntimeEngine is the core executor that matches routes against the OpenAPI definition
// and interacts with the StateStore. It strictly isolates itself from the presentation layer router.
type RuntimeEngine struct {
	logger           shared.Logger
	definition       *generator.RuntimeDefinition
	state            shared.StateStore
	chaos            shared.ChaosEvaluator
	valueProvider    shared.ValueProvider
	clock            shared.Clock
	broker           EventPublisher
	projectName      string
	mux              http.Handler // Changed to http.Handler to hold the wrapped mux
	strictValidation atomic.Bool
}

// Definition exposes the compiled runtime definition for the UI components.
func (e *RuntimeEngine) Definition() *generator.RuntimeDefinition {
	return e.definition
}

// NewRuntimeEngine creates a new instance of the RuntimeEngine.
func NewRuntimeEngine(
	l shared.Logger,
	def *generator.RuntimeDefinition,
	store shared.StateStore,
	chaos shared.ChaosEvaluator,
	vp shared.ValueProvider,
	clk shared.Clock,
	broker EventPublisher,
	projectName string,
) *RuntimeEngine {
	rawMux := http.NewServeMux()

	e := &RuntimeEngine{
		logger:        l,
		definition:    def,
		state:         store,
		chaos:         chaos,
		valueProvider: vp,
		clock:         clk,
		broker:        broker,
		projectName:   projectName,
		mux:           rawMux,
	}
	e.strictValidation.Store(true)

	e.setupRoutes(rawMux)

	// Apply TelemetryMiddleware wrapping the raw mux
	if broker != nil {
		e.mux = TelemetryMiddleware(broker, projectName)(rawMux)
	}

	return e
}

// SetStrictValidation updates the OpenAPI schema validation mode (strict vs permissive).
func (e *RuntimeEngine) SetStrictValidation(enabled bool) {
	e.strictValidation.Store(enabled)
}

// StrictValidationEnabled reports whether strict OpenAPI schema validation is active.
func (e *RuntimeEngine) StrictValidationEnabled() bool {
	return e.strictValidation.Load()
}

func (e *RuntimeEngine) Chaos() shared.ChaosEvaluator {
	return e.chaos
}

func (e *RuntimeEngine) setupRoutes(rawMux *http.ServeMux) {
	if e.definition == nil {
		return
	}

	for _, endpoint := range e.definition.Endpoints {
		pattern := endpoint.Method + " " + endpoint.Path
		ep := endpoint // Capture for closure
		rawMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			e.processRequest(w, r, ep)
		})
	}
}

// ServeHTTP makes the RuntimeEngine compatible with net/http routers (like Chi).
// It acts as the wildcard catch-all for mock routes.
func (e *RuntimeEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. MaxBytesReader (5MB DoS Protection)
	if r.Body != nil {
		r.Body = http.MaxBytesReader(w, r.Body, 5*1024*1024)
	}

	// 2. Content-Type Validation for Mutations (PKS-027)
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		if r.Header.Get("Content-Type") != "application/json" {
			e.sendError(w, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
	}

	// 3. Delegate to Internal ServeMux
	e.mux.ServeHTTP(w, r)
}

func (e *RuntimeEngine) processRequest(w http.ResponseWriter, r *http.Request, endpoint generator.EndpointHandler) {
	// 1. Chaos Evaluation (PKS-029)
	if e.chaos != nil {
		if status, err := e.chaos.Evaluate(r.Context()); status != 0 {
			msg := fmt.Sprintf("Simulated chaos failure (HTTP %d)", status)
			if err != nil {
				msg = err.Error()
			}
			e.sendError(w, "CHAOS_INJECTED", msg, status)
			return
		}
	}

	// 2. Static Override Check
	// If this endpoint has a user-defined static override, return the override directly.
	if resp, ok := endpoint.Responses[http.StatusOK]; ok && resp.IsOverride {
		e.formatResponse(w, http.StatusOK, resp.Headers, resp.Body)
		return
	}
	for status, resp := range endpoint.Responses {
		if resp.IsOverride {
			e.formatResponse(w, status, resp.Headers, resp.Body)
			return
		}
	}

	// 3. OpenAPI Request Body Validation
	if endpoint.RequestValidator != nil && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch) {
		var bodyBytes []byte
		if r.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(r.Body)
			if err != nil {
				e.sendError(w, "INVALID_REQUEST", "Failed to read request body", http.StatusBadRequest)
				return
			}
			// Restore r.Body so downstream handlers can decode it
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		if err := endpoint.RequestValidator.ValidateJSON(bodyBytes); err != nil {
			if e.strictValidation.Load() {
				e.sendError(w, "SCHEMA_VALIDATION_FAILED", err.Error(), http.StatusBadRequest)
				return
			}
			if e.logger != nil {
				e.logger.Warn("Request body failed schema validation (permissive mode allowed)", "error", err, "path", r.URL.Path)
			}
		}
	}

	// 3. Dispatch to specific handler based on Method
	switch r.Method {
	case http.MethodGet:
		e.handleGet(r.Context(), w, r, endpoint)
	case http.MethodPost:
		e.handlePost(r.Context(), w, r, endpoint)
	case http.MethodPut:
		e.handlePut(r.Context(), w, r, endpoint)
	case http.MethodPatch:
		e.handlePatch(r.Context(), w, r, endpoint)
	case http.MethodDelete:
		e.handleDelete(r.Context(), w, r, endpoint)
	default:
		e.sendError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// sendError is a helper to format standard DomainError JSON responses
func (e *RuntimeEngine) sendError(w http.ResponseWriter, code string, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	payload := map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    code,
			"message": msg,
			"status":  status,
		},
	}
	_ = json.NewEncoder(w).Encode(payload)
}
