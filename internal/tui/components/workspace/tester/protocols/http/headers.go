package http

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type HttpHeadersModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	Input     textarea.Model
	IsFocused bool
}

var _ types.Component = (*HttpHeadersModel)(nil)

func NewHttpHeadersModel(ctx types.TUIContext) *HttpHeadersModel {
	ta := textarea.New()
	ta.Placeholder = "Authorization: Bearer token\nX-Custom: value"
	ta.ShowLineNumbers = false
	ta.Prompt = "" // Remove default thick vertical line prompt
	ta.CharLimit = 10 * 1024 // 10KB max

	return &HttpHeadersModel{
		tuiCtx: ctx,
		Input:  ta,
	}
}

func (m *HttpHeadersModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *HttpHeadersModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		m.Input.SetWidth(msg.Width - 2) // Subtract 2 for lipgloss border and padding
		m.Input.SetHeight(msg.Height)
	case tea.KeyMsg:
		if m.IsFocused {
			m.Input, cmd = m.Input.Update(msg)
		}
	}
	return m, cmd
}

func (m *HttpHeadersModel) View() string {
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

func (m *HttpHeadersModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if focus {
		m.Input.Focus()
	} else {
		m.Input.Blur()
	}
}
