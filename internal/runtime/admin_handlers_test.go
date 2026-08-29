package runtime

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminHandlers_ListProjects(t *testing.T) {
	logger := &MockLogger{}
	store := NewMockSystemStore()
	store.AuthToken = "test"
	manager := NewMockProjectManager()
	
	// Pre-populate some projects in the manager
	manager.Projects["proj1"] = map[string]any{"status": "active"}
	manager.Projects["proj2"] = map[string]any{"status": "active"}

	server := NewHTTPServer(logger, store, manager, NewProjectGateway())

	req := httptest.NewRequest(http.MethodGet, "/__mockctl/projects", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Accept-Version", "v1")
	req.Header.Set("Authorization", "Bearer test")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected JSON body, got empty")
	}
}

func TestAdminHandlers_DeleteProject(t *testing.T) {
	logger := &MockLogger{}
	store := NewMockSystemStore()
	store.AuthToken = "test"
	manager := NewMockProjectManager()
	manager.Projects["proj1"] = map[string]any{"status": "active"}

	server := NewHTTPServer(logger, store, manager, NewProjectGateway())

	req := httptest.NewRequest(http.MethodDelete, "/__mockctl/projects/proj1", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Accept-Version", "v1")
	req.Header.Set("Authorization", "Bearer test")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Verify it was deleted
	if len(manager.Projects) != 0 {
		t.Error("Expected project to be deleted")
	}
}
