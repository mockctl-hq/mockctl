package history

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type HistoryModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg
	limit  int // max events in ring buffer, from SettingsConfig.HistoryLimit

	Viewport viewport.Model
	Search   textinput.Model
	Events   []types.TelemetryMsg
}

var _ types.Component = (*HistoryModel)(nil)

func NewHistoryModel(ctx types.TUIContext) *HistoryModel {
	vp := viewport.New(0, 0)
	ti := textinput.New()
	ti.Placeholder = "Search history..."

	limit := 1000
	if ctx.Settings != nil {
		limit = ctx.Settings.HistoryLimit
	}

	return &HistoryModel{
		tuiCtx:   ctx,
		Viewport: vp,
		Search:   ti,
		Events:   []types.TelemetryMsg{},
		limit:    limit,
	}
}

type HistoryLoadedMsg struct {
	Records []map[string]any
}

func (m *HistoryModel) Init(ctx types.TUIContext) tea.Cmd {
	m.tuiCtx = ctx
	if ctx.Core == nil {
		return nil
	}
	coreApp := ctx.Core
	reqCtx := ctx.Ctx
	if reqCtx == nil {
		reqCtx = context.Background()
	}
	limit := m.limit
	return func() tea.Msg {
		history, err := coreApp.GetTelemetryHistory(reqCtx, limit)
		if err != nil {
			return nil
		}
		return HistoryLoadedMsg{Records: history}
	}
}

func parseHistoryRecord(record map[string]any) types.TelemetryMsg {
	eventType, _ := record["EventType"].(string)
	tsStr, _ := record["Timestamp"].(string)
	ts, _ := time.Parse(time.RFC3339Nano, tsStr)
	if ts.IsZero() {
		ts, _ = time.Parse(time.RFC3339, tsStr)
	}

	switch eventType {
	case "REQUEST":
		method, _ := record["HTTPMethod"].(string)
		path, _ := record["Path"].(string)
		method = shared.SanitizeANSI(method)
		path = shared.SanitizeANSI(path)
		var status int
		if s, ok := record["StatusCode"].(float64); ok {
			status = int(s)
		} else if s, ok := record["StatusCode"].(int); ok {
			status = s
		}
		return types.TelemetryMsg{
			Payload: &runtime.RequestEvent{
				Timestamp:  ts,
				HTTPMethod: method,
				Path:       path,
				StatusCode: status,
			},
		}
	case "METRIC":
		var goroutines int
		if g, ok := record["NumGoroutine"].(float64); ok {
			goroutines = int(g)
		} else if g, ok := record["NumGoroutine"].(int); ok {
			goroutines = g
		}
		return types.TelemetryMsg{
			Payload: &runtime.MetricEvent{
				Timestamp:    ts,
				NumGoroutine: goroutines,
			},
		}
	case "DROP_WARNING":
		var dropped int64
		if d, ok := record["DroppedCount"].(float64); ok {
			dropped = int64(d)
		} else if d, ok := record["DroppedCount"].(int64); ok {
			dropped = d
		} else if d, ok := record["DroppedCount"].(int); ok {
			dropped = int64(d)
		}
		return types.TelemetryMsg{
			Payload: &runtime.DropWarningEvent{
				Timestamp:    ts,
				DroppedCount: dropped,
			},
		}
	}
	return types.TelemetryMsg{}
}

func (m *HistoryModel) loadHistoryRecords(records []map[string]any) {
	for _, rec := range records {
		msg := parseHistoryRecord(rec)
		if msg.Payload != nil {
			m.Events = append(m.Events, msg)
		}
	}
	if m.limit > 0 && len(m.Events) > m.limit {
		m.Events = m.Events[len(m.Events)-m.limit:]
	}
	m.refreshViewport()
}

func (m *HistoryModel) AddEvent(msg types.TelemetryMsg) {
	m.Events = append(m.Events, msg)
	if m.limit > 0 && len(m.Events) > m.limit {
		m.Events = m.Events[len(m.Events)-m.limit:]
	}
	m.refreshViewport()
}

func (m *HistoryModel) refreshViewport() {
	var lines []string
	for _, e := range m.Events {
		var line string
		switch ev := e.Payload.(type) {
		case *runtime.RequestEvent:
			line = fmt.Sprintf("[%s] %s - %s %s (%d)", ev.Timestamp.Format("15:04:05"), ev.EventType(), ev.HTTPMethod, ev.Path, ev.StatusCode)
		case *runtime.MetricEvent:
			line = fmt.Sprintf("[%s] %s - Goroutines: %d", ev.Timestamp.Format("15:04:05"), ev.EventType(), ev.NumGoroutine)
		case *runtime.DropWarningEvent:
			line = fmt.Sprintf("[%s] %s - Dropped: %d", ev.Timestamp.Format("15:04:05"), ev.EventType(), ev.DroppedCount)
		default:
			if ev != nil {
				line = fmt.Sprintf("[Unknown] %s", ev.EventType())
			}
		}
		if line != "" {
			lines = append(lines, line)
		}
	}

	filtered := shared.FilterSubstring(m.Search.Value(), lines)
	if len(filtered) == 0 {
		emptyMsg := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render("No telemetry history found.")
		content := lipgloss.Place(m.Viewport.Width, m.Viewport.Height, lipgloss.Center, lipgloss.Center, emptyMsg)
		m.Viewport.SetContent(content)
	} else {
		m.Viewport.SetContent(strings.Join(filtered, "\n"))
		m.Viewport.GotoBottom()
	}
}

func (m *HistoryModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
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
		// Header = title (1) + divider (1) = 2
		// Search = search input (1) + empty line below (1) = 2
		// Footer = divider (1) + footer text (1) = 2
		// Total non-viewport lines = 6
		vpHeight := innerHeight - 6
		if vpHeight < 1 {
			vpHeight = 1
		}

		m.Viewport.Width = innerWidth
		m.Viewport.Height = vpHeight

		m.refreshViewport()

	case HistoryLoadedMsg:
		m.loadHistoryRecords(msg.Records)
		return m, nil

	case types.TelemetryMsg:
		m.AddEvent(msg)
		return m, nil

	case types.SettingsChangedMsg:
		if msg.Config != nil {
			m.limit = msg.Config.HistoryLimit
			// Trim existing events if new limit is smaller
			if m.limit > 0 && len(m.Events) > m.limit {
				m.Events = m.Events[len(m.Events)-m.limit:]
				m.refreshViewport()
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Nested Esc routing
		if key.Matches(msg, types.GlobalKeys.Escape) {
			if m.Search.Focused() {
				m.Search.Blur()
				return m, nil
			}
			return m, func() tea.Msg { return types.OverlayStateEnum(types.OverlayNone) }
		}

		if key.Matches(msg, types.GlobalKeys.Quit) || msg.String() == "q" {
			if !m.Search.Focused() {
				return m, func() tea.Msg { return types.OverlayStateEnum(types.OverlayNone) }
			}
		}

		if key.Matches(msg, types.GlobalKeys.Search) && !m.Search.Focused() {
			m.Search.Focus()
			return m, nil
		}

		if m.Search.Focused() {
			m.Search, cmd = m.Search.Update(msg)
			cmds = append(cmds, cmd)
			m.refreshViewport()
		} else {
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *HistoryModel) View() string {
	dims := shared.ComputeOverlayDimensions(m.Size, 10, 5)
	searchStr := m.Search.View()
	body := lipgloss.JoinVertical(lipgloss.Left,
		searchStr,
		"",
		m.Viewport.View(),
	)

	return shared.RenderOverlayModal(
		m.Size,
		dims.InnerWidth,
		dims.InnerHeight,
		"Telemetry History",
		"",
		body,
		"Press / to search | Esc to close",
	)
}

func (m *HistoryModel) SetFocus(focus bool) {}
