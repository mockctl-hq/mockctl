package inner

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type ScaffoldSubmitMsg struct {
	Value    string
	IsProject bool // true if creating project, false if creating endpoint
}

// ScaffoldInputModel handles inline Context-Aware Scaffolding (triggered by '+').
type ScaffoldInputModel struct {
	tuiCtx    types.TUIContext
	Input     textinput.Model
	Active    bool
	IsFocused bool
	IsProject bool // Determines if the placeholder says "Project Name" or "Endpoint Path"
}

func NewScaffoldInputModel(ctx types.TUIContext) *ScaffoldInputModel {
	ti := textinput.New()
	ti.Placeholder = "Name..."
	ti.CharLimit = 100
	ti.Prompt = "> "
	
	return &ScaffoldInputModel{
		tuiCtx: ctx,
		Input:  ti,
	}
}

func (m *ScaffoldInputModel) Show(isProject bool) {
	m.IsProject = isProject
	if isProject {
		m.Input.Placeholder = "New Project Name..."
	} else {
		m.Input.Placeholder = "New Endpoint Path..."
	}
	m.Input.SetValue("")
	m.Active = true
	m.Input.Focus()
	m.IsFocused = true
}

func (m *ScaffoldInputModel) Hide() {
	m.Active = false
	m.Input.Blur()
	m.IsFocused = false
}

func (m *ScaffoldInputModel) Init(ctx types.TUIContext) tea.Cmd {
	return textinput.Blink
}

func (m *ScaffoldInputModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	if !m.Active || !m.IsFocused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Escape) {
			m.Hide()
			return m, nil
		}
		if msg.String() == "enter" {
			val := m.Input.Value()
			if val != "" {
				m.Hide()
				return m, func() tea.Msg { return ScaffoldSubmitMsg{Value: val, IsProject: m.IsProject} }
			}
			return m, nil // Ignore empty submits
		}
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m *ScaffoldInputModel) View() string {
	if !m.Active {
		return ""
	}

	style := lipgloss.NewStyle().
		Width(24). // 29 (Inner Sidebar) - 5 (Icons Strip)
		Padding(0, 1)

	return style.Render(m.Input.View())
}

func (m *ScaffoldInputModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if focus && m.Active {
		m.Input.Focus()
	} else {
		m.Input.Blur()
	}
}
