package body

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *ChaosControlModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
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

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.GlobalKeys.Up) || msg.String() == "k" || msg.String() == "shift+tab":
			m.ActiveField = (m.ActiveField - 1 + 4) % 4
			cmds = append(cmds, m.applyChaosSilent())
			return m, tea.Batch(cmds...)
		case key.Matches(msg, types.GlobalKeys.Down) || msg.String() == "j" || msg.String() == "tab":
			m.ActiveField = (m.ActiveField + 1) % 4
			cmds = append(cmds, m.applyChaosSilent())
			return m, tea.Batch(cmds...)
		case m.ActiveField == 0 && (msg.String() == " " || msg.String() == "enter"):
			m.Enabled = !m.Enabled
			if m.Enabled && m.ErrorRateInput.Value() == "" && m.DelayInput.Value() == "" {
				m.ErrorRateInput.SetValue("25")
				m.StatusCodeInput.SetValue("500")
			}
			return m, m.applyChaos()
		case m.ActiveField > 0 && (msg.String() == "enter" || key.Matches(msg, types.GlobalKeys.Save)):
			if !m.Enabled {
				m.Enabled = true
			}
			return m, m.applyChaos()
		}
	}

	// Update active input and auto-sync changes in real-time on every keystroke
	if m.ActiveField == 1 {
		m.DelayInput.Focus()
		oldVal := m.DelayInput.Value()
		var inputCmd tea.Cmd
		m.DelayInput, inputCmd = m.DelayInput.Update(msg)
		cmds = append(cmds, inputCmd)
		if m.DelayInput.Value() != oldVal {
			cmds = append(cmds, m.applyChaosSilent())
		}
	} else {
		m.DelayInput.Blur()
	}

	if m.ActiveField == 2 {
		m.ErrorRateInput.Focus()
		oldVal := m.ErrorRateInput.Value()
		var inputCmd tea.Cmd
		m.ErrorRateInput, inputCmd = m.ErrorRateInput.Update(msg)
		cmds = append(cmds, inputCmd)
		if m.ErrorRateInput.Value() != oldVal {
			cmds = append(cmds, m.applyChaosSilent())
		}
	} else {
		m.ErrorRateInput.Blur()
	}

	if m.ActiveField == 3 {
		m.StatusCodeInput.Focus()
		oldVal := m.StatusCodeInput.Value()
		var inputCmd tea.Cmd
		m.StatusCodeInput, inputCmd = m.StatusCodeInput.Update(msg)
		cmds = append(cmds, inputCmd)
		if m.StatusCodeInput.Value() != oldVal {
			cmds = append(cmds, m.applyChaosSilent())
		}
	} else {
		m.StatusCodeInput.Blur()
	}

	return m, tea.Batch(cmds...)
}
