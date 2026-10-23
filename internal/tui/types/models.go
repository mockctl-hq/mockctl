package types

import (
	tea "github.com/charmbracelet/bubbletea"
)

// FocusRingEnum tracks which major UI panel has keyboard focus.
type FocusRingEnum int

const (
	FocusLeftPanel FocusRingEnum = iota
	FocusWorkspace
	FocusApiTester
	FocusRightPanel
	FocusOverlay
)

// OverlayStateEnum tracks which full-screen overlay is active.
type OverlayStateEnum int

const (
	OverlayNone OverlayStateEnum = iota
	OverlayStateInspector
	OverlayHistory
	OverlayGlobalLog
	OverlaySettings
	OverlayProfiler
	OverlayDaemonCrashed
	OverlayImportSpec
	OverlayDeleteProject
)

// RequestState represents a draft API test request in the tester panel.
type RequestState struct {
	Method              string
	Path                string
	Headers             string
	Body                string
	LastResponsePayload []byte
	LastStatusCode      int
	LastLatencyMs       int
}

// Component is the standard interface all nested BubbleTea components MUST implement.
type Component interface {
	Init(ctx TUIContext) tea.Cmd
	Update(msg tea.Msg) (Component, tea.Cmd)
	View() string
	SetFocus(focus bool)
}
