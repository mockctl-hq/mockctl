package workspace_test

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestWorkspaceModel_Layout(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := workspace.NewWorkspaceModel(tuiCtx)
	_ = model.Init(tuiCtx)

	// Simulate a resize event
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model = newModel.(*workspace.WorkspaceModel)

	// Verify that height was distributed properly to Editor
	assert.Equal(t, 100, model.Size.Width)
	
	view := model.View()
	assert.Contains(t, view, "Ln") // Status bar should be rendered
}

func TestWorkspaceModel_Focus(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := workspace.NewWorkspaceModel(tuiCtx)

	model.SetFocus(true)
	assert.True(t, model.IsFocused)
	assert.True(t, model.Editor.IsFocused)

	model.SetFocus(false)
	assert.False(t, model.IsFocused)
	assert.False(t, model.Editor.IsFocused)
}

func TestWorkspaceModel_SettingsChangedRouting(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := workspace.NewWorkspaceModel(tuiCtx)
	_ = model.Init(tuiCtx)

	// Workspace is NOT focused, ApiTester is NOT focused
	assert.False(t, model.IsFocused)
	assert.False(t, model.ApiTester.IsFocused)

	// Check default timeout and redirects on tester client
	assert.Equal(t, 10*time.Second, model.ApiTester.HttpTester.Client.Timeout())
	assert.False(t, model.ApiTester.HttpTester.Client.FollowRedirectsEnabled())

	// Dispatch SettingsChangedMsg with 30s timeout and FollowRedirects=true
	cfg := &types.SettingsConfig{
		RequestTimeout:  30 * time.Second,
		FollowRedirects: true,
	}
	model.Update(types.SettingsChangedMsg{Config: cfg})

	// Verify tester client updated immediately!
	assert.Equal(t, 30*time.Second, model.ApiTester.HttpTester.Client.Timeout())
	assert.True(t, model.ApiTester.HttpTester.Client.FollowRedirectsEnabled())
}

func TestWorkspaceModel_EndpointSelectedRouting(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := workspace.NewWorkspaceModel(tuiCtx)
	_ = model.Init(tuiCtx)

	// Dispatch EndpointSelectedMsg
	model.Update(types.EndpointSelectedMsg{
		ProjectName:  "petstore",
		EndpointPath: "POST /pets",
	})

	// Verify ApiTester updated method and path input
	assert.Equal(t, "POST", model.ApiTester.HttpTester.Input.Method)
	assert.Equal(t, "/petstore/pets", model.ApiTester.HttpTester.Input.PathInput.Value())
}

func TestWorkspaceModel_Editor_SettingsRouting(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	cfg := types.DefaultSettingsConfig()
	cfg.LineWrapping = false
	cfg.TabSize = 2
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, Settings: cfg}
	model := workspace.NewWorkspaceModel(tuiCtx)
	_ = model.Init(tuiCtx)

	assert.False(t, model.Editor.LineWrappingEnabled())
	assert.Equal(t, 2, model.Editor.CurrentTabSize())

	// Update settings via SettingsChangedMsg
	newCfg := &types.SettingsConfig{
		LineWrapping:     true,
		TabSize:          4,
		AutoFormatOnSave: false,
	}
	model.Update(types.SettingsChangedMsg{Config: newCfg})

	assert.True(t, model.Editor.LineWrappingEnabled())
	assert.Equal(t, 4, model.Editor.CurrentTabSize())
	assert.False(t, model.Editor.AutoFormatEnabled())
}

func TestWorkspaceModel_Editor_LineWrapping(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	cfg := types.DefaultSettingsConfig()
	cfg.LineWrapping = false
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, Settings: cfg}
	model := workspace.NewWorkspaceModel(tuiCtx)
	_ = model.Init(tuiCtx)

	// Resize so viewport has 40 width
	model.Update(tea.WindowSizeMsg{Width: 40, Height: 20})

	// Set content with a long line (80 characters)
	longLine := "openapi: 3.0.0 info title VeryLongTitleThatDefinitelyExceedsTheThirtyCharacterWidthLimit version 1.0.0"
	_ = model.Editor.SetContent(longLine)

	// When LineWrapping is false, no continuation lines "     │ " should be present
	viewUnwrapped := model.Editor.Viewport.View()
	assert.NotContains(t, viewUnwrapped, "     │ ")

	// Enable LineWrapping
	cfg.LineWrapping = true
	model.Update(types.SettingsChangedMsg{Config: cfg})

	// When LineWrapping is true, continuation lines "     │ " should be present
	viewWrapped := model.Editor.Viewport.View()
	assert.Contains(t, viewWrapped, "     │ ")
}

func TestWorkspaceModel_Editor_TabExpansion(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	cfg := types.DefaultSettingsConfig()
	cfg.TabSize = 4
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, Settings: cfg}
	model := workspace.NewWorkspaceModel(tuiCtx)
	_ = model.Init(tuiCtx)
	model.Update(tea.WindowSizeMsg{Width: 60, Height: 20})

	_ = model.Editor.SetContent("\tkey: value")
	view := model.Editor.Viewport.View()
	assert.Contains(t, ansi.Strip(view), "    key: value") // 4 spaces

	// Switch to TabSize = 2
	cfg.TabSize = 2
	model.Update(types.SettingsChangedMsg{Config: cfg})
	view2 := model.Editor.Viewport.View()
	assert.Contains(t, ansi.Strip(view2), "  key: value") // 2 spaces
}

func TestFormatPayload_YAML(t *testing.T) {
	yamlInput := `key:
  nested:
    value: 123
`
	// Format with tabSize 4
	formatted4, err := workspace.FormatPayload(yamlInput, true, 4)
	require.NoError(t, err)
	assert.Contains(t, formatted4, "    nested:")

	// Format with tabSize 2
	formatted2, err := workspace.FormatPayload(yamlInput, true, 2)
	require.NoError(t, err)
	assert.Contains(t, formatted2, "  nested:")
}

func TestWorkspaceModel_Editor_FormatShortcut(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	cfg := types.DefaultSettingsConfig()
	cfg.TabSize = 4
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, Settings: cfg}
	model := workspace.NewWorkspaceModel(tuiCtx)
	_ = model.Init(tuiCtx)
	model.Editor.SetFocus(true)

	unformatted := "key:\n  nested:\n    val: 1\n"
	_ = model.Editor.SetContent(unformatted)

	// Press ctrl+f
	model.Editor.Update(tea.KeyMsg{Type: tea.KeyCtrlF})

	assert.Contains(t, model.Editor.RawContent, "    nested:")
}

func TestWorkspaceModel_CloseTab_RemovesTabCompletely(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core}
	model := workspace.NewWorkspaceModel(tuiCtx)
	_ = model.Init(tuiCtx)
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// 1. Open a tab for "todo-app"
	model.Update(types.ProjectSelectedMsg{ProjectName: "todo-app", FileName: "api.yaml"})
	assert.Equal(t, 1, len(model.Tabs.Tabs))
	assert.Equal(t, "todo-app", model.Tabs.Tabs[0].ProjectName)
	assert.Contains(t, model.View(), "api.yaml")

	// 2. Press Ctrl+W to close tab
	_, cmd := model.Tabs.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	require.NotNil(t, cmd)
	msg := cmd()
	assert.Equal(t, types.ProjectSelectedMsg{ProjectName: "", FileName: ""}, msg)

	// Tab count should be 0 immediately
	assert.Equal(t, 0, len(model.Tabs.Tabs))

	// 3. Dispatch the resulting ProjectSelectedMsg{ProjectName: "", FileName: ""} to workspace
	model.Update(msg)

	// Tab count MUST remain 0, and not recreate a ghost tab for ""
	assert.Equal(t, 0, len(model.Tabs.Tabs))
	assert.NotContains(t, model.View(), "api.yaml")
	assert.Contains(t, model.Editor.View(), "No Content")
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m, goleak.IgnoreAnyFunction("github.com/dlclark/regexp2/v2.runClock"))
}
