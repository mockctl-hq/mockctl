package runtime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// FakeSystemStore
type fakeSystemStore struct{}

func (f *fakeSystemStore) GetSetting(ctx context.Context, key string) (string, error)     { return "", nil }
func (f *fakeSystemStore) SetSetting(ctx context.Context, key string, value string) error { return nil }
func (f *fakeSystemStore) SaveAuthToken(ctx context.Context, token string) error          { return nil }
func (f *fakeSystemStore) GetAuthToken(ctx context.Context) (string, error) {
	return "admin-secret-token", nil
}
func (f *fakeSystemStore) LogTelemetry(ctx context.Context, event string, data map[string]any) error {
	return nil
}
func (f *fakeSystemStore) GetTelemetryHistory(ctx context.Context, limit int) ([]map[string]any, error) {
	return nil, nil
}
func (f *fakeSystemStore) SaveProject(ctx context.Context, name string, projectData []byte) error {
	return nil
}
func (f *fakeSystemStore) GetProject(ctx context.Context, name string) ([]byte, error) {
	return nil, nil
}
func (f *fakeSystemStore) ListProjects(ctx context.Context) (map[string][]byte, error) {
	return nil, nil
}
func (f *fakeSystemStore) DeleteProject(ctx context.Context, name string) error { return nil }
func (f *fakeSystemStore) Close(ctx context.Context) error                      { return nil }

// FakeEngine
type fakeEngine struct{}

func (f *fakeEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("mocked"))
}

// FakeProjectManager
type fakeProjectManager struct{}

func (f *fakeProjectManager) CreateProject(ctx context.Context, name string, payload []byte) error {
	return nil
}
func (f *fakeProjectManager) ListProjects(ctx context.Context) (map[string]map[string]any, error) {
	return nil, nil
}
func (f *fakeProjectManager) DeleteProject(ctx context.Context, name string) error { return nil }
func (f *fakeProjectManager) AddEndpoint(ctx context.Context, projectName string, endpointJSON []byte) error {
	return nil
}
func (f *fakeProjectManager) SetOverrides(ctx context.Context, projectName string, overridesJSON []byte) error {
	return nil
}
func (f *fakeProjectManager) ResetState(ctx context.Context, projectName string) error { return nil }
func (f *fakeProjectManager) UpdateChaos(ctx context.Context, projectName string, errorRate int, latencyMs int) error {
	return nil
}

func TestHTTPServer_AdminAuth(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer(&fakeLogger{}, &fakeSystemStore{}, &fakeProjectManager{}, &fakeEngine{}, nil)

	// Test 1: No Auth Header
	req := httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}

	// Test 2: Invalid Auth Header
	req = httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", rr.Code)
	}

	// Test 3: Valid Auth Header but missing Accept-Version
	req = httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Authorization", "Bearer admin-secret-token")
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest { // Missing Accept-Version -> 400
		t.Errorf("expected 400 Bad Request (missing Accept-Version), got %d", rr.Code)
	}

	// Test 4: Valid Request
	req = httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("Authorization", "Bearer admin-secret-token")
	req.Header.Set("Accept-Version", "v1")
	rr = httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
}

func TestHTTPServer_LocalhostBinding(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer(&fakeLogger{}, &fakeSystemStore{}, &fakeProjectManager{}, &fakeEngine{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/__mockctl/health", nil)
	req.RemoteAddr = "192.168.1.5:12345" // Not localhost
	rr := httptest.NewRecorder()
	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-localhost, got %d", rr.Code)
	}
}

func TestHTTPServer_MockCORS(t *testing.T) {
	t.Parallel()
	server := NewHTTPServer(&fakeLogger{}, &fakeSystemStore{}, &fakeProjectManager{}, &fakeEngine{}, nil)

	// 1. By default, CORS is enabled
	if !server.CORSEnabled() {
		t.Fatalf("expected CORS to be enabled by default")
	}

	// 2. Preflight OPTIONS request on mock route
	optReq := httptest.NewRequest(http.MethodOptions, "/petstore/pets", nil)
	optReq.Header.Set("Origin", "http://localhost:3000")
	optReq.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
	optRr := httptest.NewRecorder()
	server.router.ServeHTTP(optRr, optReq)

	if optRr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for OPTIONS, got %d", optRr.Code)
	}
	if origin := optRr.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:3000" {
		t.Errorf("expected Allow-Origin 'http://localhost:3000', got '%s'", origin)
	}
	if headers := optRr.Header().Get("Access-Control-Allow-Headers"); headers != "Content-Type, Authorization" {
		t.Errorf("expected Allow-Headers echo, got '%s'", headers)
	}

	// 3. GET mock route with CORS enabled
	getReq := httptest.NewRequest(http.MethodGet, "/petstore/pets", nil)
	getReq.Header.Set("Origin", "http://localhost:3000")
	getRr := httptest.NewRecorder()
	server.router.ServeHTTP(getRr, getReq)

	if getRr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", getRr.Code)
	}
	if origin := getRr.Header().Get("Access-Control-Allow-Origin"); origin != "http://localhost:3000" {
		t.Errorf("expected Allow-Origin 'http://localhost:3000', got '%s'", origin)
	}

	// 4. Disable CORS dynamically
	server.SetCORS(false)
	if server.CORSEnabled() {
		t.Fatalf("expected CORS to be disabled")
	}

	// 5. GET mock route with CORS disabled
	noCorsReq := httptest.NewRequest(http.MethodGet, "/petstore/pets", nil)
	noCorsReq.Header.Set("Origin", "http://localhost:3000")
	noCorsRr := httptest.NewRecorder()
	server.router.ServeHTTP(noCorsRr, noCorsReq)

	if noCorsRr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", noCorsRr.Code)
	}
	if origin := noCorsRr.Header().Get("Access-Control-Allow-Origin"); origin != "" {
		t.Errorf("expected empty Allow-Origin when CORS disabled, got '%s'", origin)
	}
}
