package shared

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// OverlayDimensions computes safe inner width, height, and middle content height for an overlay modal.
type OverlayDimensions struct {
	InnerWidth   int
	InnerHeight  int
	MiddleHeight int
}

// ComputeOverlayDimensions calculates safe dimensions for full-screen overlay modals.
func ComputeOverlayDimensions(size tea.WindowSizeMsg, minWidth, minHeight int) OverlayDimensions {
	innerWidth := size.Width - 6
	if innerWidth < minWidth {
		innerWidth = minWidth
	}

	innerHeight := size.Height - 2
	if innerHeight < minHeight {
		innerHeight = minHeight
	}

	middleHeight := innerHeight - 4 // header (2) + footer (2)
	if middleHeight < 1 {
		middleHeight = 1
	}

	return OverlayDimensions{
		InnerWidth:   innerWidth,
		InnerHeight:  innerHeight,
		MiddleHeight: middleHeight,
	}
}

// RenderOverlayModal renders a full-screen overlay modal with consistent headers, dividers, borders, and footer.
func RenderOverlayModal(size tea.WindowSizeMsg, innerWidth, innerHeight int, title string, headerExtra string, body string, footerText string) string {
	titleStr := lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Primary).Render(title)
	if headerExtra != "" {
		titleStr = lipgloss.JoinHorizontal(lipgloss.Top, titleStr, "  ", headerExtra)
	}

	divider := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(strings.Repeat("─", innerWidth))
	header := lipgloss.JoinVertical(lipgloss.Left, titleStr, divider)

	footer := ""
	if footerText != "" {
		ft := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(footerText)
		footer = lipgloss.JoinVertical(lipgloss.Left, divider, ft)
	}

	elements := []string{header, body}
	if footer != "" {
		elements = append(elements, footer)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)
	content = lipgloss.NewStyle().Width(innerWidth).Height(innerHeight).Render(content)

	return lipgloss.Place(
		size.Width,
		size.Height,
		lipgloss.Center,
		lipgloss.Center,
		types.GlobalStyles.FocusedBorder.Padding(0, 2).Render(content),
	)
}
