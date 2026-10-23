package body

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *ChaosControlModel) View() string {
	if m.ActiveEndpoint == "" {
		return lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Padding(1, 2).Render("Select an endpoint to configure chaos constraints.")
	}

	style := lipgloss.NewStyle().Padding(1, 2)
	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text)
	dimStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	badge := dimStyle.Render("● Inactive")
	if m.Enabled {
		badge = lipgloss.NewStyle().Foreground(types.DefaultTheme.Warning).Bold(true).Render("🟡 Chaos Running")
	}

	// Enable Toggle
	enableStr := "[ ] Enable Chaos"
	if m.Enabled {
		enableStr = "[x] Enable Chaos"
	}
	if m.ActiveField == 0 && m.IsFocused {
		enableStr = activeStyle.Render(enableStr)
	} else {
		enableStr = inactiveStyle.Render(enableStr)
	}

	// Delay Input
	delayLabel := "Delay (ms):"
	if m.ActiveField == 1 && m.IsFocused {
		delayLabel = activeStyle.Render(delayLabel)
	} else {
		delayLabel = dimStyle.Render(delayLabel)
	}

	// Error Rate Input
	rateLabel := "Error Rate (%):"
	if m.ActiveField == 2 && m.IsFocused {
		rateLabel = activeStyle.Render(rateLabel)
	} else {
		rateLabel = dimStyle.Render(rateLabel)
	}

	// Status Code
	statusLabel := "Force Status:"
	if m.ActiveField == 3 && m.IsFocused {
		statusLabel = activeStyle.Render(statusLabel)
	} else {
		statusLabel = dimStyle.Render(statusLabel)
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		badge,
		"",
		enableStr,
		"",
		delayLabel,
		"  "+m.DelayInput.View(),
		"",
		rateLabel,
		"  "+m.ErrorRateInput.View(),
		"",
		statusLabel,
		"  "+m.StatusCodeInput.View(),
	)

	return style.Render(content)
}
