package runtime

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// setupAdminRoutes configures the reserved /__mockctl/* namespace.
func (s *HTTPServer) setupAdminRoutes() {
	s.router.Route("/__mockctl", func(r chi.Router) {
		r.Use(s.adminCorsPreflightMiddleware)
		r.Use(s.adminSecurityMiddleware)

		r.Get("/health", s.handleHealth)
		
		// Task 3.2: Map GET /__mockctl/events to handleSSEEvents
		r.Get("/events", s.handleSSEEvents)

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
