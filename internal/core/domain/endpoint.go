package domain

// ResponseTemplate defines the expected output for a specific mock request.
type ResponseTemplate struct {
	Headers map[string]string
	Body    map[string]any
}

// EndpointHandler defines a single route and its available responses.
type EndpointHandler struct {
	Method      string
	Path        string
	PathParams  []string
	QueryParams map[string]string
	Responses   map[int]ResponseTemplate // Keyed by HTTP Status Code
}
