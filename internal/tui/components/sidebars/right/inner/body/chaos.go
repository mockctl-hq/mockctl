package body

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type ChaosControlModel struct {
	tuiCtx         types.TUIContext
	ActiveProject  string
	ActiveEndpoint string
	IsFocused      bool

	Enabled         bool
	DelayInput      textinput.Model
	ErrorRateInput  textinput.Model
	StatusCodeInput textinput.Model

	ActiveField int // 0: Enable, 1: Delay, 2: ErrorRate, 3: StatusCode
}

func NewChaosControlModel(ctx types.TUIContext) *ChaosControlModel {
	delay := textinput.New()
	delay.Placeholder = "0 (e.g. 200)"
	delay.CharLimit = 10

	rate := textinput.New()
	rate.Placeholder = "0 (e.g. 25)"
	rate.CharLimit = 4

	status := textinput.New()
	status.Placeholder = "500 (e.g. 503)"
	status.CharLimit = 3

	return &ChaosControlModel{
		tuiCtx:          ctx,
		DelayInput:      delay,
		ErrorRateInput:  rate,
		StatusCodeInput: status,
	}
}

func (m *ChaosControlModel) Init(ctx types.TUIContext) tea.Cmd {
	m.tuiCtx = ctx
	return nil
}

func (m *ChaosControlModel) loadFromCore() {
	if m.tuiCtx.Core == nil || m.ActiveProject == "" {
		return
	}
	enabled, errRate, lat, status, err := m.tuiCtx.Core.GetChaos(context.Background(), m.ActiveProject)
	if err != nil {
		return
	}
	m.Enabled = enabled
	if lat > 0 {
		m.DelayInput.SetValue(fmt.Sprintf("%d", lat))
	} else {
		m.DelayInput.SetValue("")
	}
	if errRate > 0 {
		m.ErrorRateInput.SetValue(fmt.Sprintf("%d", errRate))
	} else {
		m.ErrorRateInput.SetValue("")
	}
	if status > 0 {
		m.StatusCodeInput.SetValue(fmt.Sprintf("%d", status))
	} else {
		m.StatusCodeInput.SetValue("500")
	}
}

func parseNumericInput(val string, defaultVal int) int {
	val = strings.TrimSpace(val)
	val = strings.TrimSuffix(val, "ms")
	val = strings.TrimSuffix(val, "%")
	val = strings.TrimSpace(val)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 0 {
		return defaultVal
	}
	return n
}

func (m *ChaosControlModel) applyChaosSilent() tea.Cmd {
	delay := parseNumericInput(m.DelayInput.Value(), 0)
	errRate := parseNumericInput(m.ErrorRateInput.Value(), 0)
	if errRate > 100 {
		errRate = 100
	}
	status := parseNumericInput(m.StatusCodeInput.Value(), 500)
	if status < 100 || status > 599 {
		status = 500
	}

	// Auto-enable if user entered non-zero chaos parameters and not yet enabled
	if !m.Enabled && (delay > 0 || errRate > 0) {
		m.Enabled = true
	}

	if m.tuiCtx.Core == nil || m.ActiveProject == "" {
		return nil
	}

	if !m.Enabled {
		_ = m.tuiCtx.Core.UpdateChaosWithStatus(context.Background(), m.ActiveProject, false, 0, 0, 500)
		return func() tea.Msg {
			return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
		}
	}

	_ = m.tuiCtx.Core.UpdateChaosWithStatus(context.Background(), m.ActiveProject, true, errRate, delay, status)
	return func() tea.Msg {
		return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
	}
}

func (m *ChaosControlModel) applyChaos() tea.Cmd {
	if m.tuiCtx.Core == nil || m.ActiveProject == "" {
		return nil
	}

	if !m.Enabled {
		_ = m.tuiCtx.Core.UpdateChaosWithStatus(context.Background(), m.ActiveProject, false, 0, 0, 500)
		return tea.Batch(
			func() tea.Msg {
				return types.ShowToastMsg{
					Message: fmt.Sprintf("Chaos disabled for %s", m.ActiveProject),
					IsError: false,
				}
			},
			func() tea.Msg {
				return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
			},
		)
	}

	delay := parseNumericInput(m.DelayInput.Value(), 0)
	errRate := parseNumericInput(m.ErrorRateInput.Value(), 0)
	if errRate > 100 {
		errRate = 100
	}
	status := parseNumericInput(m.StatusCodeInput.Value(), 500)
	if status < 100 || status > 599 {
		status = 500
	}

	err := m.tuiCtx.Core.UpdateChaosWithStatus(context.Background(), m.ActiveProject, true, errRate, delay, status)
	if err != nil {
		return func() tea.Msg {
			return types.ShowToastMsg{
				Message: "Failed to update chaos: " + err.Error(),
				IsError: true,
			}
		}
	}

	return tea.Batch(
		func() tea.Msg {
			return types.ShowToastMsg{
				Message: fmt.Sprintf("Chaos active for %s: %dms, %d%% errors (%d)", m.ActiveProject, delay, errRate, status),
				IsError: false,
			}
		},
		func() tea.Msg {
			return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
		},
	)
}

func (m *ChaosControlModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if !focus {
		m.DelayInput.Blur()
		m.ErrorRateInput.Blur()
		m.StatusCodeInput.Blur()
		m.applyChaosSilent()
	}
}
