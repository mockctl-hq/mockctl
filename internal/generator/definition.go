package generator

// RuntimeDefinition is the structured blueprint containing compiled routes,
// methods, and response templates. It is strictly read-only after generation.
type RuntimeDefinition struct {
	Endpoints map[string]EndpointHandler
}

// EndpointHandler specifies the routing and response constraints for a specific mock route.
type EndpointHandler struct {
	Method      string
	Path        string
	PathParams  []string                 // e.g., ["id"] for /users/{id}
	QueryParams []string                 // Expected query parameters
	Responses   map[int]ResponseTemplate // Maps HTTP status (e.g. 200, 404) to its template
}

// ResponseTemplate holds the payload structure and custom headers for a specific HTTP status.
type ResponseTemplate struct {
	Headers map[string]string
	Body    map[string]any
}
