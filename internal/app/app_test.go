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
