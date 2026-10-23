package workspace

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// Tab struct represents a single open file tab
type Tab struct {
	ProjectName string
	FileName    string
	IsDirty     bool
	IsOrphaned  bool
}

// TabBarModel manages the display of multiple active open file tabs.
type TabBarModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	Tabs        []Tab
	ActiveIndex int
	IsFocused   bool
}

// Ensure interface compliance
var _ types.Component = (*TabBarModel)(nil)

func NewTabBarModel(ctx types.TUIContext) *TabBarModel {
	return &TabBarModel{
		tuiCtx: ctx,
		Tabs:   make([]Tab, 0),
	}
}

func (m *TabBarModel) Init(ctx types.TUIContext) tea.Cmd {
	m.tuiCtx = ctx
	return nil
}

func (m *TabBarModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	case types.ProjectSelectedMsg:
		if msg.ProjectName == "" {
			return m, nil
		}
		// Check if tab already exists
		found := false
		for i, tab := range m.Tabs {
			if tab.ProjectName == msg.ProjectName {
				m.ActiveIndex = i
				found = true
				if msg.FileName != "" {
					m.Tabs[i].FileName = msg.FileName
				}
				break
			}
		}
		if !found {
			fileName := msg.FileName
			if fileName == "" {
				fileName = "api.yaml" // Fallback
			}
			// Add new tab
			m.Tabs = append(m.Tabs, Tab{
				ProjectName: msg.ProjectName,
				FileName:    fileName,
			})
			m.ActiveIndex = len(m.Tabs) - 1
		}
	case tea.KeyMsg:
		if len(m.Tabs) > 1 {
			if key.Matches(msg, types.GlobalKeys.NextTab) {
				m.ActiveIndex = (m.ActiveIndex + 1) % len(m.Tabs)
				return m, func() tea.Msg { return types.ProjectSelectedMsg{ProjectName: m.Tabs[m.ActiveIndex].ProjectName, FileName: m.Tabs[m.ActiveIndex].FileName} }
			} else if key.Matches(msg, types.GlobalKeys.PrevTab) {
				m.ActiveIndex--
				if m.ActiveIndex < 0 {
					m.ActiveIndex = len(m.Tabs) - 1
				}
				return m, func() tea.Msg { return types.ProjectSelectedMsg{ProjectName: m.Tabs[m.ActiveIndex].ProjectName, FileName: m.Tabs[m.ActiveIndex].FileName} }
			}
		}

		if len(m.Tabs) > 0 {
			if key.Matches(msg, types.GlobalKeys.CloseTab) {
				m.Tabs = append(m.Tabs[:m.ActiveIndex], m.Tabs[m.ActiveIndex+1:]...)
				if len(m.Tabs) > 0 {
					if m.ActiveIndex >= len(m.Tabs) {
						m.ActiveIndex = len(m.Tabs) - 1
					}
					return m, func() tea.Msg { return types.ProjectSelectedMsg{ProjectName: m.Tabs[m.ActiveIndex].ProjectName, FileName: m.Tabs[m.ActiveIndex].FileName} }
				} else {
					m.ActiveIndex = 0
					return m, func() tea.Msg { return types.ProjectSelectedMsg{ProjectName: "", FileName: ""} }
				}
			}
		}
	}
	return m, nil
}

// CloseProjectTabs removes all open tabs belonging to a specific project.
func (m *TabBarModel) CloseProjectTabs(projectName string) tea.Cmd {
	var remaining []Tab
	for _, t := range m.Tabs {
		if t.ProjectName != projectName {
			remaining = append(remaining, t)
		}
	}
	m.Tabs = remaining
	if len(m.Tabs) > 0 {
		if m.ActiveIndex >= len(m.Tabs) {
			m.ActiveIndex = len(m.Tabs) - 1
		}
		return func() tea.Msg {
			return types.ProjectSelectedMsg{ProjectName: m.Tabs[m.ActiveIndex].ProjectName, FileName: m.Tabs[m.ActiveIndex].FileName}
		}
	}
	m.ActiveIndex = 0
	return func() tea.Msg {
		return types.ProjectSelectedMsg{ProjectName: "", FileName: ""}
	}
}

func (m *TabBarModel) View() string {
	if len(m.Tabs) == 0 {
		return ""
	}

	var renderedTabs []string

	for i, tab := range m.Tabs {
		tabText := tab.FileName
		
		// Extract just the filename for the tab so it doesn't take up too much space
		parts := strings.Split(tabText, "/")
		if len(parts) > 0 {
			tabText = parts[len(parts)-1]
		}
		
		if tab.IsDirty {
			tabText += "*"
		}
		if tab.IsOrphaned {
			tabText += " 🟡"
		}

		var style lipgloss.Style
		if i == m.ActiveIndex {
			if m.IsFocused {
				style = lipgloss.NewStyle().
					Background(types.DefaultTheme.Primary).
					Foreground(types.DefaultTheme.Bg).
					Border(lipgloss.NormalBorder(), false, false, true, false).
					BorderForeground(types.DefaultTheme.Primary).
					Padding(0, 1)
			} else {
				style = lipgloss.NewStyle().
					Background(types.DefaultTheme.Bg).
					Foreground(types.DefaultTheme.Primary).
					Border(lipgloss.NormalBorder(), false, false, true, false).
					BorderForeground(types.DefaultTheme.Primary).
					Padding(0, 1)
			}
		} else {
			style = lipgloss.NewStyle().
				Background(types.DefaultTheme.Bg).
				Foreground(types.DefaultTheme.Dim).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(types.DefaultTheme.Dim).
				Padding(0, 1)
		}

		renderedTabs = append(renderedTabs, style.Render("[ "+tabText+" ✖ ]"))
	}

	// Join all tabs horizontally
	allTabsStr := lipgloss.JoinHorizontal(lipgloss.Bottom, renderedTabs...)
	
	// Add bottom border to the rest of the empty space
	emptySpaceWidth := m.Size.Width - lipgloss.Width(allTabsStr)
	if emptySpaceWidth > 0 {
		emptySpace := lipgloss.NewStyle().
			Width(emptySpaceWidth).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(types.DefaultTheme.Dim).
			Render("")
		allTabsStr = lipgloss.JoinHorizontal(lipgloss.Bottom, allTabsStr, emptySpace)
	}

	// Generate Toolbar
	var toolbarStr string
	if len(m.Tabs) > 0 {
		idx := m.ActiveIndex
		if idx >= len(m.Tabs) {
			idx = len(m.Tabs) - 1
		}
		if idx < 0 {
			idx = 0
		}
		activeTab := m.Tabs[idx]
		leftToolbar := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text).Padding(0, 1).Render(activeTab.ProjectName + " / " + activeTab.FileName)
		rightToolbar := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Padding(0, 1).Render("[Import: Ctrl+O]")
		
		padding := m.Size.Width - lipgloss.Width(leftToolbar) - lipgloss.Width(rightToolbar)
		if padding < 0 {
			padding = 0
		}
		
		spaces := ""
		for i := 0; i < padding; i++ {
			spaces += " "
		}
		
		toolbarContent := leftToolbar + spaces + rightToolbar
		toolbarStr = lipgloss.NewStyle().
			Width(m.Size.Width).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(types.DefaultTheme.Dim).
			Render(toolbarContent)
	}

	finalStr := lipgloss.JoinVertical(lipgloss.Left, allTabsStr, toolbarStr)
	return lipgloss.NewStyle().Width(m.Size.Width).Render(finalStr)
}

func (m *TabBarModel) SetFocus(focus bool) {
	m.IsFocused = focus
}
