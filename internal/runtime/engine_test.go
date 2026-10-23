package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

func (f *fakeClock) Now() time.Time                         { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
func (f *fakeClock) After(d time.Duration) <-chan time.Time { return time.After(d) }

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
func (f *fakeChaosEvaluator) UpdateConfig(ctx context.Context, errorRate int, latencyMs int) {}
func (f *fakeChaosEvaluator) UpdateConfigWithStatus(ctx context.Context, enabled bool, errorRate int, latencyMs int, statusCode int) {
}
func (f *fakeChaosEvaluator) Config() (bool, int, int, int) { return false, 0, 0, 500 }

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

	engine := NewRuntimeEngine(&fakeLogger{}, def, newFakeStateStore(), &fakeChaosEvaluator{}, &fakeValueProvider{}, &fakeClock{}, nil, "test-project")

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

type mockValidator struct {
	shouldFail bool
	errMsg     string
}

func (m *mockValidator) ValidateJSON(data []byte) error {
	if m.shouldFail {
		return errors.New(m.errMsg)
	}
	return nil
}

func TestRuntimeEngine_RequestBody_Validation(t *testing.T) {
	validator := &mockValidator{
		shouldFail: true,
		errMsg:     "property 'email' is required",
	}

	def := &generator.RuntimeDefinition{
		Endpoints: map[string]generator.EndpointHandler{
			"post_users": {
				Method:           "POST",
				Path:             "/users",
				RequestValidator: validator,
				Responses: map[int]generator.ResponseTemplate{
					http.StatusCreated: {
						Headers: map[string]string{"Content-Type": "application/json"},
						Body:    map[string]any{"id": "user-1", "status": "created"},
					},
				},
			},
		},
	}

	state := newFakeStateStore()
	engine := NewRuntimeEngine(&fakeLogger{}, def, state, nil, &fakeValueProvider{}, &fakeClock{}, nil, "test-proj")

	// 1. Strict Validation is ON by default -> Should fail with 400 Bad Request
	if !engine.StrictValidationEnabled() {
		t.Fatalf("expected strict validation to be enabled by default")
	}

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name": "Alice"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request under strict validation, got %d", rr.Code)
	}

	var errResp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error JSON: %v", err)
	}
	errObj, ok := errResp["error"].(map[string]any)
	if !ok || errObj["code"] != "SCHEMA_VALIDATION_FAILED" {
		t.Errorf("expected SCHEMA_VALIDATION_FAILED error code, got %v", errResp)
	}

	// 2. Disable Strict Validation (Permissive Mode) -> Should allow request and return 201 Created
	engine.SetStrictValidation(false)
	if engine.StrictValidationEnabled() {
		t.Fatalf("expected strict validation to be disabled")
	}

	req2 := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name": "Alice"}`))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	engine.ServeHTTP(rr2, req2)

	if rr2.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created under permissive validation, got %d", rr2.Code)
	}

	// 3. Re-enable Strict Validation and test valid payload -> Should return 201 Created
	engine.SetStrictValidation(true)
	validator.shouldFail = false

	req3 := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name": "Alice", "email": "alice@example.com"}`))
	req3.Header.Set("Content-Type", "application/json")
	rr3 := httptest.NewRecorder()
	engine.ServeHTTP(rr3, req3)

	if rr3.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for valid payload, got %d", rr3.Code)
	}
}

func TestProjectGateway_StrictValidation(t *testing.T) {
	gateway := NewProjectGateway()
	if !gateway.StrictValidationEnabled() {
		t.Fatalf("expected gateway strict validation to be true by default")
	}

	engine := NewRuntimeEngine(&fakeLogger{}, nil, newFakeStateStore(), nil, &fakeValueProvider{}, &fakeClock{}, nil, "p1")
	gateway.SetEngine("p1", engine)

	if !engine.StrictValidationEnabled() {
		t.Fatalf("expected engine to inherit gateway's strict validation setting")
	}

	// Toggle gateway strict validation to false
	gateway.SetStrictValidation(false)
	if gateway.StrictValidationEnabled() {
		t.Fatalf("expected gateway strict validation to be false")
	}
	if engine.StrictValidationEnabled() {
		t.Fatalf("expected engine strict validation to be updated to false")
	}

	// Add new engine, should receive false
	engine2 := NewRuntimeEngine(&fakeLogger{}, nil, newFakeStateStore(), nil, &fakeValueProvider{}, &fakeClock{}, nil, "p2")
	gateway.SetEngine("p2", engine2)
	if engine2.StrictValidationEnabled() {
		t.Fatalf("expected new engine to receive gateway's false strict validation")
	}
}

func TestRuntimeEngine_ChaosEvaluation_CustomStatus(t *testing.T) {
	ctx := context.Background()
	clock := shared.NewRealClock()
	chaos := NewStandardChaosEvaluator(clock)
	// Enable chaos with 100% error rate and custom status 429
	chaos.UpdateConfigWithStatus(ctx, true, 100, 0, 429)

	def := &generator.RuntimeDefinition{
		Endpoints: map[string]generator.EndpointHandler{
			"GET /status": {
				Method: http.MethodGet,
				Path:   "/status",
				Responses: map[int]generator.ResponseTemplate{
					http.StatusOK: {Body: map[string]any{"ok": true}},
				},
			},
		},
	}

	engine := NewRuntimeEngine(&fakeLogger{}, def, newFakeStateStore(), chaos, &fakeValueProvider{}, clock, nil, "test-project")

	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	rr := httptest.NewRecorder()

	engine.ServeHTTP(rr, req)

	if rr.Code != 429 {
		t.Fatalf("expected status 429, got %d", rr.Code)
	}

	var resBody map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resBody["success"] != false {
		t.Fatalf("expected success: false, got %v", resBody["success"])
	}
	errMap, ok := resBody["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object, got %v", resBody["error"])
	}
	if errMap["code"] != "CHAOS_INJECTED" {
		t.Fatalf("expected code CHAOS_INJECTED, got %v", errMap["code"])
	}
	if errMap["status"] != float64(429) {
		t.Fatalf("expected status 429, got %v", errMap["status"])
	}
}

func TestRuntimeEngine_StaticOverride(t *testing.T) {
	clock := shared.NewRealClock()
	chaos := NewStandardChaosEvaluator(clock)

	overridePayload := map[string]any{
		"message": "custom static override payload",
		"version": float64(2),
	}

	def := &generator.RuntimeDefinition{
		Endpoints: map[string]generator.EndpointHandler{
			"GET /users": {
				Method: http.MethodGet,
				Path:   "/users",
				Responses: map[int]generator.ResponseTemplate{
					http.StatusOK: {
						Headers:    map[string]string{"Content-Type": "application/json"},
						Body:       overridePayload,
						IsOverride: true,
					},
				},
			},
		},
	}

	stateStore := newFakeStateStore()
	engine := NewRuntimeEngine(&fakeLogger{}, def, stateStore, chaos, &fakeValueProvider{}, clock, nil, "test-project")

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	engine.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resBody map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resBody["message"] != "custom static override payload" {
		t.Fatalf("expected override message, got %v", resBody["message"])
	}
	if resBody["version"] != float64(2) {
		t.Fatalf("expected override version 2, got %v", resBody["version"])
	}
}
