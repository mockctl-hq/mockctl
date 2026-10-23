package body_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/right/inner/body"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
)

func TestOverridesModel_EmptyEndpoint(t *testing.T) {
	model := body.NewOverridesModel(types.TUIContext{})
	view := model.View()
	assert.Contains(t, view, "Select an endpoint to create static overrides.")
}

func TestOverridesModel_ActiveEndpointView(t *testing.T) {
	model := body.NewOverridesModel(types.TUIContext{})
	model.ActiveProject = "test-project"
	model.ActiveEndpoint = "GET /users/{id}"
	model.SetFocus(true)

	view := model.View()
	assert.Contains(t, view, "Hardcoded JSON Response:")
	assert.Contains(t, view, "Default Mock")
	assert.NotContains(t, view, "Ctrl+S: Save")
}

func TestOverridesModel_InvalidJSONValidation(t *testing.T) {
	model := body.NewOverridesModel(types.TUIContext{})
	model.ActiveProject = "test-project"
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	// Set invalid JSON text in textarea
	model.Input.SetValue("{ not valid json }")

	// Trigger save via Ctrl+S
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	assert.NotNil(t, cmd)

	msg := cmd()
	toast, ok := msg.(types.ShowToastMsg)
	assert.True(t, ok)
	assert.True(t, toast.IsError)
	assert.Contains(t, toast.Message, "Invalid JSON")
}

func TestOverridesModel_ClearViaCtrlD(t *testing.T) {
	model := body.NewOverridesModel(types.TUIContext{})
	model.ActiveProject = "test-project"
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	model.Input.SetValue(`{"test": "val"}`)
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	assert.Empty(t, model.Input.Value())
	assert.False(t, model.HasOverride)
	assert.NotNil(t, cmd)
}

func TestOverridesModel_ThinPromptAndFocusColor(t *testing.T) {
	model := body.NewOverridesModel(types.TUIContext{})
	// Prompt should be thin "│ "
	assert.Equal(t, "│ ", model.Input.Prompt)

	// Blurred
	model.SetFocus(false)
	assert.False(t, model.Input.Focused())

	// Focused
	model.SetFocus(true)
	assert.True(t, model.Input.Focused())
}

func TestOverridesModel_EnterSavesValidJSON(t *testing.T) {
	model := body.NewOverridesModel(types.TUIContext{})
	model.ActiveProject = "test-project"
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	model.Input.SetValue(`{"status": "ok"}`)
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd)
	assert.True(t, model.HasOverride)
}

func TestOverridesModel_EnterReportsInvalidJSON(t *testing.T) {
	model := body.NewOverridesModel(types.TUIContext{})
	model.ActiveProject = "test-project"
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	// User types invalid JSON ending with }
	model.Input.SetValue("{\n  \"message\": hello\"\n}")
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd)

	msg := cmd()
	toast, ok := msg.(types.ShowToastMsg)
	assert.True(t, ok)
	assert.True(t, toast.IsError)
	assert.Contains(t, toast.Message, "Invalid JSON")
}

func TestOverridesModel_BadgeUpdatesWithSyntaxError(t *testing.T) {
	model := body.NewOverridesModel(types.TUIContext{})
	model.ActiveProject = "test-project"
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	// Valid JSON badge
	model.Input.SetValue(`{"test": 123}`)
	assert.Contains(t, model.View(), "Override Active (Valid JSON)")

	// Invalid JSON badge
	model.Input.SetValue(`{"test": 123`)
	assert.Contains(t, model.View(), "⚠️ Invalid JSON")
}
