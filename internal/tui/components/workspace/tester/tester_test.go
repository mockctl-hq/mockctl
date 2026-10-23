package tester_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace/tester"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestApiTesterModel_Layout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := tester.NewApiTesterModel(tuiCtx)
	_ = model.Init(tuiCtx)

	// Simulate a resize event
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model = newModel.(*tester.ApiTesterModel)

	assert.Equal(t, 100, model.Size.Width)
	
	view := model.View()
	assert.Contains(t, view, "HTTP/REST") // Protocol tab should be rendered
}

func TestApiTesterModel_Focus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := tester.NewApiTesterModel(tuiCtx)

	model.SetFocus(true)
	assert.True(t, model.IsFocused)

	model.SetFocus(false)
	assert.False(t, model.IsFocused)
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
