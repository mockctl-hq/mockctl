package inspector

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *StateInspectorModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg

	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Escape) || key.Matches(msg, types.GlobalKeys.Quit) || msg.String() == "q" {
			return m, func() tea.Msg { return types.OverlayStateEnum(types.OverlayNone) }
		}

		if len(m.FlatVisible) > 0 {
			switch msg.String() {
			case "up", "k":
				if m.Cursor > 0 {
					m.Cursor--
				} else {
					m.Cursor = len(m.FlatVisible) - 1
				}
			case "down", "j":
				if m.Cursor < len(m.FlatVisible)-1 {
					m.Cursor++
				} else {
					m.Cursor = 0
				}
			case "enter", " ", "right", "l":
				node := m.FlatVisible[m.Cursor]
				if node.IsObject || node.IsArray {
					node.Expanded = !node.Expanded
					m.rebuildFlatList()
				}
			case "left", "h":
				node := m.FlatVisible[m.Cursor]
				if (node.IsObject || node.IsArray) && node.Expanded {
					node.Expanded = false
					m.rebuildFlatList()
				} else if node.Depth > 0 {
					// jump to parent (simple heuristic)
					for i := m.Cursor - 1; i >= 0; i-- {
						if m.FlatVisible[i].Depth < node.Depth {
							m.Cursor = i
							break
						}
					}
				}
			}
		}
	}
	return m, nil
}
