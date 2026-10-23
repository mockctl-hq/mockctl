package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *RootModel) routeToActive(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Route to Overlay if active
	if m.ActiveOverlay != types.OverlayNone {
		var cmd tea.Cmd
		switch m.ActiveOverlay {
		case types.OverlaySettings:
			_, cmd = m.Settings.Update(msg)
		case types.OverlayHistory:
			_, cmd = m.History.Update(msg)
		case types.OverlayGlobalLog:
			_, cmd = m.GlobalLog.Update(msg)
		case types.OverlayProfiler:
			_, cmd = m.Profiler.Update(msg)
		case types.OverlayStateInspector:
			_, cmd = m.StateInspector.Update(msg)
		case types.OverlayImportSpec:
			_, cmd = m.ImportSpec.Update(msg)
			if !m.ImportSpec.Active {
				m.ActiveOverlay = types.OverlayNone
			}
		case types.OverlayDaemonCrashed:
			if keyMsg, ok := msg.(tea.KeyMsg); ok && key.Matches(keyMsg, types.GlobalKeys.Restart) {
				if m.Ctx.Core != nil {
					serverErrChan := m.Ctx.Core.ServeInBackground(m.Ctx.Ctx, m.Ctx.ActivePort)
					select {
					case err := <-serverErrChan:
						if err != nil {
							m.DaemonCrashed.ShowError(err.Error())
							return m, nil
						}
					default:
						m.ActiveOverlay = types.OverlayNone
						return m, func() tea.Msg {
							return types.ShowToastMsg{Message: "Daemon restarted successfully", IsError: false}
						}
					}
				}
			}
			_, cmd = m.DaemonCrashed.Update(msg)
		case types.OverlayDeleteProject:
			_, cmd = m.DeleteProject.Update(msg)
			if !m.DeleteProject.Active {
				m.ActiveOverlay = types.OverlayNone
			}
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// Route events to active component
	if _, isWindowSize := msg.(tea.WindowSizeMsg); !isWindowSize {
		// Mouse Routing
		if mouseMsg, isMouse := msg.(tea.MouseMsg); isMouse {
			leftWidth := 5
			if m.LeftSidebarVisible {
				leftWidth = 34
			}
			rightWidth := 5
			if m.RightSidebarVisible {
				rightWidth = 34
			}

			if mouseMsg.X < leftWidth {
				if !m.LeftSidebarVisible {
					m.LeftSidebarVisible = true
					m.LeftSidebar.IsCollapsed = false
					m.FocusRing = types.FocusLeftPanel
					m.syncFocus()
					return m, func() tea.Msg { return m.Size }
				}
				_, cmd := m.LeftSidebar.Update(msg)
				cmds = append(cmds, cmd)
			} else if mouseMsg.X >= m.Size.Width-rightWidth {
				if !m.RightSidebarVisible {
					m.RightSidebarVisible = true
					m.RightSidebar.IsCollapsed = false
					m.FocusRing = types.FocusRightPanel
					m.syncFocus()
					return m, func() tea.Msg { return m.Size }
				}
				_, cmd := m.RightSidebar.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				_, cmd := m.Workspace.Update(msg)
				cmds = append(cmds, cmd)
			}
		} else {
			// Keyboard & Other Routing
			switch m.FocusRing {
			case types.FocusWorkspace, types.FocusApiTester:
				_, cmd := m.Workspace.Update(msg)
				cmds = append(cmds, cmd)
			case types.FocusLeftPanel:
				_, cmd := m.LeftSidebar.Update(msg)
				cmds = append(cmds, cmd)
			case types.FocusRightPanel:
				_, cmd := m.RightSidebar.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *RootModel) syncFocus() {
	m.LeftSidebar.IsCollapsed = !m.LeftSidebarVisible
	m.RightSidebar.IsCollapsed = !m.RightSidebarVisible

	m.Workspace.SetFocus(m.FocusRing == types.FocusWorkspace)
	m.Workspace.Editor.SetFocus(m.FocusRing == types.FocusWorkspace)
	m.Workspace.ApiTester.SetFocus(m.FocusRing == types.FocusApiTester)

	m.LeftSidebar.SetFocus(m.FocusRing == types.FocusLeftPanel)
	m.LeftSidebar.WorkspaceFocused = (m.FocusRing == types.FocusWorkspace)

	m.RightSidebar.SetFocus(m.FocusRing == types.FocusRightPanel)
	m.RightSidebar.WorkspaceFocused = (m.FocusRing == types.FocusWorkspace)

	m.Footer.SetFocusContext(
		m.FocusRing,
		m.ActiveOverlay,
		m.LeftSidebar.Search.Active,
		m.LeftSidebar.Scaffold.Active,
		m.RightSidebar.Icons.ActiveIndex,
		m.RightSidebar.IsIconsFocused,
	)
}

func (m *RootModel) getActiveProject() string {
	if len(m.Workspace.Tabs.Tabs) > 0 && m.Workspace.Tabs.ActiveIndex < len(m.Workspace.Tabs.Tabs) {
		proj := m.Workspace.Tabs.Tabs[m.Workspace.Tabs.ActiveIndex].ProjectName
		if proj != "" {
			return proj
		}
	}
	if m.RightSidebar.ActiveProject != "" {
		return m.RightSidebar.ActiveProject
	}
	if m.LeftSidebar.Explorer.ActiveIndex >= 0 && m.LeftSidebar.Explorer.ActiveIndex < len(m.LeftSidebar.Explorer.VisibleNodes) {
		node := m.LeftSidebar.Explorer.VisibleNodes[m.LeftSidebar.Explorer.ActiveIndex]
		if node.IsProject {
			return node.ID
		}
	}
	return ""
}
