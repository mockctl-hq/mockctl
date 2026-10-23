package body_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/right/inner/body"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
)

func TestChaosControlModel_EmptyEndpoint(t *testing.T) {
	model := body.NewChaosControlModel(types.TUIContext{})
	view := model.View()
	assert.Contains(t, view, "Select an endpoint to configure chaos constraints.")
}

func TestChaosControlModel_FieldNavigationWrapAround(t *testing.T) {
	model := body.NewChaosControlModel(types.TUIContext{})
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	// Starts at field 0 (Enable checkbox)
	assert.Equal(t, 0, model.ActiveField)

	// Up wraps to field 3 (Force Status)
	model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 3, model.ActiveField)

	// Down wraps back to field 0
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, model.ActiveField)

	// Down moves to field 1 (Delay)
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, model.ActiveField)

	// Tab moves to field 2 (Error Rate)
	model.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 2, model.ActiveField)

	// Shift+Tab moves back to field 1
	model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, 1, model.ActiveField)
}

func TestChaosControlModel_ToggleEnable(t *testing.T) {
	model := body.NewChaosControlModel(types.TUIContext{})
	model.ActiveProject = "test-proj"
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	assert.False(t, model.Enabled)

	// Space on field 0 toggles Enable
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	assert.True(t, model.Enabled)

	view := model.View()
	assert.Contains(t, view, "[x] Enable Chaos")
	assert.Contains(t, view, "Chaos Running")

	// Space again toggles off
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	assert.False(t, model.Enabled)

	view = model.View()
	assert.Contains(t, view, "[ ] Enable Chaos")
	assert.Contains(t, view, "Inactive")
}

func TestChaosControlModel_AutoDefaultsAndInputs(t *testing.T) {
	model := body.NewChaosControlModel(types.TUIContext{})
	model.ActiveProject = "test-proj"
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	// Space on field 0 toggles Enable and fills defaults
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	assert.True(t, model.Enabled)
	assert.Equal(t, "25", model.ErrorRateInput.Value())
	assert.Equal(t, "500", model.StatusCodeInput.Value())

	// Move to status input (field 3) and change to 400
	model.ActiveField = 3
	model.StatusCodeInput.SetValue("400")
	view := model.View()
	assert.Contains(t, view, "400")
}

func TestChaosControlModel_RealtimeAutoSyncWithoutEnter(t *testing.T) {
	model := body.NewChaosControlModel(types.TUIContext{})
	model.ActiveProject = "test-proj"
	model.ActiveEndpoint = "GET /users"
	model.SetFocus(true)

	// Initially disabled
	assert.False(t, model.Enabled)

	// Move to field 1 (Delay)
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, model.ActiveField)

	// Type '3', '0', '0' without pressing Enter
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	assert.Equal(t, "300", model.DelayInput.Value())
	// Should auto-enable chaos without pressing Enter!
	assert.True(t, model.Enabled)

	// Move to field 2 (Error Rate)
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, model.ActiveField)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	assert.Equal(t, "50", model.ErrorRateInput.Value())

	// Move to field 3 (Force Status)
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 3, model.ActiveField)
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}})
	assert.Equal(t, "429", model.StatusCodeInput.Value())

	// Blur focus without hitting Enter
	model.SetFocus(false)
	assert.False(t, model.IsFocused)
	assert.True(t, model.Enabled)
	assert.Equal(t, "300", model.DelayInput.Value())
	assert.Equal(t, "50", model.ErrorRateInput.Value())
	assert.Equal(t, "429", model.StatusCodeInput.Value())

	view := model.View()
	assert.Contains(t, view, "Chaos Running")
	assert.Contains(t, view, "[x] Enable Chaos")
	assert.Contains(t, view, "Delay (ms):")
}

