package system_prompts

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// ScreenWarningModel blocks UI if terminal is too small.
type ScreenWarningModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg
}

func NewScreenWarningModel(ctx types.TUIContext) *ScreenWarningModel {
	return &ScreenWarningModel{
		tuiCtx: ctx,
	}
}

func (m *ScreenWarningModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *ScreenWarningModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	}
	return m, nil
}

func (m *ScreenWarningModel) View() string {
	// Death Spiral Prevention
	if m.Size.Width < 10 || m.Size.Height < 10 {
		return "!"
	}

	dims := fmt.Sprintf("%dx%d", m.Size.Width, m.Size.Height)
	warningBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(types.DefaultTheme.Error).
		Padding(1, 2).
		Foreground(types.DefaultTheme.Error).
		Render("Terminal is too small.\nPlease increase window size, zoom out,\nor rotate your device.\n\nMinimum required: 120x24\nCurrent: " +
			lipgloss.NewStyle().Bold(true).Render(dims))

	return lipgloss.Place(m.Size.Width, m.Size.Height, lipgloss.Center, lipgloss.Center, warningBox)
}

func (m *ScreenWarningModel) SetFocus(focus bool) {}
