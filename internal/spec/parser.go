package spec

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const maxFileSize = 50 * 1024 * 1024 // 50MB

type OpenAPIParser struct{}

func NewOpenAPIParser() *OpenAPIParser {
	return &OpenAPIParser{}
}

func (p *OpenAPIParser) ParseFile(ctx context.Context, path string) (*SpecModel, error) {
	// Security: YAML Bomb Protection
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat openapi file: %w", err)
	}
	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("openapi file exceeds maximum allowed size of 50MB")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return p.ParseData(ctx, data)
}

func (p *OpenAPIParser) ParseData(ctx context.Context, data []byte) (*SpecModel, error) {

	// Security: SSRF & LFI Prevention
	loader := openapi3.NewSwaggerLoader()
	loader.IsExternalRefsAllowed = false

	doc, err := loader.LoadSwaggerFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse openapi data: %w", err)
	}

	// Validate the document
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid openapi specification: %w", err)
	}

	model := &SpecModel{
		Title:   doc.Info.Title,
		Version: doc.Info.Version,
	}

	// Extract base path from global servers
	globalBasePath := ""
	if len(doc.Servers) > 0 {
		globalBasePath = extractBasePath(doc.Servers[0].URL)
	}

	// Extract Routes
	routes, err := p.extractRoutesFromDoc(doc, globalBasePath)
	if err != nil {
		return nil, err
	}
	model.Routes = routes

	return model, nil
}

func (p *OpenAPIParser) ExtractRoutes(model *SpecModel) ([]RouteDef, error) {
	return model.Routes, nil
}

func (p *OpenAPIParser) extractRoutesFromDoc(doc *openapi3.Swagger, globalBasePath string) ([]RouteDef, error) {
	var routes []RouteDef

	for path, pathItem := range doc.Paths {
		// Operation-level base path overrides global
		basePath := globalBasePath
		if len(pathItem.Servers) > 0 {
			basePath = extractBasePath(pathItem.Servers[0].URL)
		}

		for method, operation := range pathItem.Operations() {
			fullPath := basePath + path
			if operation.Servers != nil && len(*operation.Servers) > 0 {
				fullPath = extractBasePath((*operation.Servers)[0].URL) + path
			}

			// Iterate over responses
			for statusStr, responseRef := range operation.Responses {
				status, err := strconv.Atoi(statusStr)
				if err != nil {
					status = 200 // Default if parsing fails (e.g. "default" response)
				}

				if responseRef.Value == nil {
					continue
				}

				// If no content, add a route with nil schema
				if len(responseRef.Value.Content) == 0 {
					routes = append(routes, RouteDef{
						Method:      method,
						Path:        fullPath,
						Status:      status,
						ContentType: "",
						SchemaRef:   nil,
					})
					continue
				}

				// Prioritize application/json
				var schemaRef *openapi3.SchemaRef
				var contentType string
				if mediaType, ok := responseRef.Value.Content["application/json"]; ok {
					schemaRef = mediaType.Schema
					contentType = "application/json"
				} else {
					// Fallback to the first available if not json
					for ct, mediaType := range responseRef.Value.Content {
						schemaRef = mediaType.Schema
						contentType = ct
						break
					}
				}

				// Normalize schema with cycle detection
				visited := make(map[string]bool)
				normalized, err := normalizeSchema(schemaRef, visited, 0)
				if err != nil {
					return nil, fmt.Errorf("failed to normalize schema for %s %s: %w", method, fullPath, err)
				}

				routes = append(routes, RouteDef{
					Method:      method,
					Path:        fullPath,
					Status:      status,
					ContentType: contentType,
					SchemaRef:   normalized,
				})
			}
		}
	}

	return routes, nil
}

func extractBasePath(serverURL string) string {
	// A naive extraction of base path from a server URL.
	// For a real implementation, we should parse the URL and extract the path component.
	if strings.HasPrefix(serverURL, "http://") || strings.HasPrefix(serverURL, "https://") {
		parts := strings.SplitN(serverURL, "/", 4)
		if len(parts) == 4 {
			return "/" + parts[3]
		}
		return ""
	}
	return serverURL
}

func normalizeSchema(ref *openapi3.SchemaRef, visited map[string]bool, depth int) (*NormalizedSchema, error) {
	if ref == nil || ref.Value == nil {
		return nil, nil
	}

	if depth > 10 {
		return nil, fmt.Errorf("schema exceeds maximum recursion depth of 10")
	}

	// Cycle detection for $ref
	if ref.Ref != "" {
		if visited[ref.Ref] {
			return nil, fmt.Errorf("circular reference detected: %s", ref.Ref)
		}
		visited[ref.Ref] = true
		defer func() { visited[ref.Ref] = false }() // unmark when backtracking (optional depending on how we want to track cycles, but usually we just want to prevent infinite loops down a single path)
	}

	schema := ref.Value

	normalized := &NormalizedSchema{
		Type:     schema.Type,
		Format:   schema.Format,
		Required: schema.Required,
		Example:  schema.Example,
		Default:  schema.Default,
	}

	normalized.Enum = append(normalized.Enum, schema.Enum...)

	if schema.Items != nil {
		items, err := normalizeSchema(schema.Items, visited, depth+1)
		if err != nil {
			return nil, err
		}
		normalized.Items = items
	}

	if len(schema.Properties) > 0 {
		normalized.Properties = make(map[string]*NormalizedSchema)
		for k, v := range schema.Properties {
			prop, err := normalizeSchema(v, visited, depth+1)
			if err != nil {
				return nil, err
			}
			normalized.Properties[k] = prop
		}
	}

	if len(schema.AllOf) > 0 {
		for _, s := range schema.AllOf {
			n, err := normalizeSchema(s, visited, depth+1)
			if err != nil {
				return nil, err
			}
			normalized.AllOf = append(normalized.AllOf, n)
		}
	}

	if len(schema.OneOf) > 0 {
		for _, s := range schema.OneOf {
			n, err := normalizeSchema(s, visited, depth+1)
			if err != nil {
				return nil, err
			}
			normalized.OneOf = append(normalized.OneOf, n)
		}
	}

	if len(schema.AnyOf) > 0 {
		for _, s := range schema.AnyOf {
			n, err := normalizeSchema(s, visited, depth+1)
			if err != nil {
				return nil, err
			}
			normalized.AnyOf = append(normalized.AnyOf, n)
		}
	}

	return normalized, nil
}
