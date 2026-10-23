package left_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/left"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/left/inner/body"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestLeftSidebarModel_Update(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := left.NewLeftSidebarModel(tuiCtx)
	_ = model.Init(tuiCtx)

	model.Update(tea.WindowSizeMsg{Width: 30, Height: 40})

	model.SetFocus(true)
	assert.True(t, model.IsFocused)

	view := model.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Projects")

	model.SetFocus(false)
	assert.False(t, model.IsFocused)
}

func TestLeftSidebarModel_WrapAroundNavigation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := left.NewLeftSidebarModel(tuiCtx)
	_ = model.Init(tuiCtx)
	model.SetFocus(true)

	// Add dummy visible nodes to test navigation
	model.Explorer.VisibleNodes = []*body.TreeNode{
		{ID: "proj-1", IsProject: true},
		{ID: "proj-2", IsProject: true},
		{ID: "proj-3", IsProject: true},
	}
	model.Explorer.ActiveIndex = 0

	// Up from index 0 should wrap around to index 2 (last)
	model.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 2, model.Explorer.ActiveIndex)

	// Down from index 2 should wrap around to index 0 (first)
	model.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, model.Explorer.ActiveIndex)
}

func TestLeftSidebarModel_CollapsedView(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := left.NewLeftSidebarModel(tuiCtx)
	_ = model.Init(tuiCtx)

	model.IsCollapsed = true
	model.Update(tea.WindowSizeMsg{Width: 5, Height: 40})

	view := model.View()
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "📁")
	assert.NotContains(t, view, "Projects")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
