package body

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *ProjectTreeModel) View() string {
	if len(m.VisibleNodes) == 0 {
		return lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Padding(1, 2).Render("No projects found.")
	}

	var lines []string

	endIdx := m.ScrollYOffset + m.Size.Height
	if endIdx > len(m.VisibleNodes) {
		endIdx = len(m.VisibleNodes)
	}

	for i := m.ScrollYOffset; i < endIdx; i++ {
		node := m.VisibleNodes[i]

		// Active Selection Style
		isCursor := i == m.ActiveIndex
		isSelected := node.ID == m.SelectedID && m.SelectedID != ""

		isCursorActive := m.IsFocused && isCursor

		prefix := "  "
		if node.IsProject {
			if node.IsExpanded {
				prefix = "  ▼ "
			} else {
				prefix = "  ▶ "
			}
		} else {
			prefix = "      " // Indent endpoints (4 spaces for tree hierarchy + 2 spaces for padding)
		}

		indicators := ""
		if !node.IsProject {
			if node.HasChaos {
				indicators += " 🟡"
			}
			if node.HasOverride {
				indicators += " 🟣"
			}
		}

		lineStr := prefix + node.ID + indicators

		runes := []rune(lineStr)
		if len(runes) > m.Size.Width {
			lineStr = string(runes[:m.Size.Width-3]) + "..."
		}

		if isCursorActive {
			lineStr = lipgloss.NewStyle().
				Background(types.DefaultTheme.Primary).
				Foreground(types.DefaultTheme.Bg).
				Bold(true).
				Width(m.Size.Width).
				Render(lineStr)
		} else if isSelected {
			// If it's selected but we are unfocused (or it's selected and we are focused but it's not the cursor)
			lineStr = lipgloss.NewStyle().
				Background(types.DefaultTheme.Dim).
				Foreground(types.DefaultTheme.Bg).
				Width(m.Size.Width).
				Render(lineStr)
		} else if node.IsProject {
			lineStr = lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Text).Render(lineStr)
		} else {
			lineStr = lipgloss.NewStyle().Foreground(types.DefaultTheme.Text).Render(lineStr)
		}

		lines = append(lines, lineStr)
	}

	return strings.Join(lines, "\n")
}
