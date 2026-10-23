package inspector

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *StateInspectorModel) View() string {
	dims := shared.ComputeOverlayDimensions(m.Size, 10, 5)
	footerMsg := "Enter: Expand/Collapse • ↑/↓: Navigate • Esc / q: Close"

	if m.Error != "" {
		errMsg := lipgloss.NewStyle().Foreground(types.DefaultTheme.Error).Render(m.Error)
		centeredErr := lipgloss.Place(dims.InnerWidth, dims.MiddleHeight, lipgloss.Center, lipgloss.Center, errMsg)
		return shared.RenderOverlayModal(m.Size, dims.InnerWidth, dims.InnerHeight, "State Inspector (MemoryStore)", "", centeredErr, footerMsg)
	}

	if len(m.FlatVisible) == 0 {
		emptyMsg := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render("State is empty.")
		centeredEmpty := lipgloss.Place(dims.InnerWidth, dims.MiddleHeight, lipgloss.Center, lipgloss.Center, emptyMsg)
		return shared.RenderOverlayModal(m.Size, dims.InnerWidth, dims.InnerHeight, "State Inspector (MemoryStore)", "", centeredEmpty, footerMsg)
	}

	// Rendering the tree
	startIdx := 0
	if m.Cursor >= dims.MiddleHeight {
		startIdx = m.Cursor - dims.MiddleHeight + 1
	}

	endIdx := startIdx + dims.MiddleHeight
	if endIdx > len(m.FlatVisible) {
		endIdx = len(m.FlatVisible)
	}

	var treeLines []string
	for i := startIdx; i < endIdx; i++ {
		node := m.FlatVisible[i]

		indent := strings.Repeat("  ", node.Depth)

		prefix := "  "
		if i == m.Cursor {
			prefix = lipgloss.NewStyle().Foreground(types.DefaultTheme.Secondary).Render("▶ ")
		}

		var icon string
		if node.IsObject || node.IsArray {
			if node.Expanded {
				icon = "▼ "
			} else {
				icon = "▶ "
			}
		} else {
			icon = "  "
		}

		keyStr := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text).Bold(true).Render(node.Key)
		valStr := lipgloss.NewStyle().Foreground(types.DefaultTheme.Success).Render(node.Value)

		line := fmt.Sprintf("%s%s%s%s", prefix, indent, icon, keyStr)
		if node.Value != "" {
			line += ": " + valStr
		}

		treeLines = append(treeLines, line)
	}

	// Pad remaining lines up to middleHeight so footer is pushed to the bottom
	for len(treeLines) < dims.MiddleHeight {
		treeLines = append(treeLines, "")
	}

	middleContent := strings.Join(treeLines, "\n")
	middleView := lipgloss.NewStyle().Width(dims.InnerWidth).Height(dims.MiddleHeight).Render(middleContent)

	return shared.RenderOverlayModal(m.Size, dims.InnerWidth, dims.InnerHeight, "State Inspector (MemoryStore)", "", middleView, footerMsg)
}
