package runtime

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminRoutes_Middleware(t *testing.T) {
	logger := &MockLogger{}
	store := NewMockSystemStore()
	store.AuthToken = "valid-token"

	manager := NewMockProjectManager()
	gateway := NewProjectGateway()

	server := NewHTTPServer(logger, store, manager, gateway)

	// Test 1: Missing Accept-Version
	req := httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for missing Accept-Version, got %d", w.Code)
	}

	// Test 2: Missing Token
	req = httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Accept-Version", "v1")
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for missing token, got %d", w.Code)
	}

	// Test 3: Invalid Token
	req = httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Accept-Version", "v1")
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized for invalid token, got %d", w.Code)
	}

	// Test 4: Valid Token
	req = httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("Accept-Version", "v1")
	req.Header.Set("Authorization", "Bearer valid-token")
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for valid request, got %d", w.Code)
	}
}
