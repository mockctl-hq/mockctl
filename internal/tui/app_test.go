package tui_test

import (
	"context"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

// PKS-029: Elm Architecture Unit Tests & Parallelism

func TestRootModel_Update_Routing(t *testing.T) {

	// 1. Setup mock App Core with isolated HOME for BBolt
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{
		Ctx:        ctx,
		Core:       core,
		ActivePort: 8080,
	}

	model := tui.NewRootModel(tuiCtx)

	// 2. Simulate resizing window
	newModel, cmd := model.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	rootModel := newModel.(*tui.RootModel)
	assert.Nil(t, cmd, "Expected nil cmd for resize (except what workspace returns)")
	assert.Equal(t, 150, rootModel.Size.Width)
	assert.Equal(t, 40, rootModel.Size.Height)

	// Expand the API Tester so Tab can route into it (collapsed panels are skipped per IMP-008)
	rootModel.ApiTesterExpanded = true

	// 3. Simulate Tab Forward Focus (Workspace -> ApiTester)
	newModel, _ = rootModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	rootModel = newModel.(*tui.RootModel)
	assert.Equal(t, types.FocusApiTester, rootModel.FocusRing)

	// 4. Simulate Ctrl+S (Settings Overlay)
	// Let's manually trigger the overlay via OverlayStateEnum just to verify the state change
	newModel, _ = rootModel.Update(types.OverlayStateEnum(types.OverlaySettings))
	rootModel = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlaySettings, rootModel.ActiveOverlay)

	// 5. Test Overlay preemption: when overlay is active, Tab should not change focus ring
	newModel, _ = rootModel.Update(tea.KeyMsg{Type: tea.KeyTab})
	rootModel = newModel.(*tui.RootModel)
	// FocusRing should remain FocusApiTester because active overlay preempts Tab
	assert.Equal(t, types.FocusApiTester, rootModel.FocusRing)

	// 6. Test Quit Command
	_, cmd = rootModel.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	msg := cmd()
	assert.Equal(t, tea.Quit(), msg, "Expected tea.Quit cmd on Ctrl+C")
}

func TestRootModel_DeathSpiral_Warning(t *testing.T) {

	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, _ := app.NewApp(ctx)
	defer core.Shutdown(ctx)
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	model := tui.NewRootModel(tuiCtx)

	// Set size < 10x10
	model.Size = tea.WindowSizeMsg{Width: 9, Height: 9}
	viewStr := model.View()
	assert.Equal(t, "!", viewStr, "Expected Death Spiral Prevention fallback '!'")
}

func TestRootModel_MinimumSize_Warning(t *testing.T) {

	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, _ := app.NewApp(ctx)
	defer core.Shutdown(ctx)
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	model := tui.NewRootModel(tuiCtx)

	// Set size < 120x24 but > 10x10
	model.Size = tea.WindowSizeMsg{Width: 100, Height: 20}
	viewStr := model.View()
	assert.Contains(t, viewStr, "Terminal is too small", "Expected minimum size warning")
}

func TestRootModel_UIAnimations_Toggle(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	cfg := types.DefaultSettingsConfig()
	cfg.UIAnimations = true
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, Settings: cfg}
	model := tui.NewRootModel(tuiCtx)

	// 1. Initial state: Animating should be true
	assert.True(t, model.Animating)

	// 2. Normal BlinkMsg toggles colon and reschedules ticker
	newModel, cmd := model.Update(types.BlinkMsg{})
	rootModel := newModel.(*tui.RootModel)
	assert.False(t, rootModel.ColonVisible)
	assert.NotNil(t, cmd, "Expected CmdTickBlink to be returned when Animating is true")

	// 3. Disable UIAnimations via SettingsChangedMsg
	cfg.UIAnimations = false
	newModel, _ = rootModel.Update(types.SettingsChangedMsg{Config: cfg})
	rootModel = newModel.(*tui.RootModel)
	assert.False(t, rootModel.Animating)
	assert.True(t, rootModel.ColonVisible, "Colon should remain visible when animations disabled")
	assert.True(t, rootModel.Header.Logo.ColonVisible)

	// 4. When Animating is false, BlinkMsg should NOT reschedule ticker and keep colon visible
	newModel, cmd = rootModel.Update(types.BlinkMsg{})
	rootModel = newModel.(*tui.RootModel)
	assert.Nil(t, cmd, "Expected no ticker cmd when Animating is false")
	assert.True(t, rootModel.ColonVisible)

	// 5. Re-enable UIAnimations via SettingsChangedMsg
	cfg.UIAnimations = true
	newModel, cmd = rootModel.Update(types.SettingsChangedMsg{Config: cfg})
	rootModel = newModel.(*tui.RootModel)
	assert.True(t, rootModel.Animating)
	assert.NotNil(t, cmd, "Expected ticker to be rescheduled when Animating re-enabled")
}

func TestRootModel_Theme_Switching(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)
	defer types.ApplyTheme(0)

	cfg := types.DefaultSettingsConfig()
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, Settings: cfg}
	model := tui.NewRootModel(tuiCtx)

	// Default theme should be active
	assert.Equal(t, "Default (Adaptive)", types.DefaultTheme.Name)

	// Switch to Dracula (2) via SettingsChangedMsg
	cfg.ActiveTheme = 2
	model.Update(types.SettingsChangedMsg{Config: cfg})
	assert.Equal(t, "Dracula", types.DefaultTheme.Name)

	// Switch to Monokai (3) via SettingsChangedMsg
	cfg.ActiveTheme = 3
	model.Update(types.SettingsChangedMsg{Config: cfg})
	assert.Equal(t, "Monokai", types.DefaultTheme.Name)

	// Switch to Ocean Dark (1) via SettingsChangedMsg
	cfg.ActiveTheme = 1
	model.Update(types.SettingsChangedMsg{Config: cfg})
	assert.Equal(t, "Ocean Dark", types.DefaultTheme.Name)
}

func TestRootModel_Settings_Persistence(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	cfg := types.DefaultSettingsConfig()
	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, Settings: cfg}
	model := tui.NewRootModel(tuiCtx)

	// Update settings
	cfg.ActiveTheme = 2
	cfg.RequestTimeout = 45 * time.Second
	cfg.FollowRedirects = false
	cfg.UIAnimations = false
	cfg.GlobalCORS = false
	cfg.StrictOpenAPIValidation = false

	_, cmd := model.Update(types.SettingsChangedMsg{Config: cfg})
	require.NotNil(t, cmd, "Expected batch command on SettingsChangedMsg")

	// Execute batch commands to trigger the async SaveSettingsConfig
	batchMsg := cmd()
	if subCmds, ok := batchMsg.(tea.BatchMsg); ok {
		for _, subCmd := range subCmds {
			if subCmd != nil {
				subCmd()
			}
		}
	}

	// Verify persistence in Core
	savedBytes, err := core.LoadSettingsConfig(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, savedBytes)

	loadedCfg, err := types.DeserializeSettingsConfig(savedBytes)
	require.NoError(t, err)
	assert.Equal(t, 2, loadedCfg.ActiveTheme)
	assert.Equal(t, 45*time.Second, loadedCfg.RequestTimeout)
	assert.False(t, loadedCfg.FollowRedirects)
	assert.False(t, loadedCfg.UIAnimations)
	assert.False(t, loadedCfg.GlobalCORS)
	assert.False(t, core.GlobalCORSEnabled())
	assert.False(t, loadedCfg.StrictOpenAPIValidation)
	assert.False(t, core.StrictValidationEnabled())
}

// Leak prevention check
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
