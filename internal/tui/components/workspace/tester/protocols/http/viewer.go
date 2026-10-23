package http

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type HttpViewerModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	Viewport  viewport.Model
	Response  *TestResponse
	IsFocused bool
	IsLoading bool
	Spinner   spinner.Model
}

var _ types.Component = (*HttpViewerModel)(nil)

func NewHttpViewerModel(ctx types.TUIContext) *HttpViewerModel {
	vp := viewport.New(0, 0)
	vp.SetContent(" Send a request to see the response here.")

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	sp.Style = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary)

	return &HttpViewerModel{
		tuiCtx:   ctx,
		Viewport: vp,
		Spinner:  sp,
	}
}

func (m *HttpViewerModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *HttpViewerModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.IsLoading {
			m.Spinner, cmd = m.Spinner.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.Size = msg
		m.Viewport.Width = msg.Width - 1   // Account for 1 left border
		m.Viewport.Height = msg.Height - 2 // Minus status bar (1) and empty line (1)
	case tea.KeyMsg:
		if m.IsFocused {
			m.Viewport, cmd = m.Viewport.Update(msg)
		}
	case tea.MouseMsg:
		m.Viewport, cmd = m.Viewport.Update(msg)
	}
	return m, cmd
}

func (m *HttpViewerModel) View() string {
	var statusBarStr string
	if m.IsLoading {
		statusBarStr = lipgloss.NewStyle().
			Foreground(types.DefaultTheme.Primary).
			Render(" Loading...")
	} else if m.Response == nil {
		statusBarStr = lipgloss.NewStyle().
			Foreground(types.DefaultTheme.Dim).
			Render(" Ready")
	} else if m.Response.Error != nil {
		statusBarStr = lipgloss.NewStyle().
			Foreground(types.DefaultTheme.Error).
			Bold(true).
			Render(fmt.Sprintf(" Error: %s | %s", m.Response.Error.Error(), m.Response.Latency.String()))
	} else {
		color := types.DefaultTheme.Success
		if m.Response.StatusCode >= 400 {
			color = types.DefaultTheme.Error
		} else if m.Response.StatusCode >= 300 {
			color = types.DefaultTheme.Warning
		}

		status := lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf(" %s ", m.Response.StatusText))
		latency := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(m.Response.Latency.String())

		statusBarStr = lipgloss.JoinHorizontal(lipgloss.Top, status, "| ", latency)
	}

	var vpStr string
	if m.IsLoading {
		spinnerStr := m.Spinner.View()
		loadingText := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text).Render("Loading...")
		vpStr = " " + spinnerStr + " " + loadingText
	} else if m.Response == nil {
		vpStr = lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(" Send a request to see the response here.")
	} else {
		vpStr = m.Viewport.View()
	}

	content := lipgloss.JoinVertical(lipgloss.Left, statusBarStr, "", vpStr)

	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(types.DefaultTheme.Dim).
		Width(m.Size.Width - 1). // Account for 1 left border
		Height(m.Size.Height)

	if m.IsFocused {
		style = style.BorderForeground(types.DefaultTheme.Primary)
	}

	return style.Render(content)
}

func (m *HttpViewerModel) SetFocus(focus bool) {
	m.IsFocused = focus
}

func (m *HttpViewerModel) SetLoading(loading bool) tea.Cmd {
	m.IsLoading = loading
	if loading {
		m.Response = nil
		return m.Spinner.Tick
	}
	return nil
}

func (m *HttpViewerModel) SetResponse(resp *TestResponse) {
	m.IsLoading = false
	m.Response = resp
	if resp == nil {
		m.Viewport.SetContent(" Send a request to see the response here.")
		m.Viewport.GotoTop()
		return
	}
	if resp.Error != nil {
		m.Viewport.SetContent(" " + resp.Error.Error())
		m.Viewport.GotoTop()
		return
	}

	if resp.IsBinary {
		m.Viewport.SetContent(resp.Body)
		m.Viewport.GotoTop()
		return
	}

	content := resp.Body

	// Pretty print JSON
	if len(content) > 0 && (strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[")) {
		var obj interface{}
		if err := json.Unmarshal([]byte(content), &obj); err == nil {
			pretty, err := json.MarshalIndent(obj, "", "  ")
			if err == nil {
				content = string(pretty)
			}
		}
	}

	// Try chroma highlighting
	if len(content) < 100*1024 { // PKS-008: 100KB limit
		var buf strings.Builder
		err := quick.Highlight(&buf, content, "json", "terminal256", "monokai")
		if err == nil {
			content = buf.String()
		}
	}

	m.Viewport.SetContent(content)
	m.Viewport.GotoTop()
}
