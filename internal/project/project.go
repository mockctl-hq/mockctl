package project

// WorkspaceContext holds project settings, seeds, and overrides.
type WorkspaceContext struct {
	Seed      int64
	Overrides map[string]map[int]any // Keyed by path, then status code
}
