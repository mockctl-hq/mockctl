package logo

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type LogoModel struct {
	tuiCtx       types.TUIContext
	ColonVisible bool
}

func NewLogoModel(ctx types.TUIContext) *LogoModel {
	return &LogoModel{
		tuiCtx:       ctx,
		ColonVisible: true,
	}
}

func (m *LogoModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *LogoModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg.(type) {
	case types.BlinkMsg:
		m.ColonVisible = !m.ColonVisible
	}
	return m, nil
}

func (m *LogoModel) View() string {
	mockStyle := lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Text).Render("𝗠𝗼𝗰𝗸")
	ctlStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Render("𝗰𝘁𝗹")

	colon := "∶" // Using a slightly wider math colon for aesthetics
	if !m.ColonVisible {
		colon = " "
	}
	colonStyle := lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Primary).Render(colon)

	return lipgloss.JoinHorizontal(lipgloss.Top, mockStyle, colonStyle, ctlStyle)
}

func (m *LogoModel) SetFocus(focus bool) {}
