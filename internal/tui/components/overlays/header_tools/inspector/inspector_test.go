package inspector_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/inspector"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestStateInspectorModel_Update(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := inspector.NewStateInspectorModel(tuiCtx)
	_ = model.Init(tuiCtx)

	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Esc should return OverlayNone
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = newModel.(*inspector.StateInspectorModel)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, types.OverlayStateEnum(types.OverlayNone), msg)

	viewStr := model.View()
	assert.Contains(t, viewStr, "State Inspector")
	assert.Contains(t, viewStr, "Esc / q: Close")
}

func TestStateInspectorModel_Footer(t *testing.T) {
	model := inspector.NewStateInspectorModel(types.TUIContext{})
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// When empty
	viewEmpty := model.View()
	assert.Contains(t, viewEmpty, "State is empty.")
	assert.Contains(t, viewEmpty, "Esc / q: Close")
	assert.Contains(t, viewEmpty, "Enter: Expand/Collapse")

	// When error
	model.Error = "Sample Error"
	viewErr := model.View()
	assert.Contains(t, viewErr, "Sample Error")
	assert.Contains(t, viewErr, "Esc / q: Close")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
