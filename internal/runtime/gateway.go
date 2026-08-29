package runtime

import (
	"context"
	"net/http"
	"strings"
	"sync"
)

// ProjectManager defines the downward dependency for the Admin API to invoke the App Core.
type ProjectManager interface {
	CreateProject(ctx context.Context, name string, fileStream []byte) error
	ListProjects(ctx context.Context) (map[string]map[string]any, error)
	DeleteProject(ctx context.Context, name string) error

	AddEndpoint(ctx context.Context, projectName string, endpointJSON []byte) error
	SetOverrides(ctx context.Context, projectName string, overridesJSON []byte) error
	ResetState(ctx context.Context, projectName string) error
	UpdateChaos(ctx context.Context, projectName string, errorRate int, latencyMs int) error
}

// ProjectGateway handles dynamic routing to multiple projects.
type ProjectGateway struct {
	mu      sync.RWMutex
	engines map[string]*RuntimeEngine
}

func NewProjectGateway() *ProjectGateway {
	return &ProjectGateway{
		engines: make(map[string]*RuntimeEngine),
	}
}

func (g *ProjectGateway) SetEngine(projectName string, engine *RuntimeEngine) {
	g.mu.Lock()
	defer g.mu.Unlock()
	// Task 2.3: Terminate old engine if it exists
	if old, exists := g.engines[projectName]; exists {
		// Context cancellation can be implemented here if old engine stores a cancel func
		_ = old
	}
	g.engines[projectName] = engine
}

func (g *ProjectGateway) RemoveEngine(projectName string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.engines, projectName)
}

func (g *ProjectGateway) GetEngine(projectName string) *RuntimeEngine {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.engines[projectName]
}

// ServeHTTP delegates the request to the appropriate project's engine.
func (g *ProjectGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/"), "/", 2)
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.NotFound(w, r)
		return
	}

	projectName := pathParts[0]

	// Task 2.3: RWMutex Lock contention fix (Read lock only to copy pointer)
	g.mu.RLock()
	engine, exists := g.engines[projectName]
	g.mu.RUnlock()

	if !exists {
		http.NotFound(w, r)
		return
	}

	// Strip the /{projectName} prefix before sending to the engine router
	http.StripPrefix("/"+projectName, engine).ServeHTTP(w, r)
}
