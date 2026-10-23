package http

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *HttpTesterModel) View() string {
	// Input Bar (Top Left)
	inputStr := m.Input.View()

	// Tabs (Body/Headers) (Middle Left)
	bodyTab := " Body "
	headersTab := " Headers "

	if m.FocusRing == FocusHeaders {
		headersTab = "[Headers]"
	} else if m.FocusRing == FocusBody {
		bodyTab = "[Body]"
	}

	tabStr := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(bodyTab + " " + headersTab)

	// Active Editor
	var editorStr string
	if m.FocusRing == FocusHeaders {
		editorStr = m.Headers.View()
	} else {
		editorStr = m.Body.View()
	}

	leftPanel := lipgloss.JoinVertical(lipgloss.Left, inputStr, tabStr, editorStr)
	rightPanel := m.Viewer.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}
