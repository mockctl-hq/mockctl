package actions

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type OverlayToggleMsg struct {
	OverlayType types.OverlayStateEnum
}

type ActionIconsModel struct {
	tuiCtx      types.TUIContext
	IsFocused   bool
	ActiveIndex int
}

func NewActionIconsModel(ctx types.TUIContext) *ActionIconsModel {
	return &ActionIconsModel{
		tuiCtx: ctx,
	}
}

func (m *ActionIconsModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *ActionIconsModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	if !m.IsFocused {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "left" || msg.String() == "h" {
			if m.ActiveIndex > 0 {
				m.ActiveIndex--
			} else {
				m.ActiveIndex = 4
			}
		} else if msg.String() == "right" || msg.String() == "l" {
			if m.ActiveIndex < 4 { // 5 icons total
				m.ActiveIndex++
			} else {
				m.ActiveIndex = 0
			}
		} else if msg.String() == "enter" || msg.String() == " " {
			// Trigger overlay based on active index
			var overlay types.OverlayStateEnum
			switch m.ActiveIndex {
			case 0:
				overlay = types.OverlayStateInspector
			case 1:
				overlay = types.OverlayHistory
			case 2:
				overlay = types.OverlayGlobalLog
			case 3:
				overlay = types.OverlayProfiler
			case 4:
				overlay = types.OverlaySettings
			}
			return m, func() tea.Msg { return overlay }
		}
	}
	return m, nil
}

func (m *ActionIconsModel) View() string {
	style := lipgloss.NewStyle().Padding(0, 1)

	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	iconsStr := []string{"[ 🔍 Inspector ]", "[ ⏳ History ]", "[ 📜 Logs ]", "[ 📊 Profiler ]", "[ ⚙️ Settings ]"}

	var rendered []string
	for i, icon := range iconsStr {
		if m.IsFocused && i == m.ActiveIndex {
			rendered = append(rendered, activeStyle.Render(icon))
		} else {
			rendered = append(rendered, inactiveStyle.Render(icon))
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, rendered[0], " ", rendered[1], " ", rendered[2], " ", rendered[3], " ", rendered[4])
	return style.Render(content)
}

func (m *ActionIconsModel) SetFocus(focus bool) {
	m.IsFocused = focus
}
