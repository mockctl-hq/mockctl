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
	os.WriteFile(path, []byte(content), 0644)

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
	f.Truncate(maxFileSize + 1)
	f.Close()

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
	os.WriteFile(path, []byte(content), 0644)

	parser := NewOpenAPIParser()
	_, err := parser.ParseFile(context.Background(), path)
	if err == nil {
		t.Fatalf("expected error due to circular reference, got nil")
	}
}
