package icons

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type LeftIconsModel struct {
	tuiCtx           types.TUIContext
	Size             tea.WindowSizeMsg
	ActiveIndex      int
	IsFocused        bool
	WorkspaceFocused bool
}

func NewLeftIconsModel(ctx types.TUIContext) *LeftIconsModel {
	return &LeftIconsModel{
		tuiCtx:      ctx,
		ActiveIndex: 0,
	}
}

func (m *LeftIconsModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *LeftIconsModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
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
			}
		} else if key.Matches(msg, types.GlobalKeys.Down) {
			// No other icons to go down to for now
		}
	}
	return m, nil
}

func (m *LeftIconsModel) View() string {
	style := lipgloss.NewStyle().
		Width(4).
		Height(m.Size.Height).
		Border(lipgloss.NormalBorder(), false, true, false, false).
		BorderForeground(types.DefaultTheme.Dim)

	if m.IsFocused || m.WorkspaceFocused {
		style = style.BorderForeground(types.DefaultTheme.Primary)
	}

	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	explorerIcon := " 📁 "
	if m.ActiveIndex == 0 {
		explorerIcon = activeStyle.Render(explorerIcon)
	} else {
		explorerIcon = inactiveStyle.Render(explorerIcon)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, explorerIcon)
	return style.Render(content)
}

func (m *LeftIconsModel) SetFocus(focus bool) {
	m.IsFocused = focus
}
