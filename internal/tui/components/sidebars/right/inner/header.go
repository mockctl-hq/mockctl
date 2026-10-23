package inner

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type InnerHeaderModel struct {
	tuiCtx    types.TUIContext
	Title     string
	IsFocused bool
	Size      tea.WindowSizeMsg
}

func NewInnerHeaderModel(ctx types.TUIContext) *InnerHeaderModel {
	return &InnerHeaderModel{
		tuiCtx: ctx,
		Title:  "Chaos", // Default to Chaos
	}
}

func (m *InnerHeaderModel) SetTitle(title string) {
	m.Title = title
}

func (m *InnerHeaderModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *InnerHeaderModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.Size = msg
	}
	return m, nil
}

func (m *InnerHeaderModel) View() string {
	style := lipgloss.NewStyle().
		Width(m.Size.Width).
		Padding(0, 1).
		Bold(true).
		Foreground(types.DefaultTheme.Text).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(types.DefaultTheme.Dim)
		
	return style.Render(m.Title)
}

func (m *InnerHeaderModel) SetFocus(focus bool) {
	m.IsFocused = focus
}
