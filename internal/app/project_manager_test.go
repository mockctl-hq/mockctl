package app

import (
	"context"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"testing"
)

func TestProjectManager_CreateAndListAndGetAndDelete(t *testing.T) {
	mockStore := NewMockSystemStore()
	gateway := runtime.NewProjectGateway()

	app := &App{
		systemStore: mockStore,
		logger:      &MockLogger{},
		gateway:     gateway,
	}

	ctx := context.Background()

	// Create Project (requires a valid OpenAPI spec bytes)
	specBytes := []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
`)

	err := app.CreateProject(ctx, "test-proj", specBytes)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Verify SystemStore has it
	if _, ok := mockStore.Projects["test-proj"]; !ok {
		t.Error("Project was not saved to SystemStore")
	}

	// Verify Gateway has it
	if engine := gateway.GetEngine("test-proj"); engine == nil {
		t.Error("Project was not mounted in ProjectGateway")
	}

	// List Projects from ProjectManager
	listed, err := app.ListProjects(ctx)
	if err != nil {
		t.Fatalf("Failed to list projects: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("Expected 1 project, got %d", len(listed))
	}

	// Delete Project
	err = app.DeleteProject(ctx, "test-proj")
	if err != nil {
		t.Fatalf("Failed to delete project: %v", err)
	}

	// Verify deleted from store
	if len(mockStore.Projects) != 0 {
		t.Error("Project was not deleted from SystemStore")
	}

	// Verify unmounted
	if gateway.GetEngine("test-proj") != nil {
		t.Error("Project was not unmounted from gateway")
	}
}
