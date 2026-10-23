package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/tui/components/footer"
	"github.com/mockctl-hq/mockctl/internal/tui/components/header"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/history"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/inspector"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/logs"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/profiler"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/header_tools/settings"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/system_prompts"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/left"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/right"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// RootModel is the top-level BubbleTea model.
// It manages the global layout, focus, and delegates events to child components.
type RootModel struct {
	Ctx                 types.TUIContext
	Size                tea.WindowSizeMsg
	FocusRing           types.FocusRingEnum
	ActiveOverlay       types.OverlayStateEnum
	ApiTesterExpanded   bool
	ColonVisible        bool
	Animating           bool
	LeftSidebarVisible  bool
	RightSidebarVisible bool
	telemetryChan       chan runtime.Event

	Workspace      *workspace.WorkspaceModel
	StateInspector *inspector.StateInspectorModel
	GlobalLog      *logs.GlobalLogsModel
	Settings       *settings.SettingsModel
	Profiler       *profiler.GoProfilerModel
	History        *history.HistoryModel
	ImportSpec     *overlays.ImportSpecModel
	ScreenWarning  *system_prompts.ScreenWarningModel
	DaemonCrashed  *system_prompts.DaemonCrashedModel
	DeleteProject  *system_prompts.DeleteProjectModel

	Header       *header.HeaderModel
	Footer       *footer.FooterModel
	LeftSidebar  *left.LeftSidebarModel
	RightSidebar *right.RightSidebarModel
}

var _ tea.Model = (*RootModel)(nil)

// NewRootModel initializes the root TUI model with the global context.
func NewRootModel(ctx types.TUIContext) *RootModel {
	animating := true
	if ctx.Settings != nil {
		animating = ctx.Settings.UIAnimations
	}

	var telChan chan runtime.Event
	if ctx.Core != nil && ctx.Core.EventBroker() != nil {
		telChan = ctx.Core.EventBroker().SubscribeRaw(runtime.FilterOptions{})
	}

	m := &RootModel{
		Ctx:                 ctx,
		FocusRing:           types.FocusWorkspace,
		ActiveOverlay:       types.OverlayNone,
		ApiTesterExpanded:   false,
		ColonVisible:        true,
		Animating:           animating,
		LeftSidebarVisible:  true,
		RightSidebarVisible: true,
		telemetryChan:       telChan,
		Workspace:           workspace.NewWorkspaceModel(ctx),
		StateInspector:      inspector.NewStateInspectorModel(ctx),
		GlobalLog:           logs.NewGlobalLogsModel(ctx),
		Settings:            settings.NewSettingsModel(ctx),
		Profiler:            profiler.NewGoProfilerModel(ctx),
		History:             history.NewHistoryModel(ctx),
		ImportSpec:          overlays.NewImportSpecModel(ctx),
		ScreenWarning:       system_prompts.NewScreenWarningModel(ctx),
		DaemonCrashed:       system_prompts.NewDaemonCrashedModel(ctx),
		DeleteProject:       system_prompts.NewDeleteProjectModel(ctx),
		Header:              header.NewHeaderModel(ctx),
		Footer:              footer.NewFooterModel(ctx),
		LeftSidebar:         left.NewLeftSidebarModel(ctx),
		RightSidebar:        right.NewRightSidebarModel(ctx),
	}
	m.syncFocus()
	return m
}

// Init runs when the BubbleTea program starts.
func (m *RootModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		tea.SetWindowTitle("Mock:ctl Studio"),
		m.Workspace.Init(m.Ctx),
		m.StateInspector.Init(m.Ctx),
		m.GlobalLog.Init(m.Ctx),
		m.Settings.Init(m.Ctx),
		m.Profiler.Init(m.Ctx),
		m.History.Init(m.Ctx),
		m.ImportSpec.Init(m.Ctx),
		m.ScreenWarning.Init(m.Ctx),
		m.DaemonCrashed.Init(m.Ctx),
		m.DeleteProject.Init(m.Ctx),
		m.Header.Init(m.Ctx),
		m.Footer.Init(m.Ctx),
		m.LeftSidebar.Init(m.Ctx),
		m.RightSidebar.Init(m.Ctx),
	}
	if m.telemetryChan != nil {
		reqCtx := m.Ctx.Ctx
		if reqCtx == nil {
			reqCtx = context.Background()
		}
		cmds = append(cmds, CmdWaitForTelemetry(reqCtx, m.telemetryChan))
	}
	if m.Animating {
		cmds = append(cmds, CmdTickBlink())
	}
	return tea.Batch(cmds...)
}

// Close unsubscribes raw telemetry streams to prevent memory leaks.
func (m *RootModel) Close() {
	if m.telemetryChan != nil && m.Ctx.Core != nil && m.Ctx.Core.EventBroker() != nil {
		m.Ctx.Core.EventBroker().UnsubscribeRaw(m.telemetryChan)
		m.telemetryChan = nil
	}
}
