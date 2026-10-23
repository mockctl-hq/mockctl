package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// CmdReloadEngine wraps the backend app.ReloadEngine call in an asynchronous tea.Cmd.
func CmdReloadEngine(tuiCtx types.TUIContext, projectName string) tea.Cmd {
	return func() tea.Msg {
		err := tuiCtx.Core.ReloadEngine(tuiCtx.Ctx, projectName)
		if err != nil {
			return types.ShowToastMsg{Message: "Failed to reload engine: " + err.Error(), IsError: true}
		}
		return types.ShowToastMsg{Message: "Engine reloaded successfully", IsError: false}
	}
}

// CmdTickBlink sends a BlinkMsg every 500ms to animate the logo colon.
func CmdTickBlink() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return types.BlinkMsg{}
	})
}

// CmdWaitForTelemetry reads a single event from the broker's raw channel and returns it as a tea.Msg.
// This is called recursively in the Update loop to process continuous streams without blocking.
func CmdWaitForTelemetry(ctx context.Context, ch <-chan runtime.Event) tea.Cmd {
	return func() tea.Msg {
		select {
		case <-ctx.Done():
			return nil
		case payload, ok := <-ch:
			if !ok {
				return nil
			}
			return types.TelemetryMsg{Payload: payload}
		}
	}
}
