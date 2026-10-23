package history_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/history"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestHistoryModel_Update(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := history.NewHistoryModel(tuiCtx)
	_ = model.Init(tuiCtx)

	// Setup WindowSize
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Test Esc key to quit overlay
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = newModel.(*history.HistoryModel)
	assert.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, types.OverlayStateEnum(types.OverlayNone), msg)

	// Test search toggle '/'
	assert.False(t, model.Search.Focused())
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = newModel.(*history.HistoryModel)
	assert.True(t, model.Search.Focused())

	// Esc should blur search if active
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = newModel.(*history.HistoryModel)
	assert.False(t, model.Search.Focused())
}

func TestHistoryModel_TelemetryEvent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := history.NewHistoryModel(tuiCtx)
	model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})

	// Inject telemetry event
	event := types.TelemetryMsg{
		Payload: &runtime.RequestEvent{
			HTTPMethod: "GET",
			Path:       "/api/users",
			StatusCode: 200,
		},
	}

	newModel, _ := model.Update(event)
	model = newModel.(*history.HistoryModel)

	viewStr := model.View()
	assert.Contains(t, viewStr, "GET")
	assert.Contains(t, viewStr, "/api/users")
	assert.Contains(t, viewStr, "200")
	assert.Contains(t, viewStr, "Press / to search | Esc to close")
	assert.Contains(t, viewStr, "─")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
