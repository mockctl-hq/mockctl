package right_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/right"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRightSidebarModel_Update(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := right.NewRightSidebarModel(tuiCtx)
	_ = model.Init(tuiCtx)

	model.Update(tea.WindowSizeMsg{Width: 30, Height: 40})

	model.SetFocus(true)
	assert.True(t, model.IsFocused)

	view := model.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Chaos")

	model.SetFocus(false)
	assert.False(t, model.IsFocused)
}

func TestRightSidebarModel_TitleSyncOnNavigation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := right.NewRightSidebarModel(tuiCtx)
	_ = model.Init(tuiCtx)
	model.Update(tea.WindowSizeMsg{Width: 30, Height: 40})
	model.SetFocus(true)

	// 1. Initial tab is Chaos Controls
	assert.Equal(t, 0, model.Icons.ActiveIndex)
	assert.Equal(t, "Chaos Controls", model.Header.Title)
	assert.Contains(t, model.View(), "Chaos Controls")

	// 2. Down arrow moves to Code Snippets
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 1, model.Icons.ActiveIndex)
	assert.Equal(t, "Code Snippets", model.Header.Title)
	assert.Contains(t, model.View(), "Code Snippets")

	// 3. Down arrow moves to Static Overrides
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 2, model.Icons.ActiveIndex)
	assert.Equal(t, "Static Overrides", model.Header.Title)
	assert.Contains(t, model.View(), "Static Overrides")

	// 4. Down arrow wraps back to Chaos Controls
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, model.Icons.ActiveIndex)
	assert.Equal(t, "Chaos Controls", model.Header.Title)
	assert.Contains(t, model.View(), "Chaos Controls")
}

func TestRightSidebarModel_CollapsedView(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := right.NewRightSidebarModel(tuiCtx)
	_ = model.Init(tuiCtx)

	model.IsCollapsed = true
	model.Update(tea.WindowSizeMsg{Width: 5, Height: 40})

	view := model.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "🌀")
	assert.Contains(t, view, "💻")
	assert.Contains(t, view, "🟣")
	assert.NotContains(t, view, "Chaos Controls")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
