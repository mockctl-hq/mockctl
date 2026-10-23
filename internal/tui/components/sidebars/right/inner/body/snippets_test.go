package body_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/right/inner/body"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
)

func TestCodeSnippetsModel_DynamicPort(t *testing.T) {
	// Custom port
	model := body.NewCodeSnippetsModel(types.TUIContext{ActivePort: 9090})
	model.ActiveEndpoint = "GET /api/v1/users"
	assert.Contains(t, model.GenerateSnippet(), "http://localhost:9090/api/v1/users")

	// Fallback port (<= 0)
	modelDefault := body.NewCodeSnippetsModel(types.TUIContext{ActivePort: 0})
	modelDefault.ActiveEndpoint = "GET /api/v1/users"
	assert.Contains(t, modelDefault.GenerateSnippet(), "http://localhost:8080/api/v1/users")
}

func TestCodeSnippetsModel_UniversalPathInterpolation(t *testing.T) {
	model := body.NewCodeSnippetsModel(types.TUIContext{ActivePort: 8080})
	model.ActiveEndpoint = "GET /stores/{storeId}/pets/{petId}/orders/{orderGuid}"
	
	snippet := model.GenerateSnippet()
	assert.Contains(t, snippet, "/stores/1/pets/1/orders/1")
	assert.NotContains(t, snippet, "{storeId}")
	assert.NotContains(t, snippet, "{petId}")
	assert.NotContains(t, snippet, "{orderGuid}")
}

func TestCodeSnippetsModel_MethodSpecificTemplates(t *testing.T) {
	model := body.NewCodeSnippetsModel(types.TUIContext{ActivePort: 8080})

	// GET cURL
	model.ActiveLanguage = 0
	model.ActiveEndpoint = "GET /items"
	curlGet := model.GenerateSnippet()
	assert.Contains(t, curlGet, "curl -X GET")
	assert.NotContains(t, curlGet, "-d '{}'")

	// POST cURL
	model.ActiveEndpoint = "POST /items"
	curlPost := model.GenerateSnippet()
	assert.Contains(t, curlPost, "curl -X POST")
	assert.Contains(t, curlPost, "-H 'Content-Type: application/json'")
	assert.Contains(t, curlPost, "-d '{}'")

	// POST Fetch
	model.ActiveLanguage = 1
	fetchPost := model.GenerateSnippet()
	assert.Contains(t, fetchPost, "fetch('http://localhost:8080/items'")
	assert.Contains(t, fetchPost, "method: 'POST'")
	assert.Contains(t, fetchPost, "'Content-Type': 'application/json'")
	assert.Contains(t, fetchPost, "body: JSON.stringify({})")
	assert.Contains(t, fetchPost, ".then(res => res.json())")
	assert.Contains(t, fetchPost, ".then(data => console.log(data))")
}

func TestCodeSnippetsModel_NavigationWrapAround(t *testing.T) {
	model := body.NewCodeSnippetsModel(types.TUIContext{ActivePort: 8080})
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	// Starts at 0 (cURL)
	assert.Equal(t, 0, model.ActiveLanguage)

	// Left key wraps to 1 (Fetch)
	model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, 1, model.ActiveLanguage)

	// Left key wraps to 0 (cURL)
	model.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.Equal(t, 0, model.ActiveLanguage)

	// Right key moves to 1 (Fetch)
	model.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 1, model.ActiveLanguage)

	// Right key wraps to 0 (cURL)
	model.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.Equal(t, 0, model.ActiveLanguage)

	// Tab moves to 1
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 1, model.ActiveLanguage)

	// Shift+Tab moves back to 0
	model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, 0, model.ActiveLanguage)

	// Up arrow wraps to 1
	model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, model.ActiveLanguage)

	// Down arrow wraps to 0
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, model.ActiveLanguage)
}

func TestCodeSnippetsModel_CopyKeybinding(t *testing.T) {
	model := body.NewCodeSnippetsModel(types.TUIContext{ActivePort: 8080})
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	// Pressing 'c' or 'enter' or 'y' generates copy command and toast
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	assert.NotNil(t, cmd)

	// View should reflect focus hints
	view := model.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "cURL")
	assert.Contains(t, view, "Fetch")
}

func TestCodeSnippetsModel_EmptyEndpoint(t *testing.T) {
	model := body.NewCodeSnippetsModel(types.TUIContext{})
	view := model.View()
	assert.Contains(t, view, "Select an endpoint to generate code snippets.")
}

func TestCodeSnippetsModel_ProjectPrefixInUrl(t *testing.T) {
	model := body.NewCodeSnippetsModel(types.TUIContext{ActivePort: 8080})
	model.ActiveProject = "my-project"
	model.ActiveEndpoint = "GET /users"

	// cURL with project prefix
	model.ActiveLanguage = 0
	curlSnippet := model.GenerateSnippet()
	assert.Contains(t, curlSnippet, "http://localhost:8080/my-project/users")

	// Fetch with project prefix
	model.ActiveLanguage = 1
	fetchSnippet := model.GenerateSnippet()
	assert.Contains(t, fetchSnippet, "http://localhost:8080/my-project/users")

	// Test message updates
	model.Update(types.ProjectSelectedMsg{ProjectName: "other-project"})
	assert.Equal(t, "other-project", model.ActiveProject)
	assert.Contains(t, model.GenerateSnippet(), "http://localhost:8080/other-project/users")

	model.Update(types.EndpointSelectedMsg{ProjectName: "proj-3", EndpointPath: "POST /todos"})
	assert.Equal(t, "proj-3", model.ActiveProject)
	assert.Equal(t, "POST /todos", model.ActiveEndpoint)
	assert.Contains(t, model.GenerateSnippet(), "http://localhost:8080/proj-3/todos")
}

func TestCodeSnippetsModel_ReactiveTesterDraft(t *testing.T) {
	model := body.NewCodeSnippetsModel(types.TUIContext{ActivePort: 8080})
	model.ActiveProject = "my-proj"
	model.ActiveEndpoint = "POST /users"

	// Live draft update from API Tester
	model.Update(types.TesterDraftChangedMsg{
		Method:  "POST",
		Path:    "/my-proj/users",
		Body:    `{"username":"johndoe"}`,
		Headers: "Content-Type: application/json",
	})

	// cURL snippet should reflect custom body
	model.ActiveLanguage = 0
	curlStr := model.GenerateSnippet()
	assert.Contains(t, curlStr, `-d '{"username":"johndoe"}'`)

	// Fetch snippet should reflect custom body
	model.ActiveLanguage = 1
	fetchStr := model.GenerateSnippet()
	assert.Contains(t, fetchStr, `body: JSON.stringify({"username":"johndoe"})`)
}
