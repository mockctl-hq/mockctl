package spec

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile_Success(t *testing.T) {
	// Create a dummy valid openapi spec
	content := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
servers:
  - url: /api/v1
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    format: date-time
                    example: "2023-01-01T12:00:00Z"
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "openapi.yaml")
	_ = os.WriteFile(path, []byte(content), 0644)

	parser := NewOpenAPIParser()
	model, err := parser.ParseFile(context.Background(), path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if model.Title != "Test API" {
		t.Errorf("expected title 'Test API', got %s", model.Title)
	}

	if len(model.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(model.Routes))
	}

	route := model.Routes[0]
	if route.Path != "/api/v1/users" {
		t.Errorf("expected path '/api/v1/users', got %s", route.Path)
	}

	schema := route.SchemaRef.(*NormalizedSchema)
	if schema.Properties["id"].Example != "2023-01-01T12:00:00Z" {
		t.Errorf("expected example to be set, got %v", schema.Properties["id"].Example)
	}
}

func TestParseFile_Security_FileSizeLimit(t *testing.T) {
	// Create a file larger than 50MB
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "large.yaml")
	f, _ := os.Create(path)
	_ = f.Truncate(maxFileSize + 1)
	_ = f.Close()

	parser := NewOpenAPIParser()
	_, err := parser.ParseFile(context.Background(), path)
	if err == nil || err.Error() != "openapi file exceeds maximum allowed size of 50MB" {
		t.Errorf("expected file size limit error, got %v", err)
	}
}

func TestParseFile_Security_CircularReference(t *testing.T) {
	// Create a schema with a circular reference
	content := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        manager:
          $ref: '#/components/schemas/User'
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "circular.yaml")
	_ = os.WriteFile(path, []byte(content), 0644)

	parser := NewOpenAPIParser()
	_, err := parser.ParseFile(context.Background(), path)
	if err == nil {
		t.Fatalf("expected error due to circular reference, got nil")
	}
}

func TestParseData_RequestBody_Validation(t *testing.T) {
	content := `
openapi: 3.0.0
info:
  title: Petstore
  version: 1.0.0
paths:
  /pets:
    post:
      summary: Create a pet
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
                - age
              properties:
                name:
                  type: string
                age:
                  type: integer
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: integer
`
	parser := NewOpenAPIParser()
	model, err := parser.ParseData(context.Background(), []byte(content))
	if err != nil {
		t.Fatalf("expected no error parsing data: %v", err)
	}

	if len(model.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(model.Routes))
	}

	route := model.Routes[0]
	if route.Method != "POST" || route.Path != "/pets" {
		t.Fatalf("unexpected route: %s %s", route.Method, route.Path)
	}

	if route.RequestValidator == nil {
		t.Fatalf("expected RequestValidator to be non-nil")
	}

	if route.RequestBodySchema == nil {
		t.Fatalf("expected RequestBodySchema to be non-nil")
	}
	if route.RequestBodySchema.Type != "object" {
		t.Errorf("expected RequestBodySchema.Type == 'object', got '%s'", route.RequestBodySchema.Type)
	}

	// 1. Test Valid Payload
	validJSON := []byte(`{"name": "Rex", "age": 5}`)
	if err := route.RequestValidator.ValidateJSON(validJSON); err != nil {
		t.Errorf("expected valid JSON to pass validation, got: %v", err)
	}

	// 2. Test Missing Required Field ('age')
	missingFieldJSON := []byte(`{"name": "Rex"}`)
	if err := route.RequestValidator.ValidateJSON(missingFieldJSON); err == nil {
		t.Errorf("expected error when required field 'age' is missing, got nil")
	}

	// 3. Test Wrong Data Type ('age' as string instead of integer)
	wrongTypeJSON := []byte(`{"name": "Rex", "age": "five"}`)
	if err := route.RequestValidator.ValidateJSON(wrongTypeJSON); err == nil {
		t.Errorf("expected error when field has wrong type, got nil")
	}

	// 4. Test Malformed JSON
	malformedJSON := []byte(`{"name": `)
	if err := route.RequestValidator.ValidateJSON(malformedJSON); err == nil {
		t.Errorf("expected error on malformed JSON, got nil")
	}

	// 5. Test Empty Body when Required
	emptyJSON := []byte(``)
	if err := route.RequestValidator.ValidateJSON(emptyJSON); err == nil {
		t.Errorf("expected error on empty body when required=true, got nil")
	}
}
