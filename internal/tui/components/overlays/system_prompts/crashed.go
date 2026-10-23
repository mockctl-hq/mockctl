package system_prompts

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// DaemonCrashedModel blocks UI if background HTTP server fails to bind port.
type DaemonCrashedModel struct {
	tuiCtx       types.TUIContext
	Size         tea.WindowSizeMsg
	ErrorMessage string
}

func NewDaemonCrashedModel(ctx types.TUIContext) *DaemonCrashedModel {
	return &DaemonCrashedModel{
		tuiCtx: ctx,
	}
}

func (m *DaemonCrashedModel) ShowError(err string) {
	m.ErrorMessage = err
}

func (m *DaemonCrashedModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *DaemonCrashedModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Quit) {
			return m, tea.Quit
		}
		// R to restart handled in root model
	}
	return m, nil
}

func (m *DaemonCrashedModel) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(types.DefaultTheme.Error).
		Padding(2, 4).
		Align(lipgloss.Center)

	title := lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Error).Render("Daemon Crashed")
	body := "The background Mock Server failed to start."
	errLog := lipgloss.NewStyle().Foreground(types.DefaultTheme.Warning).Render(m.ErrorMessage)
	footer := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render("Press 'R' to Retry | 'q' to Quit")

	content := lipgloss.JoinVertical(lipgloss.Center, title, "", body, "", errLog, "", footer)

	return lipgloss.Place(m.Size.Width, m.Size.Height, lipgloss.Center, lipgloss.Center, boxStyle.Render(content))
}

func (m *DaemonCrashedModel) SetFocus(focus bool) {}
