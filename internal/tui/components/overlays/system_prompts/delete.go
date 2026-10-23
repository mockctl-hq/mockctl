package system_prompts

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// DeleteProjectMsg is emitted when a valid deletion is confirmed.
type DeleteProjectMsg struct {
	ProjectName string
}

// DeleteProjectModel blocks UI for catastrophic deletion.
type DeleteProjectModel struct {
	tuiCtx      types.TUIContext
	Size        tea.WindowSizeMsg
	Input       textinput.Model
	ProjectName string
	Active      bool
	ErrorMsg    string
}

func NewDeleteProjectModel(ctx types.TUIContext) *DeleteProjectModel {
	ti := textinput.New()
	ti.Placeholder = "Type project name here"
	ti.Focus()
	ti.CharLimit = 100

	return &DeleteProjectModel{
		tuiCtx: ctx,
		Input:  ti,
	}
}

func (m *DeleteProjectModel) Show(projectName string) {
	m.ProjectName = projectName
	m.Active = true
	m.ErrorMsg = ""
	m.Input.SetValue("")
	m.Input.Focus()
}

func (m *DeleteProjectModel) Hide() {
	m.Active = false
	m.Input.Blur()
}

func (m *DeleteProjectModel) Init(ctx types.TUIContext) tea.Cmd {
	return textinput.Blink
}

func (m *DeleteProjectModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Escape) {
			m.Hide()
			return m, nil
		}
		if msg.String() == "enter" {
			if m.Input.Value() == m.ProjectName {
				m.Hide()
				return m, func() tea.Msg { return DeleteProjectMsg{ProjectName: m.ProjectName} }
			}
			m.ErrorMsg = "Name does not match. Try again."
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m *DeleteProjectModel) View() string {
	if !m.Active {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(types.DefaultTheme.Error).
		Padding(1, 4).
		Align(lipgloss.Center)

	title := lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Error).Render("Delete Project?")
	warning := "This action cannot be undone. All endpoints and settings will be permanently lost."
	prompt := "Please type '" + lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Warning).Render(m.ProjectName) + "' to confirm."
	
	errStr := ""
	if m.ErrorMsg != "" {
		errStr = lipgloss.NewStyle().Foreground(types.DefaultTheme.Error).Render(m.ErrorMsg)
	}

	content := lipgloss.JoinVertical(lipgloss.Center, title, "", warning, prompt, "", m.Input.View(), errStr)
	
	return lipgloss.Place(m.Size.Width, m.Size.Height, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
}

func (m *DeleteProjectModel) SetFocus(focus bool) {
	if focus {
		m.Input.Focus()
	} else {
		m.Input.Blur()
	}
}
