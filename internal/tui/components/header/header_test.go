package header_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/header"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestHeaderModel_Update(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := header.NewHeaderModel(tuiCtx)
	_ = model.Init(tuiCtx)

	// Test WindowSize
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Test Blink (ColonToggle)
	assert.True(t, model.Logo.ColonVisible)
	newModel, _ := model.Update(types.BlinkMsg{})
	model = newModel.(*header.HeaderModel)
	assert.False(t, model.Logo.ColonVisible)

	view := model.View()
	assert.Contains(t, view, "𝗠𝗼𝗰𝗸")
	assert.Contains(t, view, "Inspector")

	// Test Toast Rendering
	newModel, _ = model.Update(types.ShowToastMsg{Message: "Test toast message", IsError: false})
	model = newModel.(*header.HeaderModel)
	toastView := model.View()
	assert.Contains(t, toastView, "Test toast message")
	assert.Contains(t, toastView, "✅")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
