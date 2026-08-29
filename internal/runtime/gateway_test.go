package runtime

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjectGateway_SetAndGetEngine(t *testing.T) {
	g := NewProjectGateway()

	engine := &RuntimeEngine{}
	g.SetEngine("test-proj", engine)

	retrieved := g.GetEngine("test-proj")
	if retrieved != engine {
		t.Error("Expected to retrieve the same engine instance")
	}
}

func TestProjectGateway_RemoveEngine(t *testing.T) {
	g := NewProjectGateway()

	engine := &RuntimeEngine{}
	g.SetEngine("test-proj", engine)
	g.RemoveEngine("test-proj")

	retrieved := g.GetEngine("test-proj")
	if retrieved != nil {
		t.Error("Expected engine to be nil after removal")
	}
}

func TestProjectGateway_ServeHTTP_NotFound(t *testing.T) {
	g := NewProjectGateway()

	req := httptest.NewRequest(http.MethodGet, "/unknown-proj/some/path", nil)
	w := httptest.NewRecorder()
	
	g.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %d", w.Code)
	}
}
