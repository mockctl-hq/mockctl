package body

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *OverridesModel) View() string {
	if m.ActiveEndpoint == "" {
		return lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Padding(1, 2).Render("Select an endpoint to create static overrides.")
	}

	style := lipgloss.NewStyle().Padding(1, 2)
	dimStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	raw := strings.TrimSpace(m.Input.Value())
	var badge string
	if raw == "" {
		if m.HasOverride {
			badge = lipgloss.NewStyle().Foreground(types.DefaultTheme.Secondary).Bold(true).Render("🟣 Override Active")
		} else {
			badge = dimStyle.Render("● Default Mock")
		}
	} else {
		var parsed any
		if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
			errStr := err.Error()
			if len(errStr) > 28 {
				errStr = errStr[:28] + "..."
			}
			badge = lipgloss.NewStyle().Foreground(types.DefaultTheme.Error).Bold(true).Render("⚠️ Invalid JSON (" + errStr + ")")
		} else {
			badge = lipgloss.NewStyle().Foreground(types.DefaultTheme.Secondary).Bold(true).Render("🟣 Override Active (Valid JSON)")
		}
	}

	labelStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text)
	if m.IsFocused {
		labelStyle = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	}
	label := labelStyle.Render("Hardcoded JSON Response:")

	m.Input.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary)
	m.Input.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("0"))

	content := lipgloss.JoinVertical(lipgloss.Left,
		badge,
		"",
		label,
		"",
		m.Input.View(),
	)

	return style.Render(content)
}
