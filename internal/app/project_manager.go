package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mockctl-hq/mockctl/internal/data"
	"github.com/mockctl-hq/mockctl/internal/generator"
	"github.com/mockctl-hq/mockctl/internal/project"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/shared"
	"github.com/mockctl-hq/mockctl/internal/spec"
)

// CreateProject handles incoming project payloads from the Admin API.
func (a *App) CreateProject(ctx context.Context, name string, payload []byte) error {
	// 1. Parse OpenAPI
	parser := spec.NewOpenAPIParser()
	specModel, err := parser.ParseData(ctx, payload)
	if err != nil {
		return fmt.Errorf("invalid openapi spec: %w", err)
	}

	// 2. Create Project Data Model
	proj := project.Project{
		Name:            name,
		OpenAPISpec:     string(payload),
		CustomEndpoints: []project.EndpointHandler{},
		Workspace:       project.WorkspaceContext{Overrides: make(map[string]map[int]any)},
	}

	// 3. Compile Engine (Compile First, Save Later)
	engine, err := a.compileEngine(ctx, proj, specModel)
	if err != nil {
		return fmt.Errorf("failed to compile runtime engine: %w", err)
	}

	// 4. Save to BBolt
	projBytes, err := json.Marshal(proj)
	if err != nil {
		return err
	}
	if err := a.systemStore.SaveProject(ctx, name, projBytes); err != nil {
		return err
	}

	// 5. Hot Swap
	a.gateway.SetEngine(name, engine)
	return nil
}

func (a *App) compileEngine(ctx context.Context, proj project.Project, specModel *spec.SpecModel) (*runtime.RuntimeEngine, error) {
	// Setup dependencies
	vp := data.NewFakeValueProvider()

	// Create chaos evaluator (stub for now, needs real one if available)
	var chaos shared.ChaosEvaluator // TODO: inject fake or real

	gen := generator.NewMockGenerator(a.logger, vp)

	def, err := gen.Generate(ctx, specModel, &proj.Workspace)
	if err != nil {
		return nil, err
	}

	clock := shared.NewRealClock()

	engine := runtime.NewRuntimeEngine(a.logger, def, a.stateStore, chaos, vp, clock)
	return engine, nil
}

func (a *App) ListProjects(ctx context.Context) (map[string]map[string]any, error) {
	projects, err := a.systemStore.ListProjects(ctx)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]any)
	for name := range projects {
		result[name] = map[string]any{"status": "active"} // Simplistic summary
	}
	return result, nil
}

func (a *App) DeleteProject(ctx context.Context, name string) error {
	return a.systemStore.DeleteProject(ctx, name)
}

func (a *App) AddEndpoint(ctx context.Context, projectName string, endpointJSON []byte) error {
	var endpoint project.EndpointHandler
	if err := json.Unmarshal(endpointJSON, &endpoint); err != nil {
		return fmt.Errorf("invalid endpoint json: %w", err)
	}

	projData, err := a.systemStore.GetProject(ctx, projectName)
	if err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	var proj project.Project
	if err := json.Unmarshal(projData, &proj); err != nil {
		return fmt.Errorf("failed to parse existing project: %w", err)
	}

	proj.CustomEndpoints = append(proj.CustomEndpoints, endpoint)

	// Recompile Engine
	parser := spec.NewOpenAPIParser()
	specModel, err := parser.ParseData(ctx, []byte(proj.OpenAPISpec))
	if err != nil {
		return fmt.Errorf("failed to parse stored openapi spec: %w", err)
	}

	engine, err := a.compileEngine(ctx, proj, specModel)
	if err != nil {
		return fmt.Errorf("failed to recompile engine: %w", err)
	}

	// Save
	newProjBytes, err := json.Marshal(proj)
	if err != nil {
		return err
	}
	if err := a.systemStore.SaveProject(ctx, projectName, newProjBytes); err != nil {
		return err
	}

	// Hot Swap
	a.gateway.SetEngine(projectName, engine)
	return nil
}
func (a *App) SetOverrides(ctx context.Context, projectName string, overridesJSON []byte) error {
	var overrides map[string]map[int]any
	if err := json.Unmarshal(overridesJSON, &overrides); err != nil {
		return fmt.Errorf("invalid overrides json: %w", err)
	}

	projData, err := a.systemStore.GetProject(ctx, projectName)
	if err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	var proj project.Project
	if err := json.Unmarshal(projData, &proj); err != nil {
		return fmt.Errorf("failed to parse existing project: %w", err)
	}

	proj.Workspace.Overrides = overrides

	// Recompile Engine
	parser := spec.NewOpenAPIParser()
	specModel, err := parser.ParseData(ctx, []byte(proj.OpenAPISpec))
	if err != nil {
		return fmt.Errorf("failed to parse stored openapi spec: %w", err)
	}

	engine, err := a.compileEngine(ctx, proj, specModel)
	if err != nil {
		return fmt.Errorf("failed to recompile engine: %w", err)
	}

	// Save
	newProjBytes, err := json.Marshal(proj)
	if err != nil {
		return err
	}
	if err := a.systemStore.SaveProject(ctx, projectName, newProjBytes); err != nil {
		return err
	}

	// Hot Swap
	a.gateway.SetEngine(projectName, engine)
	return nil
}
func (a *App) ResetState(ctx context.Context, projectName string) error {
	// Task 3.2: Flushes live CRUD data. Since StateStore doesn't yet support prefix deletion, we flush entirely for now.
	// In the future, StateStore should implement ResetProject(projectName).
	return a.stateStore.Reset(ctx)
}

func (a *App) UpdateChaos(ctx context.Context, projectName string, errorRate int, latencyMs int) error {
	engine := a.gateway.GetEngine(projectName)
	if engine == nil {
		return fmt.Errorf("project %s not currently running", projectName)
	}

	if engine.Chaos() != nil {
		engine.Chaos().UpdateConfig(ctx, errorRate, latencyMs)
	}
	return nil
}

func (a *App) HydrateProjects(ctx context.Context) error {
	projectsBytes, err := a.systemStore.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list projects for hydration: %w", err)
	}

	for name, data := range projectsBytes {
		var proj project.Project
		if err := json.Unmarshal(data, &proj); err != nil {
			a.logger.Error("Failed to unmarshal project during hydration", err, "project", name)
			continue
		}

		parser := spec.NewOpenAPIParser()
		specModel, err := parser.ParseData(ctx, []byte(proj.OpenAPISpec))
		if err != nil {
			a.logger.Error("Failed to parse openapi spec during hydration", err, "project", name)
			continue
		}

		engine, err := a.compileEngine(ctx, proj, specModel)
		if err != nil {
			a.logger.Error("Failed to compile engine during hydration", err, "project", name)
			continue
		}

		a.gateway.SetEngine(name, engine)
		a.logger.Info("Hydrated project successfully", "project", name)
	}
	return nil
}
