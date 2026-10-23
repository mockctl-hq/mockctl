package http

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type HttpBodyModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	Input     textarea.Model
	IsFocused bool
	
	// Local History
	History []string
	HistoryIdx int
}

var _ types.Component = (*HttpBodyModel)(nil)

func NewHttpBodyModel(ctx types.TUIContext) *HttpBodyModel {
	ta := textarea.New()
	ta.Placeholder = "{\n  \"key\": \"value\"\n}"
	ta.ShowLineNumbers = false
	ta.Prompt = "" // Remove default thick vertical line prompt
	ta.CharLimit = 100 * 1024 // 100KB max for manual entry

	return &HttpBodyModel{
		tuiCtx:  ctx,
		Input:   ta,
		History: []string{},
	}
}

func (m *HttpBodyModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *HttpBodyModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		m.Input.SetWidth(msg.Width - 2) // Subtract 2 for lipgloss border and padding
		m.Input.SetHeight(msg.Height)
	case tea.KeyMsg:
		if m.IsFocused {
			// Local History Up/Down
			if msg.String() == "up" && m.Input.Line() == 0 {
				if m.HistoryIdx > 0 {
					m.HistoryIdx--
					m.Input.SetValue(m.History[m.HistoryIdx])
				}
				return m, nil
			}
			if msg.String() == "down" && m.Input.Line() == m.Input.LineCount()-1 {
				if m.HistoryIdx < len(m.History)-1 {
					m.HistoryIdx++
					m.Input.SetValue(m.History[m.HistoryIdx])
				}
				return m, nil
			}
			
			m.Input, cmd = m.Input.Update(msg)
		}
	}
	return m, cmd
}

func (m *HttpBodyModel) View() string {
	content := m.Input.View()
	style := lipgloss.NewStyle().
		Width(m.Size.Width - 2). // Account for left border and padding
		Height(m.Size.Height).
		MaxWidth(m.Size.Width - 2).
		MaxHeight(m.Size.Height).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		PaddingLeft(1)
		
	if m.IsFocused {
		style = style.BorderForeground(types.DefaultTheme.Primary)
	} else {
		// Make border invisible when not focused to preserve exact dimensions
		style = style.BorderForeground(types.DefaultTheme.Bg)
	}
	return style.Render(content)
}

func (m *HttpBodyModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if focus {
		m.Input.Focus()
	} else {
		m.Input.Blur()
	}
}

func (m *HttpBodyModel) PushHistory() {
	val := m.Input.Value()
	if val != "" {
		m.History = append(m.History, val)
		m.HistoryIdx = len(m.History)
	}
}
