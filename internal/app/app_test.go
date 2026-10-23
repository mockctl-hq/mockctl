package app

import (
	"context"
	"testing"

	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/shared"
)

// MockSystemStore implements shared.SystemStore for testing.
type MockSystemStore struct {
	Settings  map[string]string
	AuthToken string
	Projects  map[string][]byte
}

func NewMockSystemStore() *MockSystemStore {
	return &MockSystemStore{
		Settings: make(map[string]string),
		Projects: make(map[string][]byte),
	}
}

func (m *MockSystemStore) GetSetting(ctx context.Context, key string) (string, error) {
	if val, ok := m.Settings[key]; ok {
		return val, nil
	}
	return "", shared.ErrNotFound
}

func (m *MockSystemStore) SetSetting(ctx context.Context, key string, value string) error {
	m.Settings[key] = value
	return nil
}

func (m *MockSystemStore) SaveAuthToken(ctx context.Context, token string) error {
	m.AuthToken = token
	return nil
}

func (m *MockSystemStore) GetAuthToken(ctx context.Context) (string, error) {
	if m.AuthToken == "" {
		return "", shared.ErrNotFound
	}
	return m.AuthToken, nil
}

func (m *MockSystemStore) LogTelemetry(ctx context.Context, event string, data map[string]any) error {
	return nil
}

func (m *MockSystemStore) GetTelemetryHistory(ctx context.Context, limit int) ([]map[string]any, error) {
	return nil, nil
}

func (m *MockSystemStore) SaveProject(ctx context.Context, name string, projectData []byte) error {
	m.Projects[name] = projectData
	return nil
}

func (m *MockSystemStore) GetProject(ctx context.Context, name string) ([]byte, error) {
	if proj, ok := m.Projects[name]; ok {
		return proj, nil
	}
	return nil, shared.ErrNotFound
}

func (m *MockSystemStore) ListProjects(ctx context.Context) (map[string][]byte, error) {
	return m.Projects, nil
}

func (m *MockSystemStore) DeleteProject(ctx context.Context, name string) error {
	delete(m.Projects, name)
	return nil
}

func (m *MockSystemStore) Close(ctx context.Context) error {
	return nil
}

type MockLogger struct{}

func (m *MockLogger) Info(msg string, args ...any)             {}
func (m *MockLogger) Error(msg string, err error, args ...any) {}
func (m *MockLogger) Debug(msg string, args ...any)            {}
func (m *MockLogger) Warn(msg string, args ...any)             {}

func TestApp_Initialization(t *testing.T) {
	app := &App{
		systemStore: NewMockSystemStore(),
		logger:      &MockLogger{},
		gateway:     runtime.NewProjectGateway(),
	}

	var _ runtime.ProjectManager = app
}

func TestApp_SettingsPersistence(t *testing.T) {
	ctx := context.Background()
	app := &App{
		systemStore: NewMockSystemStore(),
		logger:      &MockLogger{},
		gateway:     runtime.NewProjectGateway(),
	}

	// 1. Initially should return error
	_, err := app.LoadSettingsConfig(ctx)
	if err == nil {
		t.Fatalf("Expected error when no settings saved")
	}

	// 2. Save settings
	sampleJSON := []byte(`{"active_theme": 2, "history_limit": 5000}`)
	err = app.SaveSettingsConfig(ctx, sampleJSON)
	if err != nil {
		t.Fatalf("Failed to save settings: %v", err)
	}

	// 3. Load settings
	loaded, err := app.LoadSettingsConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to load settings: %v", err)
	}
	if string(loaded) != string(sampleJSON) {
		t.Errorf("Expected %s, got %s", string(sampleJSON), string(loaded))
	}
}

func TestApp_GlobalCORS(t *testing.T) {
	ctx := context.Background()
	app := &App{
		systemStore: NewMockSystemStore(),
		logger:      &MockLogger{},
		gateway:     runtime.NewProjectGateway(),
	}
	app.corsEnabled.Store(true)

	// 1. Check default
	if !app.GlobalCORSEnabled() {
		t.Fatalf("expected CORS to be enabled by default")
	}

	// 2. SetGlobalCORS(false)
	app.SetGlobalCORS(false)
	if app.GlobalCORSEnabled() {
		t.Fatalf("expected CORS to be disabled")
	}

	// 3. SaveSettingsConfig with global_cors: true
	cfgJSON := []byte(`{"global_cors": true}`)
	err := app.SaveSettingsConfig(ctx, cfgJSON)
	if err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	if !app.GlobalCORSEnabled() {
		t.Fatalf("expected CORS to be enabled after SaveSettingsConfig")
	}

	// 4. SaveSettingsConfig with global_cors: false
	cfgJSON = []byte(`{"global_cors": false}`)
	err = app.SaveSettingsConfig(ctx, cfgJSON)
	if err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	if app.GlobalCORSEnabled() {
		t.Fatalf("expected CORS to be disabled after SaveSettingsConfig")
	}
}

func TestApp_StrictValidation(t *testing.T) {
	ctx := context.Background()
	app := &App{
		systemStore: NewMockSystemStore(),
		logger:      &MockLogger{},
		gateway:     runtime.NewProjectGateway(),
	}

	// 1. By default, strict validation is true
	if !app.StrictValidationEnabled() {
		t.Fatalf("expected strict validation to be enabled by default")
	}

	// 2. SetStrictValidation(false)
	app.SetStrictValidation(false)
	if app.StrictValidationEnabled() {
		t.Fatalf("expected strict validation to be false")
	}

	// 3. SaveSettingsConfig with strict_openapi_validation: true
	cfgJSON := []byte(`{"strict_openapi_validation": true}`)
	err := app.SaveSettingsConfig(ctx, cfgJSON)
	if err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	if !app.StrictValidationEnabled() {
		t.Fatalf("expected strict validation to be true after SaveSettingsConfig")
	}

	// 4. SaveSettingsConfig with strict_openapi_validation: false
	cfgJSON = []byte(`{"strict_openapi_validation": false}`)
	err = app.SaveSettingsConfig(ctx, cfgJSON)
	if err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	if app.StrictValidationEnabled() {
		t.Fatalf("expected strict validation to be false after SaveSettingsConfig")
	}
}
