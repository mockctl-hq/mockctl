package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// handleKeyMsg handles all global keyboard shortcuts and focus navigation.
func (m *RootModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global Quit Command (Ctrl+C / Ctrl+Q / q if not in text input)
	if key.Matches(msg, types.GlobalKeys.QuitContext) {
		return m, tea.Quit
	}

	if m.ActiveOverlay == types.OverlayNone {
		if key.Matches(msg, types.GlobalKeys.Quit) {
			return m, tea.Quit
		}

		// Shortcut to open overlays
		if key.Matches(msg, types.GlobalKeys.Settings) {
			m.ActiveOverlay = types.OverlaySettings
			m.Settings.Active = true
			return m, nil
		}
		if key.Matches(msg, types.GlobalKeys.History) {
			m.ActiveOverlay = types.OverlayHistory
			return m, nil
		}
		if key.Matches(msg, types.GlobalKeys.GlobalLog) {
			m.ActiveOverlay = types.OverlayGlobalLog
			m.GlobalLog.RefreshFromDisk()
			return m, nil
		}
		if key.Matches(msg, types.GlobalKeys.Profiler) {
			m.ActiveOverlay = types.OverlayProfiler
			return m, nil
		}
		if key.Matches(msg, types.GlobalKeys.StateInspector) {
			m.ActiveOverlay = types.OverlayStateInspector
			m.StateInspector.RefreshState(m.getActiveProject())
			return m, nil
		}
		if key.Matches(msg, types.GlobalKeys.Delete) {
			activeProj := m.getActiveProject()
			if activeProj != "" {
				m.ActiveOverlay = types.OverlayDeleteProject
				m.DeleteProject.Show(activeProj)
				return m, nil
			}
		}
		if key.Matches(msg, types.GlobalKeys.ImportSpec) {
			activeProject := ""
			if len(m.Workspace.Tabs.Tabs) > 0 {
				activeProject = m.Workspace.Tabs.Tabs[m.Workspace.Tabs.ActiveIndex].ProjectName
			}
			if activeProject != "" {
				m.ActiveOverlay = types.OverlayImportSpec
				m.ImportSpec.SetProject(activeProject)
				m.ImportSpec.Active = true
			}
			return m, nil
		}

		if key.Matches(msg, types.GlobalKeys.ToggleApiTester) {
			return m, func() tea.Msg { return types.TesterToggleMsg{} }
		}

		if key.Matches(msg, types.GlobalKeys.ResetState) {
			activeProj := m.getActiveProject()
			if activeProj != "" && m.Ctx.Core != nil {
				err := m.Ctx.Core.ResetState(m.Ctx.Ctx, activeProj)
				if err != nil {
					return m, func() tea.Msg {
						return types.ShowToastMsg{Message: "Failed to reset state: " + err.Error(), IsError: true}
					}
				}
				if m.ActiveOverlay == types.OverlayStateInspector {
					m.StateInspector.RefreshState(activeProj)
				}
				return m, func() tea.Msg {
					return types.ShowToastMsg{Message: "Memory state reset for " + activeProj, IsError: false}
				}
			}
		}

		if key.Matches(msg, types.GlobalKeys.ToggleLeft) {
			m.LeftSidebarVisible = !m.LeftSidebarVisible
			if !m.LeftSidebarVisible && m.FocusRing == types.FocusLeftPanel {
				m.FocusRing = types.FocusWorkspace
			}
			m.syncFocus()
			return m, func() tea.Msg { return m.Size }
		}

		if key.Matches(msg, types.GlobalKeys.ToggleRight) {
			m.RightSidebarVisible = !m.RightSidebarVisible
			if !m.RightSidebarVisible && m.FocusRing == types.FocusRightPanel {
				m.FocusRing = types.FocusWorkspace
			}
			m.syncFocus()
			return m, func() tea.Msg { return m.Size }
		}

		if key.Matches(msg, types.GlobalKeys.NextTab, types.GlobalKeys.PrevTab, types.GlobalKeys.CloseTab) {
			oldTabCount := len(m.Workspace.Tabs.Tabs)
			// Send directly to Workspace (which handles Tabs)
			_, cmd := m.Workspace.Tabs.Update(msg)
			if oldTabCount != len(m.Workspace.Tabs.Tabs) {
				// Force workspace to recalculate heights
				m.Workspace.Update(m.Workspace.Size)
			}
			return m, cmd
		}

		// Tab Focus Routing
		if key.Matches(msg, types.GlobalKeys.TabForward) {
			if m.FocusRing == types.FocusApiTester {
				if m.Workspace.ApiTester.HttpTester.AdvanceFocus() {
					return m, nil
				}
			}

			for i := 0; i < 4; i++ {
				m.FocusRing = (m.FocusRing + 1) % 4
				if m.FocusRing == types.FocusApiTester && !m.ApiTesterExpanded {
					continue
				}
				if m.FocusRing == types.FocusLeftPanel && !m.LeftSidebarVisible {
					continue
				}
				if m.FocusRing == types.FocusRightPanel && !m.RightSidebarVisible {
					continue
				}
				break
			}
			m.syncFocus()
			return m, nil
		}
		if key.Matches(msg, types.GlobalKeys.TabBackward) {
			if m.FocusRing == types.FocusApiTester {
				if m.Workspace.ApiTester.HttpTester.ReverseFocus() {
					return m, nil
				}
			}

			for i := 0; i < 4; i++ {
				if m.FocusRing == 0 {
					m.FocusRing = types.FocusRightPanel
				} else {
					m.FocusRing--
				}
				if m.FocusRing == types.FocusApiTester && !m.ApiTesterExpanded {
					continue
				}
				if m.FocusRing == types.FocusLeftPanel && !m.LeftSidebarVisible {
					continue
				}
				if m.FocusRing == types.FocusRightPanel && !m.RightSidebarVisible {
					continue
				}
				break
			}
			m.syncFocus()
			return m, nil
		}
	}

	// Route unhandled keypress to active component or overlay
	return m.routeToActive(msg)
}
