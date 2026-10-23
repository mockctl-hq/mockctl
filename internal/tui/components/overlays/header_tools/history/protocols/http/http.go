package http

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/runtime"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type HTTPEventViewer struct {
	tuiCtx types.TUIContext
	Event  *runtime.RequestEvent
	Size   tea.WindowSizeMsg
}

func NewHTTPEventViewer(ctx types.TUIContext) *HTTPEventViewer {
	return &HTTPEventViewer{
		tuiCtx: ctx,
	}
}

func (m *HTTPEventViewer) SetEvent(event *runtime.RequestEvent) {
	m.Event = event
}

func (m *HTTPEventViewer) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *HTTPEventViewer) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	}
	return m, nil
}

func (m *HTTPEventViewer) View() string {
	if m.Event == nil {
		return lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render("Select an event to view details")
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(types.DefaultTheme.Primary)
	labelStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)
	valueStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text)

	// Summary
	summary := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Request Summary"),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("Method:"), valueStyle.Render(m.Event.HTTPMethod)),
		fmt.Sprintf("%s %s", labelStyle.Render("Path:"), valueStyle.Render(m.Event.Path)),
		fmt.Sprintf("%s %s", labelStyle.Render("Status:"), valueStyle.Render(fmt.Sprintf("%d", m.Event.StatusCode))),
		fmt.Sprintf("%s %s", labelStyle.Render("Latency:"), valueStyle.Render(fmt.Sprintf("%dms", m.Event.LatencyMs))),
		fmt.Sprintf("%s %s", labelStyle.Render("Matched Route:"), valueStyle.Render(m.Event.MatchedRouteID)),
	)

	// Headers
	var reqHeaders []string
	for k, v := range m.Event.RequestHeaders {
		reqHeaders = append(reqHeaders, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
	}

	reqHeadersStr := labelStyle.Render("None")
	if len(reqHeaders) > 0 {
		reqHeadersStr = valueStyle.Render(strings.Join(reqHeaders, "\n"))
	}

	headersBlock := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Request Headers"),
		"",
		reqHeadersStr,
	)

	// Chaos
	chaosStr := labelStyle.Render("None")
	if m.Event.ChaosLatencyMs > 0 || m.Event.ChaosErrorInjected != "" {
		chaosStr = lipgloss.NewStyle().Foreground(types.DefaultTheme.Warning).Render("Yes (Delay/Error applied)")
	}

	chaosBlock := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Chaos Engineering"),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("Applied:"), chaosStr),
	)

	// Layout
	leftCol := lipgloss.JoinVertical(lipgloss.Left, summary, "", chaosBlock)
	rightCol := headersBlock

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftCol, "    ", rightCol)

	return lipgloss.NewStyle().Padding(1, 2).Render(content)
}

func (m *HTTPEventViewer) SetFocus(focus bool) {}
