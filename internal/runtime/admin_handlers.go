package runtime

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func (s *HTTPServer) sendSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"data":    data,
	})
}

func (s *HTTPServer) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.manager.ListProjects(r.Context())
	if err != nil {
		s.sendAdminError(w, "INTERNAL_ERROR", "Failed to list projects", http.StatusInternalServerError)
		return
	}
	s.sendSuccess(w, projects)
}

func (s *HTTPServer) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	// Restrict payload to 10MB to prevent memory exhaustion (G120)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	// Task 3.1 & Phase 4 Concurrency Fix: Parse multipart stream efficiently
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		s.sendAdminError(w, "BAD_REQUEST", "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	projectName := r.FormValue("name")
	if !slugRegex.MatchString(projectName) {
		s.sendAdminError(w, "INVALID_NAME", "Project name must be a valid slug [a-z0-9-]+", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("spec")
	if err != nil {
		s.sendAdminError(w, "BAD_REQUEST", "Missing spec file", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		s.sendAdminError(w, "INTERNAL_ERROR", "Failed to read uploaded file", http.StatusInternalServerError)
		return
	}

	if err := s.manager.CreateProject(r.Context(), projectName, fileBytes); err != nil {
		s.sendAdminError(w, "PROJECT_CREATION_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}

	s.sendSuccess(w, map[string]string{"project_name": projectName, "status": "created"})
}

func (s *HTTPServer) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "projectName")
	if err := s.manager.DeleteProject(r.Context(), projectName); err != nil {
		s.sendAdminError(w, "DELETE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendSuccess(w, map[string]string{"project_name": projectName, "status": "deleted"})
}

func (s *HTTPServer) handleAddEndpoint(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "projectName")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendAdminError(w, "BAD_REQUEST", "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := s.manager.AddEndpoint(r.Context(), projectName, body); err != nil {
		s.sendAdminError(w, "ADD_ENDPOINT_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendSuccess(w, map[string]string{"status": "endpoint_added"})
}

func (s *HTTPServer) handleSetOverrides(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "projectName")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.sendAdminError(w, "BAD_REQUEST", "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := s.manager.SetOverrides(r.Context(), projectName, body); err != nil {
		s.sendAdminError(w, "SET_OVERRIDES_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendSuccess(w, map[string]string{"status": "overrides_set"})
}

func (s *HTTPServer) handleResetState(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "projectName")
	if err := s.manager.ResetState(r.Context(), projectName); err != nil {
		s.sendAdminError(w, "RESET_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendSuccess(w, map[string]string{"status": "state_reset"})
}

func (s *HTTPServer) handleUpdateChaos(w http.ResponseWriter, r *http.Request) {
	projectName := chi.URLParam(r, "projectName")

	var req struct {
		ErrorRate int `json:"error_rate"`
		LatencyMs int `json:"latency_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendAdminError(w, "BAD_REQUEST", "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := s.manager.UpdateChaos(r.Context(), projectName, req.ErrorRate, req.LatencyMs); err != nil {
		s.sendAdminError(w, "CHAOS_UPDATE_FAILED", err.Error(), http.StatusInternalServerError)
		return
	}
	s.sendSuccess(w, map[string]string{"status": "chaos_updated"})
}
