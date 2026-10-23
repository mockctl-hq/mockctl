package overlays

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *ImportSpecModel) View() string {
	if !m.Active {
		return ""
	}

	// Fixed width to prevent expanding/jumping
	boxWidth := 70
	paddingCols := 2

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(types.DefaultTheme.Primary)

	// Helper to create a solid, fully colored line of exactly 70 characters
	padLine := func(s string) string {
		w := lipgloss.Width(s)
		rightPad := boxWidth - paddingCols - w
		if rightPad < 0 {
			rightPad = 0
		}

		bg := lipgloss.NewStyle().Background(types.DefaultTheme.Bg)
		return bg.Render(strings.Repeat(" ", paddingCols)) + s + bg.Render(strings.Repeat(" ", rightPad))
	}

	titleStr := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Background(types.DefaultTheme.Bg).Bold(true).Render("Import OpenAPI Spec")
	descStr := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Background(types.DefaultTheme.Bg).Render("Target Project: " + m.ProjectName)

	title := padLine(titleStr)
	desc := padLine(descStr)
	emptyLine := padLine(lipgloss.NewStyle().Background(types.DefaultTheme.Bg).Render(""))

	errStr := ""
	if m.ErrorMsg != "" {
		errStr = padLine(lipgloss.NewStyle().Foreground(types.DefaultTheme.Error).Background(types.DefaultTheme.Bg).Render("Error: " + m.ErrorMsg))
	}

	// Search Results
	var results []string
	if len(m.Matches) > 0 {
		for i, match := range m.Matches {
			displayMatch := match
			// truncate if too long
			maxLen := boxWidth - (paddingCols * 2) - 4
			if len(displayMatch) > maxLen {
				displayMatch = "..." + displayMatch[len(displayMatch)-maxLen:]
			}

			if i == m.CursorIndex {
				results = append(results, padLine(lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Background(types.DefaultTheme.Bg).Bold(true).Render("▶ "+displayMatch)))
			} else {
				results = append(results, padLine(lipgloss.NewStyle().Foreground(types.DefaultTheme.Text).Background(types.DefaultTheme.Bg).Render("  "+displayMatch)))
			}
		}
	} else if m.Input.Value() != "" && len(m.AllFiles) > 0 {
		results = append(results, padLine(lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Background(types.DefaultTheme.Bg).Render("  No matches found.")))
	}

	// Input Field
	inputView := padLine(m.Input.View())

	elements := []string{
		emptyLine, // Top Padding
		title,
		desc,
		emptyLine,
		inputView,
	}

	if len(results) > 0 {
		elements = append(elements, emptyLine)
		elements = append(elements, results...)
	}

	if errStr != "" {
		elements = append(elements, emptyLine)
		elements = append(elements, errStr)
	}

	elements = append(elements, emptyLine)
	elements = append(elements, padLine(lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Background(types.DefaultTheme.Bg).Render("(Press Enter to Import, Esc to Cancel)")))
	elements = append(elements, emptyLine) // Bottom Padding

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)

	// Wrap with border
	box := borderStyle.Render(content)

	return lipgloss.Place(m.Size.Width, m.Size.Height, lipgloss.Center, lipgloss.Center, box)
}
