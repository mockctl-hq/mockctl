package logs_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/logs"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestGlobalLogsModel_Update(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := logs.NewGlobalLogsModel(tuiCtx)
	_ = model.Init(tuiCtx)

	// Test Esc key to quit overlay
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = newModel.(*logs.GlobalLogsModel)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, types.OverlayStateEnum(types.OverlayNone), msg)

	// Test search toggle '/'
	assert.False(t, model.Search.Focused())
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = newModel.(*logs.GlobalLogsModel)
	assert.True(t, model.Search.Focused())

	// Esc should cancel search mode instead of closing overlay if search mode is active
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = newModel.(*logs.GlobalLogsModel)
	assert.False(t, model.Search.Focused())
}

func TestGlobalLogsModel_TelemetryEvent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := logs.NewGlobalLogsModel(tuiCtx)

	// Setup window size so viewport can render
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Inject a raw telemetry event
	model.AddLogLine("INFO Test log message")

	// Check if view renders the event
	viewStr := model.View()
	assert.Contains(t, viewStr, "Test log message")
	assert.Contains(t, viewStr, "INFO")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
