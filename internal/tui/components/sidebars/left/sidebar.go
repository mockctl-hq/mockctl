package left

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/left/icons"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/left/inner"
	"github.com/mockctl-hq/mockctl/internal/tui/components/sidebars/left/inner/body"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type LeftSidebarModel struct {
	tuiCtx           types.TUIContext
	Size             tea.WindowSizeMsg
	IsFocused        bool
	WorkspaceFocused bool
	IsCollapsed      bool

	Icons    *icons.LeftIconsModel
	Header   *inner.InnerHeaderModel
	Scaffold *inner.ScaffoldInputModel
	Search   *shared.SearchModel
	Explorer *body.ProjectTreeModel
	
	ActiveInnerFocus int // 0: Search, 1: Explorer, 2: Scaffold
}

var _ types.Component = (*LeftSidebarModel)(nil)

func NewLeftSidebarModel(ctx types.TUIContext) *LeftSidebarModel {
	return &LeftSidebarModel{
		tuiCtx:   ctx,
		Icons:    icons.NewLeftIconsModel(ctx),
		Header:   inner.NewInnerHeaderModel(ctx),
		Scaffold: inner.NewScaffoldInputModel(ctx),
		Search:   shared.NewSearchModel(),
		Explorer: body.NewProjectTreeModel(ctx),
		ActiveInnerFocus: 1, // Default to explorer
	}
}

func (m *LeftSidebarModel) Init(ctx types.TUIContext) tea.Cmd {
	return tea.Batch(
		m.Icons.Init(ctx),
		m.Header.Init(ctx),
		m.Scaffold.Init(ctx),
		m.Search.Init(ctx),
		m.Explorer.Init(ctx),
	)
}

func (m *LeftSidebarModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		m.Icons.Update(tea.WindowSizeMsg{Width: 4, Height: msg.Height})
		if m.IsCollapsed {
			return m, nil
		}
		innerPanelWidth := msg.Width - 6 // 4 for icons, 1 for border, 1 for outer border
		if innerPanelWidth < 0 {
			innerPanelWidth = 0
		}
		innerSizeMsg := tea.WindowSizeMsg{Width: innerPanelWidth, Height: msg.Height - 4}
		m.Explorer.Update(tea.WindowSizeMsg{Width: innerPanelWidth, Height: msg.Height - 5})
		m.Header.Update(innerSizeMsg)
		return m, nil
	
	case tea.KeyMsg:
		if !m.IsFocused || m.IsCollapsed {
			return m, nil
		}
		
		// If Scaffold is active, route exclusively to it
		if m.Scaffold.Active {
			_, cmd := m.Scaffold.Update(msg)
			return m, cmd
		}
		
		// If Search is active, route exclusively to it
		if m.Search.Active {
			_, cmd := m.Search.Update(msg)
			return m, cmd
		}
		
		// Shortcut to trigger Scaffold project
		if msg.String() == "+" {
			m.Scaffold.Show(true) // true = create project
			return m, nil
		}
		
		// Shortcut to trigger Search
		if msg.String() == "/" {
			m.Search.Active = true
			m.Search.SetFocus(true)
			return m, nil
		}
	}

	// Normal routing
	if m.IsFocused && !m.Scaffold.Active && !m.Search.Active {
		_, cmd1 := m.Icons.Update(msg)
		_, cmd2 := m.Explorer.Update(msg)
		cmds = append(cmds, cmd1, cmd2)
	}

	return m, tea.Batch(cmds...)
}

func (m *LeftSidebarModel) View() string {
	if m.IsCollapsed {
		m.Icons.WorkspaceFocused = m.WorkspaceFocused
		m.Icons.Update(tea.WindowSizeMsg{Width: 4, Height: m.Size.Height})
		return m.Icons.View()
	}

	borderStyle := types.GlobalStyles.UnfocusedBorder
	if m.IsFocused || m.WorkspaceFocused {
		borderStyle = types.GlobalStyles.FocusedBorder
	}

	// Ensure icons strip stretches full height
	m.Icons.WorkspaceFocused = false
	m.Icons.Update(tea.WindowSizeMsg{Width: 4, Height: m.Size.Height})
	iconsStr := m.Icons.View()
	headerStr := m.Header.View()
	scaffoldStr := m.Scaffold.View()
	searchStr := m.Search.View()
	explorerStr := m.Explorer.View()

	innerStack := []string{headerStr}
	if m.Scaffold.Active {
		innerStack = append(innerStack, scaffoldStr)
	}
	if m.Search.Active {
		innerStack = append(innerStack, searchStr)
	}
	innerStack = append(innerStack, explorerStr)

	innerPanel := lipgloss.JoinVertical(lipgloss.Left, innerStack...)
	innerPanelWidth := m.Size.Width - 6
	if innerPanelWidth < 0 {
		innerPanelWidth = 0
	}
	innerPanel = lipgloss.NewStyle().Width(innerPanelWidth).Render(innerPanel)

	content := lipgloss.JoinHorizontal(lipgloss.Top, iconsStr, innerPanel)

	return borderStyle.
		Width(m.Size.Width - 1).
		Height(m.Size.Height).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		Render(content)
}

func (m *LeftSidebarModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if focus {
		m.Header.SetFocus(true)
		m.Icons.SetFocus(true)
		if m.Scaffold.Active {
			m.Scaffold.SetFocus(true)
		} else if m.Search.Active {
			m.Search.SetFocus(true)
		} else {
			m.Explorer.SetFocus(true)
		}
	} else {
		m.Header.SetFocus(false)
		m.Icons.SetFocus(false)
		m.Scaffold.SetFocus(false)
		m.Search.SetFocus(false)
		m.Explorer.SetFocus(false)
	}
}
