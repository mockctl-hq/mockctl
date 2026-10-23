package tui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays"
	"github.com/mockctl-hq/mockctl/internal/tui/components/overlays/system_prompts"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/left/inner"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace/tester/protocols/http"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// Update handles incoming messages and routes them to the appropriate child components.
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Size = msg
		// Auto-Collapse API Tester if terminal height drops below safe threshold (landscape)
		if m.Size.Height < 30 {
			m.ApiTesterExpanded = false
		}
		// Route size down to child components
		// Header is 4 lines (3 text + 1 border), Footer is 2 lines (1 text + 1 border). Total = 6
		middleHeight := m.Size.Height - 6
		if middleHeight < 0 {
			middleHeight = 0
		}
		leftWidth := 5
		if m.LeftSidebarVisible {
			leftWidth = 34
		}
		rightWidth := 5
		if m.RightSidebarVisible {
			rightWidth = 34
		}
		centerWidth := m.Size.Width - leftWidth - rightWidth
		if centerWidth < 0 {
			centerWidth = 0
		}
		innerMsg := tea.WindowSizeMsg{Width: centerWidth, Height: middleHeight}
		
		_, cmd1 := m.Workspace.Update(innerMsg)
		_, cmd2 := m.Header.Update(msg)
		_, cmd3 := m.Footer.Update(msg)
		cmds = append(cmds, cmd1, cmd2, cmd3)
		
		m.LeftSidebar.IsCollapsed = !m.LeftSidebarVisible
		m.RightSidebar.IsCollapsed = !m.RightSidebarVisible
		leftMsg := tea.WindowSizeMsg{Width: leftWidth, Height: middleHeight}
		_, cmd4 := m.LeftSidebar.Update(leftMsg)
		cmds = append(cmds, cmd4)

		rightMsg := tea.WindowSizeMsg{Width: rightWidth, Height: middleHeight}
		_, cmd5 := m.RightSidebar.Update(rightMsg)
		cmds = append(cmds, cmd5)
		
		// Pass WindowSizeMsg to all overlays so they know the screen dimensions
		m.GlobalLog.Update(msg)
		m.Settings.Update(msg)
		m.History.Update(msg)
		m.Profiler.Update(msg)
		m.StateInspector.Update(msg)
		m.ScreenWarning.Update(msg)
		m.DaemonCrashed.Update(msg)
		m.DeleteProject.Update(msg)
		
		return m, tea.Batch(cmds...)

	case types.BlinkMsg:
		if !m.Animating {
			m.ColonVisible = true
			m.Header.Logo.ColonVisible = true
			return m, nil
		}
		m.ColonVisible = !m.ColonVisible
		m.Header.Update(msg)
		if m.ActiveOverlay == types.OverlayGlobalLog {
			m.GlobalLog.RefreshFromDisk()
		}
		return m, CmdTickBlink()

	case types.TelemetryMsg:
		m.History.Update(msg)
		if m.telemetryChan != nil {
			reqCtx := m.Ctx.Ctx
			if reqCtx == nil {
				reqCtx = context.Background()
			}
			return m, CmdWaitForTelemetry(reqCtx, m.telemetryChan)
		}
		return m, nil

	case types.ShowToastMsg, shared.HideToastMsg:
		_, cmd := m.Header.Update(msg)
		return m, cmd

	case types.DaemonCrashedMsg:
		m.ActiveOverlay = types.OverlayDaemonCrashed
		m.DaemonCrashed.ShowError(msg.Error)
		return m, nil

	case types.OverlayStateEnum:
		m.ActiveOverlay = msg
		if msg == types.OverlaySettings {
			m.Settings.Active = true
		}
		if msg == types.OverlayGlobalLog {
			m.GlobalLog.RefreshFromDisk()
		}
		if msg == types.OverlayStateInspector {
			m.StateInspector.RefreshState(m.getActiveProject())
		}
		return m, nil

	case types.SettingsChangedMsg:
		// Broadcast to all components that are wired to SettingsConfig.
		// Each component is responsible for reacting only to settings it cares about.
		m.ActiveOverlay = types.OverlayNone
		var settingsCmds []tea.Cmd
		_, c1 := m.History.Update(msg)
		_, c2 := m.Workspace.Update(msg)
		_, c3 := m.Header.Update(msg)
		settingsCmds = append(settingsCmds, c1, c2, c3)

		if msg.Config != nil {
			if m.Ctx.Settings != nil {
				*m.Ctx.Settings = *msg.Config
			}

			// Apply active theme immediately across the entire TUI
			types.ApplyTheme(msg.Config.ActiveTheme)

			if msg.Config.UIAnimations {
				if !m.Animating {
					m.Animating = true
					settingsCmds = append(settingsCmds, CmdTickBlink())
				}
			} else {
				m.Animating = false
				m.ColonVisible = true
				m.Header.Logo.ColonVisible = true
			}

			// Persist updated settings to BBolt system.db asynchronously via Core
			if m.Ctx.Core != nil {
				coreApp := m.Ctx.Core
				cfgCopy := *msg.Config
				settingsCmds = append(settingsCmds, func() tea.Msg {
					if data, err := cfgCopy.Serialize(); err == nil {
						_ = coreApp.SaveSettingsConfig(context.Background(), data)
					}
					return nil
				})
			}
		}

		_, c4 := m.ImportSpec.Update(msg)
		settingsCmds = append(settingsCmds, c4)

		return m, tea.Batch(settingsCmds...)

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case inner.ScaffoldSubmitMsg:
		if msg.IsProject {
			defaultSpec := fmt.Sprintf(`openapi: 3.0.0
info:
  title: %s API
  version: 1.0.0
paths:
  /ping:
    get:
      summary: Health check
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
`, msg.Value)
			
			err := m.Ctx.Core.CreateProject(context.Background(), msg.Value, []byte(defaultSpec))
			if err == nil {
				// Refresh the sidebar project tree
				m.LeftSidebar.Explorer.RefreshFromCore()
			}
		}
		return m, nil

	case types.TesterToggleMsg:
		m.ApiTesterExpanded = !m.ApiTesterExpanded
		
		if m.ApiTesterExpanded {
			m.FocusRing = types.FocusApiTester
		} else if m.FocusRing == types.FocusApiTester {
			m.FocusRing = types.FocusWorkspace
		}
		m.syncFocus()

		// Send message down to Workspace so it can expand its tester panel
		_, cmd := m.Workspace.Update(msg)
		return m, cmd
		
	case types.ProjectStructureChangedMsg:
		m.LeftSidebar.Explorer.RefreshFromCore()
		return m, nil
		
	case types.EndpointSelectedMsg, types.ProjectSelectedMsg:
		// Broadcast to all major panels regardless of focus
		_, cmd1 := m.Workspace.Update(msg)
		_, cmd2 := m.LeftSidebar.Update(msg)
		_, cmd3 := m.RightSidebar.Update(msg)
		return m, tea.Batch(cmd1, cmd2, cmd3)
		
	case overlays.FilesLoadedMsg:
		_, cmd := m.ImportSpec.Update(msg)
		return m, cmd

	case types.TesterDraftChangedMsg:
		_, cmd := m.RightSidebar.Update(msg)
		return m, cmd

	case spinner.TickMsg, http.ResponseMsg:
		_, cmd := m.Workspace.Update(msg)
		return m, cmd

	case system_prompts.DeleteProjectMsg:
		m.ActiveOverlay = types.OverlayNone
		if m.Ctx.Core != nil && msg.ProjectName != "" {
			err := m.Ctx.Core.DeleteProject(context.Background(), msg.ProjectName)
			if err != nil {
				return m, func() tea.Msg {
					return types.ShowToastMsg{Message: "Failed to delete: " + err.Error(), IsError: true}
				}
			}
			cmdTab := m.Workspace.Tabs.CloseProjectTabs(msg.ProjectName)
			m.LeftSidebar.Explorer.RefreshFromCore()
			toastCmd := func() tea.Msg {
				return types.ShowToastMsg{Message: "Project deleted: " + msg.ProjectName, IsError: false}
			}
			return m, tea.Batch(cmdTab, toastCmd)
		}
		return m, nil
	}
	
	return m.routeToActive(msg)
}

