package shared

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// ConfirmResultMsg is returned when the user completes the modal.
type ConfirmResultMsg struct {
	Confirmed bool
	ID        string // To distinguish between different active confirmations
}

// ConfirmModel handles reusable Yes/No modals.
type ConfirmModel struct {
	tuiCtx   types.TUIContext
	Message  string
	ID       string
	Active   bool
	Selected bool // true = Yes, false = No
}

func NewConfirmModel(ctx types.TUIContext) *ConfirmModel {
	return &ConfirmModel{
		tuiCtx: ctx,
	}
}

func (m *ConfirmModel) Show(id string, message string) {
	m.ID = id
	m.Message = message
	m.Active = true
	m.Selected = false // Default to No (safer)
}

func (m *ConfirmModel) Hide() {
	m.Active = false
}

func (m *ConfirmModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *ConfirmModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Escape) {
			m.Hide()
			return m, func() tea.Msg { return ConfirmResultMsg{Confirmed: false, ID: m.ID} }
		}

		switch msg.String() {
		case "left", "h":
			m.Selected = true
		case "right", "l":
			m.Selected = false
		case "enter":
			m.Hide()
			return m, func() tea.Msg { return ConfirmResultMsg{Confirmed: m.Selected, ID: m.ID} }
		case "y":
			m.Hide()
			return m, func() tea.Msg { return ConfirmResultMsg{Confirmed: true, ID: m.ID} }
		case "n":
			m.Hide()
			return m, func() tea.Msg { return ConfirmResultMsg{Confirmed: false, ID: m.ID} }
		}
	}
	return m, nil
}

func (m *ConfirmModel) View() string {
	if !m.Active {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(types.DefaultTheme.Warning).
		Padding(1, 4).
		Align(lipgloss.Center)

	msgStyle := lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Text).Render(m.Message)

	btnStyle := lipgloss.NewStyle().Padding(0, 2)
	activeBtnStyle := btnStyle.Background(types.DefaultTheme.Primary).Foreground(types.DefaultTheme.Bg)

	yesBtn := btnStyle.Render("Yes")
	noBtn := btnStyle.Render("No")

	if m.Selected {
		yesBtn = activeBtnStyle.Render("Yes")
	} else {
		noBtn = activeBtnStyle.Render("No")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesBtn, "   ", noBtn)

	content := lipgloss.JoinVertical(lipgloss.Center, msgStyle, "", buttons)

	return boxStyle.Render(content)
}

func (m *ConfirmModel) SetFocus(focus bool) {}
