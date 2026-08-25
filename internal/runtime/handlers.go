package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mockctl-hq/mockctl/internal/core/ports"
	"github.com/mockctl-hq/mockctl/internal/generator"
)

// buildFlatPath extracts variables from the Go 1.22 mux and maps them into the path
func buildFlatPath(r *http.Request, endpoint generator.EndpointHandler) string {
	// The path from definition might be /users/{id}/posts/{post_id}
	// We need to replace these with the actual values.
	// But in a simpler way, the actual request path represents the flat collection/ID hierarchy.
	// e.g. /users/123/posts/456 -> collection: users/123/posts, id: 456
	// For now, we can just use the raw path, minus the leading slash.
	return strings.TrimPrefix(r.URL.Path, "/")
}

// extractCollectionAndID splits the flat path.
func extractCollectionAndID(flatPath string) (string, string) {
	parts := strings.Split(flatPath, "/")
	if len(parts) == 0 || flatPath == "" {
		return "", ""
	}
	if len(parts)%2 == 0 {
		// Even number of parts: e.g. users/123 -> Collection: "users", ID: "123"
		return strings.Join(parts[:len(parts)-1], "/"), parts[len(parts)-1]
	}
	// Odd number of parts: e.g. users -> Collection: "users", ID: ""
	return flatPath, ""
}

// formatResponse sets the custom headers and JSON body from the StateStore or Template.
func (e *RuntimeEngine) formatResponse(w http.ResponseWriter, status int, headers map[string]string, body any) {
	w.Header().Set("Content-Type", "application/json")
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
	if body != nil {
		_ = json.NewEncoder(w).Encode(body)
	}
}

func (e *RuntimeEngine) handleGet(ctx context.Context, w http.ResponseWriter, r *http.Request, endpoint generator.EndpointHandler) {
	flatPath := buildFlatPath(r, endpoint)
	collection, id := extractCollectionAndID(flatPath)

	var result any
	var err error

	if id == "" {
		// List collection
		result, err = e.state.List(ctx, collection)
		// TODO: Filter list by query parameters
	} else {
		// Get single item
		result, err = e.state.Get(ctx, collection, id)
	}

	if err != nil {
		if err == ports.ErrNotFound {
			// Fallback to generating template data
			template, ok := endpoint.Responses[http.StatusOK]
			if !ok {
				e.sendError(w, "NOT_FOUND", "Resource not found and no template exists", http.StatusNotFound)
				return
			}
			// TODO: Use e.valueProvider to generate dynamic values in the template body
			e.formatResponse(w, http.StatusOK, template.Headers, template.Body)
			return
		}
		e.sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	// Found in StateStore
	e.formatResponse(w, http.StatusOK, nil, result)
}

func (e *RuntimeEngine) handlePost(ctx context.Context, w http.ResponseWriter, r *http.Request, endpoint generator.EndpointHandler) {
	flatPath := buildFlatPath(r, endpoint)
	collection, _ := extractCollectionAndID(flatPath)

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		e.sendError(w, "INVALID_JSON", "Failed to parse JSON payload", http.StatusBadRequest)
		return
	}

	// Ensure ID exists
	idStr := ""
	if idVal, ok := payload["id"]; ok {
		// Simplified conversion
		idStr, _ = idVal.(string)
	}
	if idStr == "" {
		idStr = e.valueProvider.GenerateUUID(ctx)
		payload["id"] = idStr
	}

	// TODO: Inject timestamps if schema expects them

	if err := e.state.Insert(ctx, collection, idStr, payload); err != nil {
		if err == ports.ErrLimitReached {
			e.sendError(w, "RATE_LIMIT", err.Error(), http.StatusTooManyRequests)
			return
		}
		e.sendError(w, "INTERNAL_ERROR", err.Error(), http.StatusInternalServerError)
		return
	}

	template := endpoint.Responses[http.StatusCreated]
	e.formatResponse(w, http.StatusCreated, template.Headers, payload)
}

func (e *RuntimeEngine) handlePut(ctx context.Context, w http.ResponseWriter, r *http.Request, endpoint generator.EndpointHandler) {
	flatPath := buildFlatPath(r, endpoint)
	collection, id := extractCollectionAndID(flatPath)

	if id == "" {
		e.sendError(w, "METHOD_NOT_ALLOWED", "Cannot PUT to a collection", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		e.sendError(w, "INVALID_JSON", "Failed to parse JSON payload", http.StatusBadRequest)
		return
	}

	payload["id"] = id // Ensure ID matches URL

	if err := e.state.Update(ctx, collection, id, payload); err != nil {
		status := http.StatusInternalServerError
		if err == ports.ErrNotFound {
			status = http.StatusNotFound
		}
		e.sendError(w, "UPDATE_FAILED", err.Error(), status)
		return
	}

	e.formatResponse(w, http.StatusOK, nil, payload)
}

func (e *RuntimeEngine) handlePatch(ctx context.Context, w http.ResponseWriter, r *http.Request, endpoint generator.EndpointHandler) {
	flatPath := buildFlatPath(r, endpoint)
	collection, id := extractCollectionAndID(flatPath)

	if id == "" {
		e.sendError(w, "METHOD_NOT_ALLOWED", "Cannot PATCH a collection", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		e.sendError(w, "INVALID_JSON", "Failed to parse JSON payload", http.StatusBadRequest)
		return
	}

	if err := e.state.Patch(ctx, collection, id, payload); err != nil {
		status := http.StatusInternalServerError
		if err == ports.ErrNotFound {
			status = http.StatusNotFound
		}
		e.sendError(w, "PATCH_FAILED", err.Error(), status)
		return
	}

	updated, _ := e.state.Get(ctx, collection, id)
	e.formatResponse(w, http.StatusOK, nil, updated)
}

func (e *RuntimeEngine) handleDelete(ctx context.Context, w http.ResponseWriter, r *http.Request, endpoint generator.EndpointHandler) {
	flatPath := buildFlatPath(r, endpoint)
	collection, id := extractCollectionAndID(flatPath)

	if id == "" {
		e.sendError(w, "METHOD_NOT_ALLOWED", "Cannot DELETE a collection", http.StatusMethodNotAllowed)
		return
	}

	if err := e.state.Delete(ctx, collection, id); err != nil {
		status := http.StatusInternalServerError
		if err == ports.ErrNotFound {
			status = http.StatusNotFound
		}
		e.sendError(w, "DELETE_FAILED", err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
