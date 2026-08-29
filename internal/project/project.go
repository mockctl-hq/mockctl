package project

// WorkspaceContext holds project settings, seeds, and overrides.
type WorkspaceContext struct {
	Seed      int64
	Overrides map[string]map[int]any // Keyed by path, then status code
}

// Project represents the hybrid data model for a Mock:ctl project.
type Project struct {
	Name            string            `json:"name"`
	OpenAPISpec     string            `json:"openapi_spec"`     // Raw YAML/JSON string
	CustomEndpoints []EndpointHandler `json:"custom_endpoints"` // Manually added endpoints
	Workspace       WorkspaceContext  `json:"workspace"`
}
