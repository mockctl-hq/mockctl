package header

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/components/header/actions"
	"github.com/mockctl-hq/mockctl/internal/tui/components/header/logo"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type HeaderModel struct {
	tuiCtx    types.TUIContext
	Size      tea.WindowSizeMsg
	IsFocused bool

	Logo    *logo.LogoModel
	Actions *actions.ActionIconsModel
	Toast   *shared.ToastModel
}

var _ types.Component = (*HeaderModel)(nil)

func NewHeaderModel(ctx types.TUIContext) *HeaderModel {
	return &HeaderModel{
		tuiCtx:  ctx,
		Logo:    logo.NewLogoModel(ctx),
		Actions: actions.NewActionIconsModel(ctx),
		Toast:   shared.NewToastModel(ctx),
	}
}

func (m *HeaderModel) Init(ctx types.TUIContext) tea.Cmd {
	return tea.Batch(
		m.Logo.Init(ctx),
		m.Actions.Init(ctx),
		m.Toast.Init(ctx),
	)
}

func (m *HeaderModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	case types.BlinkMsg:
		_, cmd := m.Logo.Update(msg)
		cmds = append(cmds, cmd)
	case types.ShowToastMsg, shared.HideToastMsg:
		_, cmd := m.Toast.Update(msg)
		cmds = append(cmds, cmd)
	case types.SettingsChangedMsg:
		if msg.Config != nil && !msg.Config.UIAnimations {
			m.Logo.ColonVisible = true
		}
	case tea.KeyMsg:
		if m.IsFocused {
			_, cmd := m.Actions.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *HeaderModel) View() string {
	logoStr := m.Logo.View()
	
	port := m.tuiCtx.ActivePort
	if port <= 0 {
		port = 8080
	}
	statusStr := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(fmt.Sprintf("[ Daemon: Port %d ]", port))
	if m.Toast != nil && m.Toast.Visible {
		statusStr = m.Toast.View()
	}
	
	actionsStr := m.Actions.View()

	// Calculate spacing
	usedWidth := lipgloss.Width(logoStr) + lipgloss.Width(statusStr) + lipgloss.Width(actionsStr)
	padding := m.Size.Width - usedWidth - 4 // -4 for some edge padding
	if padding < 0 {
		padding = 0
	}
	
	spaces := ""
	for i := 0; i < padding; i++ {
		spaces += " "
	}
	
	content := lipgloss.JoinHorizontal(lipgloss.Top, " ", logoStr, "  ", statusStr, spaces, actionsStr)
	
	// Vertically center by pushing it down one line (since height is 3)
	content = "\n" + content

	borderStyle := types.GlobalStyles.UnfocusedBorder
	if m.IsFocused {
		borderStyle = types.GlobalStyles.FocusedBorder
	}

	return borderStyle.
		Width(m.Size.Width).
		Height(3).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		Render(content)
}

func (m *HeaderModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if focus {
		m.Actions.SetFocus(true)
	} else {
		m.Actions.SetFocus(false)
	}
}
