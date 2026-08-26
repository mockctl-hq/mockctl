package runtime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mockctl-hq/mockctl/internal/generator"
	"github.com/mockctl-hq/mockctl/internal/shared"
)

// FakeLogger
type fakeLogger struct{}

func (f *fakeLogger) Info(msg string, args ...any)             {}
func (f *fakeLogger) Warn(msg string, args ...any)             {}
func (f *fakeLogger) Error(msg string, err error, args ...any) {}
func (f *fakeLogger) Debug(msg string, args ...any)            {}

// FakeClock
type fakeClock struct{}

func (f *fakeClock) Now() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

// FakeValueProvider
type fakeValueProvider struct{}

func (f *fakeValueProvider) GenerateUUID(ctx context.Context) string                  { return "test-uuid-1234" }
func (f *fakeValueProvider) GenerateString(ctx context.Context, format string) string { return "test" }
func (f *fakeValueProvider) GenerateInt(ctx context.Context, min, max int) int        { return 42 }
func (f *fakeValueProvider) GenerateBoolean(ctx context.Context) bool                 { return true }

// FakeChaosEvaluator
type fakeChaosEvaluator struct {
	status int
	err    error
}

func (f *fakeChaosEvaluator) Evaluate(ctx context.Context) (int, error) {
	return f.status, f.err
}

// FakeStateStore
type fakeStateStore struct {
	data map[string]map[string]map[string]any
}

func newFakeStateStore() *fakeStateStore {
	return &fakeStateStore{
		data: make(map[string]map[string]map[string]any),
	}
}
func (f *fakeStateStore) Insert(ctx context.Context, col string, id string, payload map[string]any) error {
	if f.data[col] == nil {
		f.data[col] = make(map[string]map[string]any)
	}
	f.data[col][id] = payload
	return nil
}
func (f *fakeStateStore) Get(ctx context.Context, col string, id string) (map[string]any, error) {
	if f.data[col] != nil && f.data[col][id] != nil {
		return f.data[col][id], nil
	}
	return nil, shared.ErrNotFound
}
func (f *fakeStateStore) List(ctx context.Context, col string) ([]map[string]any, error) {
	var res []map[string]any
	for _, v := range f.data[col] {
		res = append(res, v)
	}
	return res, nil
}
func (f *fakeStateStore) Update(ctx context.Context, col string, id string, payload map[string]any) error {
	return f.Insert(ctx, col, id, payload)
}
func (f *fakeStateStore) Patch(ctx context.Context, col string, id string, partial map[string]any) error {
	return f.Insert(ctx, col, id, partial) // simplified
}
func (f *fakeStateStore) Delete(ctx context.Context, col string, id string) error {
	if f.data[col] != nil {
		delete(f.data[col], id)
		return nil
	}
	return shared.ErrNotFound
}
func (f *fakeStateStore) Reset(ctx context.Context) error { return nil }

func TestRuntimeEngine_HandleGetFallbackTemplate(t *testing.T) {
	def := &generator.RuntimeDefinition{
		Endpoints: map[string]generator.EndpointHandler{
			"get_users_123": {
				Method: "GET",
				Path:   "/users/{id}",
				Responses: map[int]generator.ResponseTemplate{
					http.StatusOK: {
						Headers: map[string]string{"X-Test": "Mockctl"},
						Body:    map[string]any{"id": "template-id", "name": "Template User"},
					},
				},
			},
		},
	}

	engine := NewRuntimeEngine(&fakeLogger{}, def, newFakeStateStore(), &fakeChaosEvaluator{}, &fakeValueProvider{}, &fakeClock{})

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	rr := httptest.NewRecorder()

	engine.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v but got %v", http.StatusOK, status)
	}

	if rr.Header().Get("X-Test") != "Mockctl" {
		t.Errorf("expected custom header X-Test: Mockctl")
	}

	var resBody map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resBody); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resBody["name"] != "Template User" {
		t.Errorf("expected fallback template body, got %v", resBody)
	}
}
