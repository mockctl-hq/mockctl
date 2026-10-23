package settings

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type TesterTabModel struct {
	tuiCtx          types.TUIContext
	FollowRedirects bool
	ActiveField     int // 0: Redirects
}

func NewTesterTabModel(ctx types.TUIContext) *TesterTabModel {
	return &TesterTabModel{
		tuiCtx:          ctx,
		FollowRedirects: false,
	}
}

func (m *TesterTabModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *TesterTabModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == " " || msg.String() == "enter" {
			if m.ActiveField == 0 {
				m.FollowRedirects = !m.FollowRedirects
			}
		}
	}
	return m, nil
}

func (m *TesterTabModel) View() string {
	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text)
	dimStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	redirStr := "[ ] Follow HTTP Redirects"
	if m.FollowRedirects {
		redirStr = "[x] Follow HTTP Redirects"
	}
	
	if m.ActiveField == 0 {
		redirStr = activeStyle.Render(redirStr)
	} else {
		redirStr = inactiveStyle.Render(redirStr)
	}
	
	helpStr := dimStyle.Render("Press Space/Enter to toggle")

	return lipgloss.JoinVertical(lipgloss.Left, redirStr, "", helpStr)
}

func (m *TesterTabModel) SetFocus(focus bool) {}
