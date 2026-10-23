package body

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *OverridesModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		panelWidth := msg.Width - 6
		if panelWidth < 10 {
			panelWidth = 10
		}
		m.Input.SetWidth(panelWidth)
		return m, nil
	case types.EndpointSelectedMsg:
		m.ActiveProject = msg.ProjectName
		m.ActiveEndpoint = msg.EndpointPath
		m.loadFromCore()
		return m, nil
	case types.ProjectSelectedMsg:
		if msg.ProjectName != "" {
			m.ActiveProject = msg.ProjectName
			m.loadFromCore()
		}
		return m, nil
	}

	if !m.IsFocused || m.ActiveEndpoint == "" {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(keyMsg, types.GlobalKeys.Save) || keyMsg.String() == "ctrl+s" || (keyMsg.Type == tea.KeyEnter && keyMsg.Alt):
			return m, m.saveOverride()

		case keyMsg.Type == tea.KeyEnter && !keyMsg.Alt:
			raw := strings.TrimSpace(m.Input.Value())
			if raw != "" {
				var parsed any
				if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
					// Valid JSON submitted with Enter! Save immediately!
					return m, m.saveOverride()
				}
				// If closing brace or bracket entered, inform user of syntax error
				if strings.HasSuffix(raw, "}") || strings.HasSuffix(raw, "]") {
					err := json.Unmarshal([]byte(raw), &parsed)
					return m, func() tea.Msg {
						return types.ShowToastMsg{Message: "Invalid JSON: " + err.Error(), IsError: true}
					}
				}
			}

		case keyMsg.String() == "ctrl+d":
			return m, m.clearOverride()
		}
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)

	// Live validation: if valid JSON, auto-sync to core silently!
	raw := strings.TrimSpace(m.Input.Value())
	if raw != "" {
		var parsed any
		if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
			if !m.HasOverride && m.tuiCtx.Core != nil && m.ActiveProject != "" {
				path := m.getEndpointPath()
				_ = m.tuiCtx.Core.SetEndpointOverride(context.Background(), m.ActiveProject, path, 200, parsed)
				m.HasOverride = true
				return m, tea.Batch(cmd, func() tea.Msg {
					return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
				})
			}
		}
	}

	return m, cmd
}
