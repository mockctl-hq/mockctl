package http

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type HttpInputModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	Method      string
	PathInput   textinput.Model
	IsFocused   bool
	
	// Local History
	History []string
	HistoryIdx int
}

// Ensure compliance
var _ types.Component = (*HttpInputModel)(nil)

func NewHttpInputModel(ctx types.TUIContext) *HttpInputModel {
	ti := textinput.New()
	ti.Placeholder = "/users/1"
	ti.CharLimit = 256

	return &HttpInputModel{
		tuiCtx:    ctx,
		Method:    "GET",
		PathInput: ti,
		History:   []string{},
	}
}

func (m *HttpInputModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *HttpInputModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.IsFocused {
			if msg.String() == "ctrl+up" {
				m.Method = cycleMethod(m.Method, -1)
				return m, nil
			}
			if msg.String() == "ctrl+down" {
				m.Method = cycleMethod(m.Method, 1)
				return m, nil
			}

			// Local History Up/Down
			if msg.String() == "up" {
				if m.HistoryIdx > 0 {
					m.HistoryIdx--
					m.PathInput.SetValue(m.History[m.HistoryIdx])
				}
				return m, nil
			}
			if msg.String() == "down" {
				if m.HistoryIdx < len(m.History)-1 {
					m.HistoryIdx++
					m.PathInput.SetValue(m.History[m.HistoryIdx])
				}
				return m, nil
			}
			
			m.PathInput, cmd = m.PathInput.Update(msg)
		}
	}
	return m, cmd
}

var methods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}

func cycleMethod(current string, dir int) string {
	for i, m := range methods {
		if m == current {
			next := (i + dir + len(methods)) % len(methods)
			return methods[next]
		}
	}
	return "GET"
}

func (m *HttpInputModel) View() string {
	methodStyle := lipgloss.NewStyle().
		Background(types.DefaultTheme.Primary).
		Foreground(types.DefaultTheme.Bg).
		Padding(0, 1).
		Bold(true).
		Render(m.Method)

	inputStr := lipgloss.NewStyle().
		Width(m.Size.Width - lipgloss.Width(methodStyle) - 1).
		Render(m.PathInput.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, methodStyle, " ", inputStr)
}

func (m *HttpInputModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if focus {
		m.PathInput.Focus()
	} else {
		m.PathInput.Blur()
	}
}

func (m *HttpInputModel) PushHistory() {
	val := m.PathInput.Value()
	if val != "" {
		m.History = append(m.History, val)
		m.HistoryIdx = len(m.History)
	}
}
