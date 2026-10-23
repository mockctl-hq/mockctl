package right

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/right/icons"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/right/inner"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/right/inner/body"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type RightSidebarModel struct {
	tuiCtx           types.TUIContext
	Size             tea.WindowSizeMsg
	IsFocused        bool
	WorkspaceFocused bool
	IsIconsFocused   bool
	IsCollapsed      bool
	ActiveProject    string
	ActiveEndpoint   string

	Icons    *icons.RightIconsModel
	Header   *inner.InnerHeaderModel
	Chaos    *body.ChaosControlModel
	Snippets *body.CodeSnippetsModel
	Static   *body.OverridesModel
}

var _ types.Component = (*RightSidebarModel)(nil)

func NewRightSidebarModel(ctx types.TUIContext) *RightSidebarModel {
	m := &RightSidebarModel{
		tuiCtx:         ctx,
		IsIconsFocused: true,
		Icons:          icons.NewRightIconsModel(ctx),
		Header:         inner.NewInnerHeaderModel(ctx),
		Chaos:          body.NewChaosControlModel(ctx),
		Snippets:       body.NewCodeSnippetsModel(ctx),
		Static:         body.NewOverridesModel(ctx),
	}
	m.syncTitle()
	return m
}

func (m *RightSidebarModel) Init(ctx types.TUIContext) tea.Cmd {
	return tea.Batch(
		m.Icons.Init(ctx),
		m.Header.Init(ctx),
		m.Chaos.Init(ctx),
		m.Snippets.Init(ctx),
		m.Static.Init(ctx),
	)
}

func (m *RightSidebarModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		m.Icons.Update(tea.WindowSizeMsg{Width: 4, Height: msg.Height})
		if m.IsCollapsed {
			return m, nil
		}
		innerPanelWidth := msg.Width - 6
		if innerPanelWidth < 0 {
			innerPanelWidth = 0
		}
		innerSizeMsg := tea.WindowSizeMsg{Width: innerPanelWidth, Height: msg.Height}
		m.Header.Update(innerSizeMsg)
		m.Chaos.Update(innerSizeMsg)
		m.Snippets.Update(innerSizeMsg)
		m.Static.Update(innerSizeMsg)
		return m, nil
	case types.EndpointSelectedMsg:
		m.ActiveProject = msg.ProjectName
		m.ActiveEndpoint = msg.EndpointPath
		m.Chaos.ActiveProject = msg.ProjectName
		m.Chaos.ActiveEndpoint = msg.EndpointPath
		m.Snippets.ActiveProject = msg.ProjectName
		m.Snippets.ActiveEndpoint = msg.EndpointPath
		m.Static.ActiveProject = msg.ProjectName
		m.Static.ActiveEndpoint = msg.EndpointPath
		_, cmd1 := m.Chaos.Update(msg)
		_, cmd2 := m.Snippets.Update(msg)
		_, cmd3 := m.Static.Update(msg)
		return m, tea.Batch(cmd1, cmd2, cmd3)
	case types.ProjectSelectedMsg:
		if msg.ProjectName != "" {
			m.ActiveProject = msg.ProjectName
			m.Chaos.ActiveProject = msg.ProjectName
			m.Snippets.ActiveProject = msg.ProjectName
			m.Static.ActiveProject = msg.ProjectName
			_, cmd1 := m.Chaos.Update(msg)
			_, cmd2 := m.Snippets.Update(msg)
			_, cmd3 := m.Static.Update(msg)
			return m, tea.Batch(cmd1, cmd2, cmd3)
		}
	case types.TesterDraftChangedMsg:
		_, cmd := m.Snippets.Update(msg)
		return m, cmd
	}

	if m.IsFocused {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "esc" && !m.IsIconsFocused {
				m.IsIconsFocused = true
				m.syncFocus()
				return m, nil
			}
			if (keyMsg.String() == "enter" || keyMsg.String() == "left") && m.IsIconsFocused {
				m.IsIconsFocused = false
				m.syncFocus()
				return m, nil
			}
		}

		if m.IsIconsFocused {
			_, cmd1 := m.Icons.Update(msg)
			cmds = append(cmds, cmd1)
			m.syncTitle()
		} else {
			m.syncTitle()
			if m.Icons.ActiveIndex == 0 {
				_, cmd2 := m.Chaos.Update(msg)
				cmds = append(cmds, cmd2)
			} else if m.Icons.ActiveIndex == 1 {
				_, cmd3 := m.Snippets.Update(msg)
				cmds = append(cmds, cmd3)
			} else {
				_, cmd4 := m.Static.Update(msg)
				cmds = append(cmds, cmd4)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *RightSidebarModel) syncTitle() {
	switch m.Icons.ActiveIndex {
	case 0:
		m.Header.SetTitle("Chaos Controls")
	case 1:
		m.Header.SetTitle("Code Snippets")
	case 2:
		m.Header.SetTitle("Static Overrides")
	default:
		m.Header.SetTitle("Chaos Controls")
	}
}

func (m *RightSidebarModel) View() string {
	if m.IsCollapsed {
		m.Icons.WorkspaceFocused = m.WorkspaceFocused
		m.Icons.Update(tea.WindowSizeMsg{Width: 4, Height: m.Size.Height})
		return m.Icons.View()
	}

	borderStyle := types.GlobalStyles.UnfocusedBorder
	if m.IsFocused || m.WorkspaceFocused {
		borderStyle = types.GlobalStyles.FocusedBorder
	}

	m.syncTitle()

	// Ensure icons strip stretches full height
	m.Icons.WorkspaceFocused = false
	m.Icons.Update(tea.WindowSizeMsg{Width: 4, Height: m.Size.Height})
	iconsStr := m.Icons.View()
	headerStr := m.Header.View()
	
	var bodyStr string
	if m.Icons.ActiveIndex == 0 {
		bodyStr = m.Chaos.View()
	} else if m.Icons.ActiveIndex == 1 {
		bodyStr = m.Snippets.View()
	} else {
		bodyStr = m.Static.View()
	}

	innerPanel := lipgloss.JoinVertical(lipgloss.Left, headerStr, bodyStr)
	innerPanelWidth := m.Size.Width - 6
	if innerPanelWidth < 0 {
		innerPanelWidth = 0
	}
	innerPanel = lipgloss.NewStyle().Width(innerPanelWidth).Render(innerPanel)
	content := lipgloss.JoinHorizontal(lipgloss.Top, innerPanel, iconsStr)

	return borderStyle.
		Width(m.Size.Width - 1).
		Height(m.Size.Height).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		Render(content)
}

func (m *RightSidebarModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if !focus {
		// Reset inner focus when completely losing focus
		m.IsIconsFocused = true
	}
	m.syncFocus()
}

func (m *RightSidebarModel) syncFocus() {
	if m.IsFocused {
		m.Icons.SetFocus(m.IsIconsFocused)
		
		isBodyFocused := !m.IsIconsFocused
		if m.Icons.ActiveIndex == 0 {
			m.Chaos.SetFocus(isBodyFocused)
			m.Snippets.SetFocus(false)
			m.Static.SetFocus(false)
		} else if m.Icons.ActiveIndex == 1 {
			m.Chaos.SetFocus(false)
			m.Snippets.SetFocus(isBodyFocused)
			m.Static.SetFocus(false)
		} else {
			m.Chaos.SetFocus(false)
			m.Snippets.SetFocus(false)
			m.Static.SetFocus(isBodyFocused)
		}
	} else {
		m.Icons.SetFocus(false)
		m.Chaos.SetFocus(false)
		m.Snippets.SetFocus(false)
		m.Static.SetFocus(false)
	}
}
