package settings_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/settings"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestSettingsModel_Update_Tabs(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := settings.NewSettingsModel(tuiCtx)
	model.Active = true

	// Ensure ActiveTab is 0 initially (Theme)
	assert.Equal(t, 0, model.ActiveTab)

	// Tab Forward
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newModel.(*settings.SettingsModel)
	assert.Equal(t, 1, model.ActiveTab) // Network

	// Tab Forward again
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = newModel.(*settings.SettingsModel)
	assert.Equal(t, 2, model.ActiveTab) // Editor

	// Shift+Tab backward
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = newModel.(*settings.SettingsModel)
	assert.Equal(t, 1, model.ActiveTab) // Back to Network

	// Quit overlay (Esc)
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = newModel.(*settings.SettingsModel)
	assert.False(t, model.Active)
	msg := cmd()
	assert.IsType(t, types.SettingsChangedMsg{}, msg, "Expected Esc to return SettingsChangedMsg to root model")
}

func TestSettingsModel_View_Renders(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, _ := app.NewApp(ctx)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := settings.NewSettingsModel(tuiCtx)
	model.Active = true
	model.Size = tea.WindowSizeMsg{Width: 100, Height: 40}

	viewStr := model.View()
	assert.Contains(t, viewStr, "Settings")
	assert.Contains(t, viewStr, "Theme")
	assert.Contains(t, viewStr, "Network")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
