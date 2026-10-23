package tui_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/tui"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootModel_OverlayShortcuts_AltAndCtrl(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	model := tui.NewRootModel(tuiCtx)

	// Alt+S -> Settings
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}, Alt: true})
	root := newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlaySettings, root.ActiveOverlay)

	// Reset overlay
	root.ActiveOverlay = types.OverlayNone

	// Alt+H -> History
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}, Alt: true})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayHistory, root.ActiveOverlay)

	// Reset overlay
	root.ActiveOverlay = types.OverlayNone

	// Alt+L -> Logs
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}, Alt: true})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayGlobalLog, root.ActiveOverlay)

	// Reset overlay
	root.ActiveOverlay = types.OverlayNone

	// Alt+P -> Profiler
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}, Alt: true})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayProfiler, root.ActiveOverlay)

	// Reset overlay
	root.ActiveOverlay = types.OverlayNone

	// Alt+X -> Inspector
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayStateInspector, root.ActiveOverlay)

	// Reset overlay
	root.ActiveOverlay = types.OverlayNone

	// Alt+I -> Inspector as well
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}, Alt: true})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayStateInspector, root.ActiveOverlay)

	// Reset overlay
	root.ActiveOverlay = types.OverlayNone

	// Ctrl+H should NOT open History overlay anymore (strict Alt+ standard)
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayNone, root.ActiveOverlay)
}

func TestRootModel_ShowToastMsg(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	root := tui.NewRootModel(tuiCtx)
	_ = root.Init()

	newModel, _ := root.Update(types.ShowToastMsg{Message: "Saved successfully", IsError: false})
	root = newModel.(*tui.RootModel)

	headerView := root.Header.View()
	assert.Contains(t, headerView, "Saved successfully")
	assert.Contains(t, headerView, "✅")
}

func TestRootModel_DeleteProject_Overlay(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	root := tui.NewRootModel(tuiCtx)
	_ = root.Init()

	// Select a project
	newModel, _ := root.Update(types.ProjectSelectedMsg{ProjectName: "my-api"})
	root = newModel.(*tui.RootModel)

	// Press 'd' to open delete overlay
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayDeleteProject, root.ActiveOverlay)
	assert.True(t, root.DeleteProject.Active)

	// Simulate Esc to cancel
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyEscape})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayNone, root.ActiveOverlay)
	assert.False(t, root.DeleteProject.Active)
}

func TestRootModel_DaemonCrashed_Overlay(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	root := tui.NewRootModel(tuiCtx)
	_ = root.Init()
	root.Size = tea.WindowSizeMsg{Width: 150, Height: 40}

	newModel, _ := root.Update(types.DaemonCrashedMsg{Error: "port 8080 already in use"})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayDaemonCrashed, root.ActiveOverlay)

	view := root.View()
	assert.Contains(t, view, "Daemon Crashed")
	assert.Contains(t, view, "port 8080 already in use")
}

func TestRootModel_StateInspector_ActiveProject(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	root := tui.NewRootModel(tuiCtx)
	_ = root.Init()

	// When no project is selected
	newModel, _ := root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayStateInspector, root.ActiveOverlay)
	assert.Equal(t, "No project selected.", root.StateInspector.Error)

	// Reset overlay
	root.ActiveOverlay = types.OverlayNone

	// Now select a project
	newModel, _ = root.Update(types.ProjectSelectedMsg{ProjectName: "test-proj"})
	root = newModel.(*tui.RootModel)

	// Open State Inspector again
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}, Alt: true})
	root = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlayStateInspector, root.ActiveOverlay)
	assert.NotEqual(t, "No project selected.", root.StateInspector.Error)
}

func TestRootModel_SidebarToggling_CtrlB_CtrlE(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	root := tui.NewRootModel(tuiCtx)
	_ = root.Init()
	root.Size = tea.WindowSizeMsg{Width: 150, Height: 40}

	assert.True(t, root.LeftSidebarVisible)
	assert.True(t, root.RightSidebarVisible)

	// Ctrl+B toggles left sidebar off (collapsed into icon strip)
	newModel, _ := root.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	root = newModel.(*tui.RootModel)
	assert.False(t, root.LeftSidebarVisible)
	assert.True(t, root.LeftSidebar.IsCollapsed)

	// Ctrl+E toggles right sidebar off (collapsed into icon strip)
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	root = newModel.(*tui.RootModel)
	assert.False(t, root.RightSidebarVisible)
	assert.True(t, root.RightSidebar.IsCollapsed)

	// View renders without panicking and keeps icons visible even when drawers collapsed
	view := root.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "📁")                                     // Left icons strip remains visible
	assert.Contains(t, view, "🌀")                                     // Right icons strip remains visible
	assert.NotContains(t, root.LeftSidebar.View(), "Projects")        // Inner explorer drawer is collapsed
	assert.NotContains(t, root.RightSidebar.View(), "Chaos Controls") // Inner right drawer is collapsed

	// Ctrl+B toggles left sidebar back on (expanded)
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyCtrlB})
	root = newModel.(*tui.RootModel)
	assert.True(t, root.LeftSidebarVisible)
	assert.False(t, root.LeftSidebar.IsCollapsed)

	// Ctrl+E toggles right sidebar back on (expanded)
	newModel, _ = root.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	root = newModel.(*tui.RootModel)
	assert.True(t, root.RightSidebarVisible)
	assert.False(t, root.RightSidebar.IsCollapsed)
}

func TestRootModel_Telemetry_Streaming(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	root := tui.NewRootModel(tuiCtx)
	cmd := root.Init()
	require.NotNil(t, cmd)

	event := types.TelemetryMsg{
		Payload: &runtime.RequestEvent{
			HTTPMethod: "POST",
			Path:       "/api/orders",
			StatusCode: 201,
		},
	}

	newModel, _ := root.Update(event)
	root = newModel.(*tui.RootModel)
	assert.NotEmpty(t, root.History.Events)
}
