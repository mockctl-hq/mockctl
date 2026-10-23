//go:build integration

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

// PKS-029: Headless Integration Tests

func TestIntegration_RootModel_Headless(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Setenv("HOME", t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Boot the Master Daemon
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer func() { _ = core.Shutdown(ctx) }()

	// Start server on random port in background
	errChan := core.ServeInBackground(ctx, 8081)
	time.Sleep(100 * time.Millisecond) // Give server time to bind

	select {
	case err := <-errChan:
		t.Fatalf("Server failed to start: %v", err)
	default:
	}

	// 2. Initialize TUI Root Model
	tuiCtx := types.TUIContext{
		Ctx:        ctx,
		Core:       core,
		ActivePort: 8081,
	}

	model := tui.NewRootModel(tuiCtx)
	_ = model.Init()

	// Set terminal size
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 150, Height: 40})
	rootModel := newModel.(*tui.RootModel)

	// Verify that the view renders without panicking
	view := rootModel.View()
	assert.NotEmpty(t, view)

	// 3. Open Settings Overlay
	newModel, _ = rootModel.Update(types.OverlayStateEnum(types.OverlaySettings))
	rootModel = newModel.(*tui.RootModel)
	assert.Equal(t, types.OverlaySettings, rootModel.ActiveOverlay)

	view = rootModel.View()
	assert.Contains(t, view, "Settings")
}
