package workspace

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace/tester"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace/tester/protocols/http"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// WorkspaceModel represents the Middle Workspace Container.
type WorkspaceModel struct {
	tuiCtx            types.TUIContext
	Size              tea.WindowSizeMsg
	IsFocused         bool
	ApiTesterExpanded bool

	// Children
	Tabs      *TabBarModel
	Editor    *YamlEditorModel
	ApiTester *tester.ApiTesterModel
}

// Ensure interface compliance
var _ types.Component = (*WorkspaceModel)(nil)

func NewWorkspaceModel(ctx types.TUIContext) *WorkspaceModel {
	return &WorkspaceModel{
		tuiCtx:    ctx,
		Tabs:      NewTabBarModel(ctx),
		Editor:    NewYamlEditorModel(ctx),
		ApiTester: tester.NewApiTesterModel(ctx),
	}
}

func (m *WorkspaceModel) Init(ctx types.TUIContext) tea.Cmd {
	m.tuiCtx = ctx
	return tea.Batch(
		m.Tabs.Init(ctx),
		m.Editor.Init(ctx),
		m.ApiTester.Init(ctx),
	)
}

func (m *WorkspaceModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg

		tabSize := tea.WindowSizeMsg{Width: msg.Width, Height: 1}
		_, cmd = m.Tabs.Update(tabSize)
		cmds = append(cmds, cmd)

		testerHeight := 0
		if m.ApiTesterExpanded {
			testerHeight = 14 // Fixed height for tester panel when expanded
			if testerHeight > msg.Height-3 {
				testerHeight = msg.Height - 3 // Ensure editor has at least 1 line
			}
		}

		tabHeight := 0
		if len(m.Tabs.Tabs) > 0 {
			tabHeight = 4
		}

		editorSize := tea.WindowSizeMsg{Width: msg.Width, Height: msg.Height - tabHeight - 2 - testerHeight} // 2 is for statusBar (text + top divider)
		if editorSize.Height < 0 {
			editorSize.Height = 0
		}
		_, cmd = m.Editor.Update(editorSize)
		cmds = append(cmds, cmd)

		if m.ApiTesterExpanded {
			_, cmd = m.ApiTester.Update(tea.WindowSizeMsg{Width: msg.Width, Height: testerHeight})
			cmds = append(cmds, cmd)
		}

	case types.TesterToggleMsg:
		m.ApiTesterExpanded = !m.ApiTesterExpanded
		// Force resize to recompute heights
		m.Update(m.Size)
		return m, nil
	case types.SettingsChangedMsg:
		_, cmd1 := m.Editor.Update(msg)
		_, cmd2 := m.ApiTester.Update(msg)
		return m, tea.Batch(cmd1, cmd2)
	case types.EndpointSelectedMsg:
		_, cmd1 := m.Editor.Update(msg)
		_, cmd2 := m.ApiTester.Update(msg)
		return m, tea.Batch(cmd1, cmd2)
	case types.ProjectSelectedMsg:
		oldTabCount := len(m.Tabs.Tabs)
		_, cmdTab := m.Tabs.Update(msg)
		if oldTabCount != len(m.Tabs.Tabs) {
			m.Update(m.Size)
		}
		_, cmdEditor := m.Editor.Update(msg)
		return m, tea.Batch(cmdTab, cmdEditor)
	case spinner.TickMsg, http.ResponseMsg:
		_, cmd := m.ApiTester.Update(msg)
		return m, cmd
	}

	// Route events
	if _, isWindowSize := msg.(tea.WindowSizeMsg); !isWindowSize {
		if mouseMsg, isMouse := msg.(tea.MouseMsg); isMouse {
			// Mouse Routing
			if m.ApiTesterExpanded && mouseMsg.Y >= m.Editor.Viewport.Height {
				// Mouse is over ApiTester
				_, cmd = m.ApiTester.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				// Mouse is over Editor
				_, cmd = m.Editor.Update(msg)
				cmds = append(cmds, cmd)
			}
		} else {
			// Keyboard Routing
			if m.ApiTester.IsFocused {
				_, cmd = m.ApiTester.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				_, cmd = m.Editor.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *WorkspaceModel) View() string {
	// 3. Cursor Status Bar (Owned by workspace container or editor)
	currentLine := m.Editor.Viewport.YOffset + 1
	totalLines := m.Editor.Viewport.TotalLineCount()
	pct := int(m.Editor.Viewport.ScrollPercent() * 100)
	statusText := fmt.Sprintf(" Ln %d/%d (%d%%) • OpenAPI YAML Viewer", currentLine, totalLines, pct)

	statusBar := lipgloss.NewStyle().
		Width(m.Size.Width).
		Foreground(types.DefaultTheme.Dim).
		Border(lipgloss.NormalBorder(), true, false, false, false). // Add horizontal divider on top
		BorderForeground(types.DefaultTheme.Dim).
		Render(statusText)

	var elements []string
	if len(m.Tabs.Tabs) > 0 {
		elements = append(elements, m.Tabs.View())
	}
	elements = append(elements, m.Editor.View(), statusBar)

	if m.ApiTesterExpanded {
		elements = append(elements, m.ApiTester.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, elements...)
}

func (m *WorkspaceModel) SetFocus(focus bool) {
	m.IsFocused = focus
	m.Tabs.SetFocus(focus)
	m.Editor.SetFocus(focus)
}
