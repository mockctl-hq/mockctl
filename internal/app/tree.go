package app

import (
	"context"
	"os"
	"path/filepath"
	"sort"
)

// EndpointSummaryDTO provides a presentation-ready summary of an endpoint.
type EndpointSummaryDTO struct {
	Path        string
	HasOverride bool
	HasChaos    bool
}

// ProjectSummaryDTO provides a presentation-ready tree structure of a project and its endpoints.
type ProjectSummaryDTO struct {
	Name         string
	SpecFilename string
	Endpoints    []EndpointSummaryDTO
}

// GetRequestsLogContent retrieves the contents of the requests.log file through the FileSystem abstraction.
func (a *App) GetRequestsLogContent() ([]byte, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	logPath := filepath.Join(userHome, ".mockctl", "requests.log")
	return a.fileSystem.ReadFile(logPath)
}

// GetProjectTree returns a hierarchical summary of all active projects and their endpoints for presentation clients.
func (a *App) GetProjectTree(ctx context.Context) ([]ProjectSummaryDTO, error) {
	projects, err := a.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	if a.gateway == nil {
		return nil, nil
	}

	engines := a.gateway.ListEngines()

	var projNames []string
	for name := range engines {
		projNames = append(projNames, name)
	}
	sort.Strings(projNames)

	var summaries []ProjectSummaryDTO
	for _, name := range projNames {
		engine := engines[name]
		projData := projects[name]
		fileName := projData.SpecFilename
		if fileName == "" {
			fileName = "api.yaml"
		}

		projSummary := ProjectSummaryDTO{
			Name:         name,
			SpecFilename: fileName,
			Endpoints:    make([]EndpointSummaryDTO, 0),
		}

		def := engine.Definition()
		var epPaths []string
		for path := range def.Endpoints {
			epPaths = append(epPaths, path)
		}
		sort.Strings(epPaths)

		for _, path := range epPaths {
			hasOverride := false
			if epHandler, ok := def.Endpoints[path]; ok {
				for _, resp := range epHandler.Responses {
					if resp.IsOverride {
						hasOverride = true
						break
					}
				}
			}

			hasChaos := false
			if engine.Chaos() != nil {
				enabled, _, _, _ := engine.Chaos().Config()
				if enabled {
					hasChaos = true
				}
			}

			projSummary.Endpoints = append(projSummary.Endpoints, EndpointSummaryDTO{
				Path:        path,
				HasOverride: hasOverride,
				HasChaos:    hasChaos,
			})
		}
		summaries = append(summaries, projSummary)
	}

	return summaries, nil
}
