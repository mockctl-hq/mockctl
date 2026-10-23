package logs

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *GlobalLogsModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		innerWidth := msg.Width - 6
		if innerWidth < 10 {
			innerWidth = 10
		}

		innerHeight := msg.Height - 2
		if innerHeight < 5 {
			innerHeight = 5
		}

		m.Search.Width = innerWidth
		vpHeight := innerHeight - 9
		if vpHeight < 1 {
			vpHeight = 1
		}

		m.Viewport.Width = innerWidth
		m.Viewport.Height = vpHeight

		m.mu.Lock()
		m.refreshViewport()
		m.mu.Unlock()

	case tea.KeyMsg:
		// Nested Esc routing
		if key.Matches(msg, types.GlobalKeys.Escape) {
			if m.Search.Focused() {
				m.Search.Blur()
				return m, nil
			}
			// Instruct root to close overlay
			return m, func() tea.Msg { return types.OverlayStateEnum(types.OverlayNone) }
		}

		if key.Matches(msg, types.GlobalKeys.AutoTail) && !m.Search.Focused() {
			m.AutoTail = !m.AutoTail
			if m.AutoTail {
				m.Viewport.GotoBottom()
			}
			return m, nil
		}

		if key.Matches(msg, types.GlobalKeys.Search) && !m.Search.Focused() {
			m.Search.Focus()
			return m, nil
		}

		if m.Search.Focused() {
			m.Search, cmd = m.Search.Update(msg)
			cmds = append(cmds, cmd)
			m.mu.Lock()
			m.refreshViewport()
			m.mu.Unlock()
		} else {
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)

			// Detect manual scrolling to turn off auto-tail
			if msg.String() == "up" || msg.String() == "k" {
				m.AutoTail = false
			}
		}

	case tea.MouseMsg:
		if !m.Search.Focused() {
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)
			if msg.Button == tea.MouseButtonWheelUp {
				m.AutoTail = false
			}
		}
	}

	return m, tea.Batch(cmds...)
}
