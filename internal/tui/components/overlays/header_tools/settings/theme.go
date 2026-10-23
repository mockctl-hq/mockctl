package settings

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type ThemeTabModel struct {
	tuiCtx      types.TUIContext
	Animations  bool
	ActiveTheme int // 0: Default (Adaptive), 1: Ocean Dark, 2: Dracula, 3: Monokai
	ActiveField int // 0: Theme, 1: UI Animations
}

func NewThemeTabModel(ctx types.TUIContext) *ThemeTabModel {
	return &ThemeTabModel{
		tuiCtx:      ctx,
		ActiveTheme: 0,
		Animations:  true,
	}
}

func (m *ThemeTabModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *ThemeTabModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Up) || msg.String() == "k" {
			if m.ActiveField > 0 {
				m.ActiveField--
			} else {
				m.ActiveField = 1
			}
		} else if key.Matches(msg, types.GlobalKeys.Down) || msg.String() == "j" {
			if m.ActiveField < 1 {
				m.ActiveField++
			} else {
				m.ActiveField = 0
			}
		} else if msg.String() == " " || msg.String() == "enter" {
			if m.ActiveField == 0 {
				m.ActiveTheme = (m.ActiveTheme + 1) % 4
			} else if m.ActiveField == 1 {
				m.Animations = !m.Animations
			}
		}
	}
	return m, nil
}

func (m *ThemeTabModel) View() string {
	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text)
	dimStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	themeName := "Default (Adaptive)"
	if m.ActiveTheme >= 0 && m.ActiveTheme < len(types.ThemeList) {
		themeName = types.ThemeList[m.ActiveTheme].Name
	}
	themeStr := "Active Theme: [" + themeName + "]"
	
	if m.ActiveField == 0 {
		themeStr = activeStyle.Render(themeStr)
	} else {
		themeStr = inactiveStyle.Render(themeStr)
	}

	animStr := "[ ] Enable UI Animations"
	if m.Animations {
		animStr = "[x] Enable UI Animations"
	}
	
	if m.ActiveField == 1 {
		animStr = activeStyle.Render(animStr)
	} else {
		animStr = inactiveStyle.Render(animStr)
	}

	helpStr := dimStyle.Render("Press Space/Enter to cycle themes. Applied across the TUI.")

	return lipgloss.JoinVertical(lipgloss.Left, themeStr, "", animStr, "", helpStr)
}

func (m *ThemeTabModel) SetFocus(focus bool) {}
