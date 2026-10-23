package profiler_test

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/profiler"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestGoProfilerModel_Update(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := profiler.NewGoProfilerModel(tuiCtx)
	_ = model.Init(tuiCtx)

	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Esc should return OverlayNone
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = newModel.(*profiler.GoProfilerModel)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, types.OverlayStateEnum(types.OverlayNone), msg)

	viewStr := model.View()
	assert.Contains(t, viewStr, "Go Runtime Memory Profiler")
}

func TestGoProfilerModel_TelemetryEvent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := profiler.NewGoProfilerModel(tuiCtx)
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Trigger a TickMsg to read stats
	newModel, _ := model.Update(profiler.TickMsg(time.Now()))
	model = newModel.(*profiler.GoProfilerModel)

	viewStr := model.View()
	assert.Contains(t, viewStr, "Goroutines:")
	assert.Contains(t, viewStr, "Heap Alloc:")
	assert.Contains(t, viewStr, "Sys Memory:")
	assert.Contains(t, viewStr, "GC Pauses:")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
