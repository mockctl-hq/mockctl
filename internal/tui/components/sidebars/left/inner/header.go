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
		Title:  "Projects",
	}
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

	title := m.Title
	addIcon := "[+]"

	if m.IsFocused {
		addIcon = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Render(addIcon)
	}

	// Justify space between title and addIcon
	padding := m.Size.Width - lipgloss.Width(title) - lipgloss.Width(addIcon) - 2 // -2 for side padding
	if padding < 0 {
		padding = 0
	}

	spaces := ""
	for i := 0; i < padding; i++ {
		spaces += " "
	}

	return style.Render(title + spaces + addIcon)
}

func (m *InnerHeaderModel) SetFocus(focus bool) {
	m.IsFocused = focus
}
