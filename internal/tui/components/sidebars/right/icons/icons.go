package icons

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type RightIconsModel struct {
	tuiCtx           types.TUIContext
	Size             tea.WindowSizeMsg
	ActiveIndex      int
	IsFocused        bool
	WorkspaceFocused bool
}

func NewRightIconsModel(ctx types.TUIContext) *RightIconsModel {
	return &RightIconsModel{
		tuiCtx:      ctx,
		ActiveIndex: 0,
	}
}

func (m *RightIconsModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *RightIconsModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	case tea.KeyMsg:
		if !m.IsFocused {
			return m, nil
		}
		if key.Matches(msg, types.GlobalKeys.Up) {
			if m.ActiveIndex > 0 {
				m.ActiveIndex--
			} else {
				m.ActiveIndex = 2
			}
		} else if key.Matches(msg, types.GlobalKeys.Down) {
			if m.ActiveIndex < 2 { // 3 icons total: 0, 1, 2
				m.ActiveIndex++
			} else {
				m.ActiveIndex = 0
			}
		}
	}
	return m, nil
}

func (m *RightIconsModel) View() string {
	style := lipgloss.NewStyle().
		Width(4).
		Height(m.Size.Height).
		Border(lipgloss.NormalBorder(), false, false, false, true). // Left border
		BorderForeground(types.DefaultTheme.Dim)

	if m.IsFocused || m.WorkspaceFocused {
		style = style.BorderForeground(types.DefaultTheme.Primary)
	}

	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	iconsStr := []string{" 🌀 ", " 💻 ", " 🟣 "}

	var rendered []string
	for i, icon := range iconsStr {
		if i == m.ActiveIndex {
			rendered = append(rendered, activeStyle.Render(icon))
		} else {
			rendered = append(rendered, inactiveStyle.Render(icon))
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rendered[0], "", rendered[1], "", rendered[2])
	return style.Render(content)
}

func (m *RightIconsModel) SetFocus(focus bool) {
	m.IsFocused = focus
}
