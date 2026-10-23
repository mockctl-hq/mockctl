package body

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *ProjectTreeModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		return m, nil
		
	case tea.KeyMsg:
		if !m.IsFocused {
			return m, nil
		}
		
		if len(m.VisibleNodes) == 0 {
			return m, nil
		}

		if key.Matches(msg, types.GlobalKeys.Up) {
			if m.ActiveIndex > 0 {
				m.ActiveIndex--
			} else {
				m.ActiveIndex = len(m.VisibleNodes) - 1
			}
			m.syncScroll()
		} else if key.Matches(msg, types.GlobalKeys.Down) {
			if m.ActiveIndex < len(m.VisibleNodes)-1 {
				m.ActiveIndex++
			} else {
				m.ActiveIndex = 0
			}
			m.syncScroll()
		} else if msg.String() == "enter" || msg.String() == " " || msg.String() == "right" {
			// Toggle expansion or select
			node := m.VisibleNodes[m.ActiveIndex]
			if node.IsProject {
				node.IsExpanded = !node.IsExpanded
				m.SelectedID = node.ID // Update SelectedID BEFORE rebuild so cursor doesn't jump
				m.rebuildVisibleNodes()
				return m, func() tea.Msg { return types.ProjectSelectedMsg{ProjectName: node.ID, FileName: node.FileName} }
			} else {
				// It's an endpoint
				m.SelectedID = node.ID
				
				// Find parent project by iterating backwards
				projectName := ""
				for i := m.ActiveIndex - 1; i >= 0; i-- {
					if m.VisibleNodes[i].IsProject {
						projectName = m.VisibleNodes[i].ID
						break
					}
				}
				
				return m, func() tea.Msg { return types.EndpointSelectedMsg{ProjectName: projectName, EndpointPath: node.ID} }
			}
		} else if msg.String() == "shift+left" {
			// Collapse All
			for _, proj := range m.Nodes {
				proj.IsExpanded = false
			}
			m.ActiveIndex = 0
			m.rebuildVisibleNodes()
			m.syncScroll()
		}
	}

	return m, nil
}

func (m *ProjectTreeModel) syncScroll() {
	if m.Size.Height <= 0 {
		return
	}
	if m.ActiveIndex < m.ScrollYOffset {
		m.ScrollYOffset = m.ActiveIndex
	} else if m.ActiveIndex >= m.ScrollYOffset+m.Size.Height {
		m.ScrollYOffset = m.ActiveIndex - m.Size.Height + 1
	}
	if m.ScrollYOffset < 0 {
		m.ScrollYOffset = 0
	}
}
