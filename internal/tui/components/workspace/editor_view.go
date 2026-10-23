package workspace

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *YamlEditorModel) View() string {
	if m.HighlightedContent == "" {
		return lipgloss.Place(m.Size.Width, m.Size.Height, lipgloss.Center, lipgloss.Center,
			types.GlobalStyles.Base.Foreground(types.DefaultTheme.Dim).Render("No Content"))
	}

	return m.Viewport.View()
}
