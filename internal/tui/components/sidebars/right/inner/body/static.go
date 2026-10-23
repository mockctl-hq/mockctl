package body

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type OverridesModel struct {
	tuiCtx         types.TUIContext
	ActiveProject  string
	ActiveEndpoint string
	IsFocused      bool
	Input          textarea.Model
	HasOverride    bool
	Size           tea.WindowSizeMsg
}

func NewOverridesModel(ctx types.TUIContext) *OverridesModel {
	ti := textarea.New()
	ti.Placeholder = "{\n  \"message\": \"override payload\"\n}"
	ti.CharLimit = 4000
	ti.SetHeight(8)
	ti.Prompt = lipgloss.NormalBorder().Left + " "
	ti.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary)
	ti.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("0"))
	ti.SetWidth(ti.Width())

	return &OverridesModel{
		tuiCtx: ctx,
		Input:  ti,
	}
}

func (m *OverridesModel) Init(ctx types.TUIContext) tea.Cmd {
	m.tuiCtx = ctx
	return nil
}

func (m *OverridesModel) getEndpointPath() string {
	parts := strings.SplitN(strings.TrimSpace(m.ActiveEndpoint), " ", 2)
	path := "/"
	if len(parts) == 2 {
		path = parts[1]
	} else if len(parts) == 1 && parts[0] != "" {
		path = parts[0]
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func (m *OverridesModel) loadFromCore() {
	if m.tuiCtx.Core == nil || m.ActiveProject == "" || m.ActiveEndpoint == "" {
		return
	}
	path := m.getEndpointPath()
	overrides, err := m.tuiCtx.Core.GetOverrides(context.Background(), m.ActiveProject)
	if err != nil {
		return
	}
	pathMap, ok := overrides[path]
	if !ok {
		pathMap, ok = overrides[m.ActiveEndpoint]
	}
	if ok {
		if val, exists := pathMap[200]; exists {
			b, err := json.MarshalIndent(val, "", "  ")
			if err == nil {
				m.Input.SetValue(string(b))
				m.HasOverride = true
				return
			}
		}
		for _, val := range pathMap {
			b, err := json.MarshalIndent(val, "", "  ")
			if err == nil {
				m.Input.SetValue(string(b))
				m.HasOverride = true
				return
			}
		}
	}
	m.Input.SetValue("")
	m.HasOverride = false
}

func (m *OverridesModel) saveOverride() tea.Cmd {
	path := m.getEndpointPath()
	raw := strings.TrimSpace(m.Input.Value())
	if raw == "" {
		if m.tuiCtx.Core != nil && m.ActiveProject != "" {
			err := m.tuiCtx.Core.SetEndpointOverride(context.Background(), m.ActiveProject, path, 200, nil)
			if err != nil {
				return func() tea.Msg {
					return types.ShowToastMsg{Message: "Failed to remove override: " + err.Error(), IsError: true}
				}
			}
		}
		m.HasOverride = false
		return tea.Batch(
			func() tea.Msg {
				return types.ShowToastMsg{Message: "Override removed (mock restored)", IsError: false}
			},
			func() tea.Msg {
				return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
			},
		)
	}

	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return func() tea.Msg {
			return types.ShowToastMsg{Message: "Invalid JSON: " + err.Error(), IsError: true}
		}
	}

	if m.tuiCtx.Core != nil && m.ActiveProject != "" {
		err := m.tuiCtx.Core.SetEndpointOverride(context.Background(), m.ActiveProject, path, 200, parsed)
		if err != nil {
			return func() tea.Msg {
				return types.ShowToastMsg{Message: "Failed to apply override: " + err.Error(), IsError: true}
			}
		}
	}
	m.HasOverride = true
	return tea.Batch(
		func() tea.Msg {
			return types.ShowToastMsg{Message: "Override saved & hot-swapped", IsError: false}
		},
		func() tea.Msg {
			return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
		},
	)
}

func (m *OverridesModel) clearOverride() tea.Cmd {
	path := m.getEndpointPath()
	m.Input.SetValue("")
	if m.tuiCtx.Core != nil && m.ActiveProject != "" {
		err := m.tuiCtx.Core.SetEndpointOverride(context.Background(), m.ActiveProject, path, 200, nil)
		if err != nil {
			return func() tea.Msg {
				return types.ShowToastMsg{Message: "Failed to clear override: " + err.Error(), IsError: true}
			}
		}
	}
	m.HasOverride = false
	return tea.Batch(
		func() tea.Msg {
			return types.ShowToastMsg{Message: "Override deleted", IsError: false}
		},
		func() tea.Msg {
			return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
		},
	)
}

func (m *OverridesModel) autoSyncOverride() tea.Cmd {
	if m.tuiCtx.Core == nil || m.ActiveProject == "" || m.ActiveEndpoint == "" {
		return nil
	}
	path := m.getEndpointPath()
	raw := strings.TrimSpace(m.Input.Value())

	if raw == "" {
		if m.HasOverride {
			_ = m.tuiCtx.Core.SetEndpointOverride(context.Background(), m.ActiveProject, path, 200, nil)
			m.HasOverride = false
			return func() tea.Msg {
				return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
			}
		}
		return nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil
	}

	err := m.tuiCtx.Core.SetEndpointOverride(context.Background(), m.ActiveProject, path, 200, parsed)
	if err == nil {
		m.HasOverride = true
		return func() tea.Msg {
			return types.ProjectStructureChangedMsg{ProjectName: m.ActiveProject}
		}
	}
	return nil
}

func (m *OverridesModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if focus {
		m.Input.Focus()
	} else {
		m.Input.Blur()
		m.autoSyncOverride()
	}
}
