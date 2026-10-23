package footer_test

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/footer"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestFooterModel_WorkspaceShortcuts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := footer.NewFooterModel(tuiCtx)
	_ = model.Init(tuiCtx)

	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model.SetFocusContext(types.FocusWorkspace, types.OverlayNone, false, false, 0, false)

	view := model.View()
	assert.Contains(t, view, "[Workspace]")
	assert.Contains(t, view, "Ctrl+S")
	assert.Contains(t, view, "Save")
	assert.Contains(t, view, "Tester")
}

func TestFooterModel_LeftPanelShortcuts(t *testing.T) {
	model := footer.NewFooterModel(types.TUIContext{})
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Normal Left Panel
	model.SetFocusContext(types.FocusLeftPanel, types.OverlayNone, false, false, 0, false)
	view := model.View()
	assert.Contains(t, view, "[Left Panel]")
	assert.Contains(t, view, "Nav")
	assert.Contains(t, view, "Open")
	assert.Contains(t, view, "Search")

	// Search active
	model.SetFocusContext(types.FocusLeftPanel, types.OverlayNone, true, false, 0, false)
	viewSearch := model.View()
	assert.Contains(t, viewSearch, "[Search]")
	assert.Contains(t, viewSearch, "Filter")

	// Scaffold active
	model.SetFocusContext(types.FocusLeftPanel, types.OverlayNone, false, true, 0, false)
	viewScaffold := model.View()
	assert.Contains(t, viewScaffold, "[New Project]")
	assert.Contains(t, viewScaffold, "Create")
}

func TestFooterModel_ApiTesterShortcuts(t *testing.T) {
	model := footer.NewFooterModel(types.TUIContext{})
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	model.SetFocusContext(types.FocusApiTester, types.OverlayNone, false, false, 0, false)
	view := model.View()
	assert.Contains(t, view, "[API Tester]")
	assert.Contains(t, view, "Send")
	assert.Contains(t, view, "Execute")
	assert.Contains(t, view, "Cancel")
}

func TestFooterModel_RightPanelShortcuts(t *testing.T) {
	model := footer.NewFooterModel(types.TUIContext{})
	model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Icons focused
	model.SetFocusContext(types.FocusRightPanel, types.OverlayNone, false, false, 0, true)
	viewIcons := model.View()
	assert.Contains(t, viewIcons, "[Right Panel]")
	assert.Contains(t, viewIcons, "Select Tab")

	// Chaos tab body
	model.SetFocusContext(types.FocusRightPanel, types.OverlayNone, false, false, 0, false)
	viewChaos := model.View()
	assert.Contains(t, viewChaos, "[Chaos Controls]")
	assert.Contains(t, viewChaos, "Toggle")
	assert.Contains(t, viewChaos, "Tabs")

	// Snippets tab body
	model.SetFocusContext(types.FocusRightPanel, types.OverlayNone, false, false, 1, false)
	viewSnippets := model.View()
	assert.Contains(t, viewSnippets, "[Code Snippets]")
	assert.Contains(t, viewSnippets, "Copy")

	// Static Overrides tab body
	model.SetFocusContext(types.FocusRightPanel, types.OverlayNone, false, false, 2, false)
	viewStatic := model.View()
	assert.Contains(t, viewStatic, "[Static Overrides]")
	assert.Contains(t, viewStatic, "Save")
	assert.Contains(t, viewStatic, "Clear")
}

func TestFooterModel_OverlayShortcuts(t *testing.T) {
	model := footer.NewFooterModel(types.TUIContext{})
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Settings
	model.SetFocusContext(types.FocusOverlay, types.OverlaySettings, false, false, 0, false)
	assert.Contains(t, model.View(), "[Settings]")

	// History
	model.SetFocusContext(types.FocusOverlay, types.OverlayHistory, false, false, 0, false)
	assert.Contains(t, model.View(), "[History]")

	// Logs
	model.SetFocusContext(types.FocusOverlay, types.OverlayGlobalLog, false, false, 0, false)
	assert.Contains(t, model.View(), "[Logs]")

	// Profiler
	model.SetFocusContext(types.FocusOverlay, types.OverlayProfiler, false, false, 0, false)
	assert.Contains(t, model.View(), "[Profiler]")

	// Inspector
	model.SetFocusContext(types.FocusOverlay, types.OverlayStateInspector, false, false, 0, false)
	assert.Contains(t, model.View(), "[Inspector]")

	// Import Spec
	model.SetFocusContext(types.FocusOverlay, types.OverlayImportSpec, false, false, 0, false)
	assert.Contains(t, model.View(), "[Import Spec]")
}

func TestFooterModel_NarrowWidthNoOverflow(t *testing.T) {
	model := footer.NewFooterModel(types.TUIContext{})
	model.Update(tea.WindowSizeMsg{Width: 30, Height: 40})
	model.SetFocusContext(types.FocusWorkspace, types.OverlayNone, false, false, 0, false)

	view := model.View()
	// Should contain the badge without panicking
	assert.Contains(t, view, "[Workspace]")
	// Height should remain exactly 2 lines (1 border + 1 text line)
	lines := strings.Split(view, "\n")
	assert.Equal(t, 2, len(lines))
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
