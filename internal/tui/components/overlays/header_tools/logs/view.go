package logs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *GlobalLogsModel) View() string {
	dims := shared.ComputeOverlayDimensions(m.Size, 10, 5)

	tailStatus := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render("[Auto-Tail: OFF]")
	if m.AutoTail {
		tailStatus = lipgloss.NewStyle().Foreground(types.DefaultTheme.Success).Render("[Auto-Tail: ON]")
	}

	searchStr := m.Search.View()

	pathWidth := dims.InnerWidth - 60
	if pathWidth < 10 {
		pathWidth = 10
	}

	timeH := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true).Width(10).Render("TIME")
	methodH := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true).Width(7).Render("METHOD")
	pathH := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true).Width(pathWidth).Render("PATH")
	statusH := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true).Width(8).Render("STATUS")
	bytesH := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true).Width(10).Align(lipgloss.Right).Render("SIZE")
	latH := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true).Width(10).Align(lipgloss.Right).Render("LATENCY")

	tableHeader := lipgloss.NewStyle().Render(fmt.Sprintf("%s │ %s │ %s │ %s │ %s │ %s", timeH, methodH, pathH, statusH, bytesH, latH))
	divider := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(strings.Repeat("─", dims.InnerWidth))

	body := lipgloss.JoinVertical(lipgloss.Left,
		searchStr,
		"",
		divider,
		tableHeader,
		divider,
		m.Viewport.View(),
	)

	return shared.RenderOverlayModal(
		m.Size,
		dims.InnerWidth,
		dims.InnerHeight,
		"Global HTTP Request Logs",
		tailStatus,
		body,
		"Press / to search | f to toggle tail | Esc or q to close",
	)
}
