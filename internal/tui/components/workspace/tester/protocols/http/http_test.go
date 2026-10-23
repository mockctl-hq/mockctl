package http_test

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace/tester/protocols/http"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestHttpTesterModel_FocusRouting(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := http.NewHttpTesterModel(tuiCtx)
	_ = model.Init(tuiCtx)

	model.SetFocus(true)
	
	// Initial focus should be FocusInput
	assert.Equal(t, http.FocusInput, model.FocusRing)

	// AdvanceFocus should route to FocusBody
	stay := model.AdvanceFocus()
	assert.True(t, stay)
	assert.Equal(t, http.FocusBody, model.FocusRing)

	// AdvanceFocus should route to FocusHeaders
	stay = model.AdvanceFocus()
	assert.True(t, stay)
	assert.Equal(t, http.FocusHeaders, model.FocusRing)

	// AdvanceFocus should route to FocusViewer
	stay = model.AdvanceFocus()
	assert.True(t, stay)
	assert.Equal(t, http.FocusViewer, model.FocusRing)

	// AdvanceFocus should wrap/reset to FocusInput and return false (telling parent to move to next panel)
	stay = model.AdvanceFocus()
	assert.False(t, stay)
	assert.Equal(t, http.FocusInput, model.FocusRing)

	// ReverseFocus from FocusInput should return false (telling parent to move to prev panel)
	stay = model.ReverseFocus()
	assert.False(t, stay)

	// If at FocusViewer, ReverseFocus should move to FocusHeaders
	model.FocusRing = http.FocusViewer
	stay = model.ReverseFocus()
	assert.True(t, stay)
	assert.Equal(t, http.FocusHeaders, model.FocusRing)
}

func TestHttpTesterModel_Cancellation(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := http.NewHttpTesterModel(tuiCtx)
	model.SetFocus(true)

	// Execute Test command (Ctrl+R)
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlR})
	model = newModel.(*http.HttpTesterModel)
	assert.NotNil(t, cmd, "Expected cmd to execute request")

	// Press Esc to cancel pending request
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = newModel.(*http.HttpTesterModel)
	
	viewStr := model.View()
	assert.Contains(t, viewStr, "request cancelled by user")
}

func TestHttpTesterModel_StartupSettingsSync(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	cfg := types.DefaultSettingsConfig()
	cfg.RequestTimeout = 30 * time.Second
	cfg.FollowRedirects = true

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, Settings: cfg}
	model := http.NewHttpTesterModel(tuiCtx)
	_ = model.Init(tuiCtx)

	require.NotNil(t, model.Client)
	assert.Equal(t, 30*time.Second, model.Client.Timeout())
}

func TestHttpViewerModel_PlaceholderSpace(t *testing.T) {
	model := http.NewHttpViewerModel(types.TUIContext{})
	view := model.View()
	// Should have space before "Send a request to see the response here."
	assert.Contains(t, view, " Send a request to see the response here.")
}

func TestHttpViewerModel_LoadingSpinner(t *testing.T) {
	model := http.NewHttpViewerModel(types.TUIContext{})
	assert.False(t, model.IsLoading)

	cmd := model.SetLoading(true)
	assert.NotNil(t, cmd)
	assert.True(t, model.IsLoading)

	view := model.View()
	assert.Contains(t, view, "Loading...")
	// Spinner view character is present
	assert.Contains(t, view, model.Spinner.View())

	// SetResponse completes loading
	model.SetResponse(&http.TestResponse{
		StatusCode: 200,
		StatusText: "200 OK",
		Body:       `{"status":"ok"}`,
	})
	assert.False(t, model.IsLoading)
	assert.Contains(t, model.View(), "200 OK")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/dlclark/regexp2/v2.runClock"))
}
