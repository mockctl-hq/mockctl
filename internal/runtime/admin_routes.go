package runtime

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// setupAdminRoutes configures the reserved /__mockctl/* namespace.
func (s *HTTPServer) setupAdminRoutes() {
	s.router.Route("/__mockctl", func(r chi.Router) {
		r.Use(s.adminSecurityMiddleware)

		r.Get("/health", s.handleHealth)
		
		r.Route("/projects", func(pr chi.Router) {
			pr.Get("/", s.handleListProjects)
			pr.Post("/", s.handleCreateProject)
			
			pr.Route("/{projectName}", func(pnr chi.Router) {
				pnr.Delete("/", s.handleDeleteProject)
				pnr.Post("/endpoints", s.handleAddEndpoint)
				pnr.Post("/overrides", s.handleSetOverrides)
				pnr.Post("/state/reset", s.handleResetState)
				pnr.Patch("/chaos", s.handleUpdateChaos)
			})
		})
	})
}

// adminSecurityMiddleware enforces Localhost binding, Rate Limiting, and Authorization
func (s *HTTPServer) adminSecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Localhost Binding (Basic check)
		host := r.RemoteAddr
		if !strings.HasPrefix(host, "127.0.0.1:") && !strings.HasPrefix(host, "[::1]:") {
			s.sendAdminError(w, "FORBIDDEN", "Admin API is restricted to localhost", http.StatusForbidden)
			return
		}

		// 2. Rate Limiting (EDL-054)
		if !s.rateLimiter.Allow() {
			s.sendAdminError(w, "RATE_LIMIT_EXCEEDED", "Admin API rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// 3. Authorization Bearer Token
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			s.sendAdminError(w, "UNAUTHORIZED", "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		token = strings.Trim(token, `"'`)
		expectedToken, err := s.systemStore.GetAuthToken(r.Context())
		if err != nil || expectedToken == "" {
			s.sendAdminError(w, "UNAUTHORIZED", "Admin token not configured", http.StatusUnauthorized)
			return
		}

		// Security Rule (PKS-028): Constant-Time Comparison
		if subtle.ConstantTimeCompare([]byte(token), []byte(expectedToken)) != 1 {
			s.sendAdminError(w, "UNAUTHORIZED", fmt.Sprintf("Invalid admin token. Received: '%s', Expected: '%s'", token, expectedToken), http.StatusUnauthorized)
			return
		}

		// 4. Content-Type Validation for Mutations
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if !strings.HasPrefix(contentType, "application/json") && !strings.HasPrefix(contentType, "multipart/form-data") {
				s.sendAdminError(w, "UNSUPPORTED_MEDIA_TYPE", "Content-Type must be application/json or multipart/form-data", http.StatusUnsupportedMediaType)
				return
			}
		}

		// Accept-Version & CORS
		if r.Header.Get("Accept-Version") != "v1" {
			s.sendAdminError(w, "BAD_REQUEST", "Accept-Version header must be v1", http.StatusBadRequest)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

func (s *HTTPServer) sendAdminError(w http.ResponseWriter, code, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error": map[string]any{
			"code":    code,
			"message": msg,
			"status":  status,
		},
	})
}

// handleHealth returns the current status of the Mock Server Engine.
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status": "healthy",
	})
}
