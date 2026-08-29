package runtime

import (
	"context"

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

// mockProjectManager implements runtime.ProjectManager
type MockProjectManager struct {
	Projects map[string]map[string]any
}

func NewMockProjectManager() *MockProjectManager {
	return &MockProjectManager{
		Projects: make(map[string]map[string]any),
	}
}

func (m *MockProjectManager) CreateProject(ctx context.Context, name string, fileStream []byte) error {
	m.Projects[name] = map[string]any{"status": "active"}
	return nil
}

func (m *MockProjectManager) ListProjects(ctx context.Context) (map[string]map[string]any, error) {
	return m.Projects, nil
}

func (m *MockProjectManager) DeleteProject(ctx context.Context, name string) error {
	delete(m.Projects, name)
	return nil
}

func (m *MockProjectManager) AddEndpoint(ctx context.Context, projectName string, endpointJSON []byte) error {
	return nil
}

func (m *MockProjectManager) SetOverrides(ctx context.Context, projectName string, overridesJSON []byte) error {
	return nil
}

func (m *MockProjectManager) ResetState(ctx context.Context, projectName string) error {
	return nil
}

func (m *MockProjectManager) UpdateChaos(ctx context.Context, projectName string, errorRate int, latencyMs int) error {
	return nil
}

type MockLogger struct{}

func (m *MockLogger) Info(msg string, args ...any)             {}
func (m *MockLogger) Error(msg string, err error, args ...any) {}
func (m *MockLogger) Debug(msg string, args ...any)            {}
func (m *MockLogger) Warn(msg string, args ...any)             {}
