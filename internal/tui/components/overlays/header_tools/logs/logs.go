package logs

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type GlobalLogsModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	Viewport viewport.Model
	Search   textinput.Model

	Lines    []string
	mu       sync.RWMutex
	AutoTail bool

	lastFileContent string
}

var _ types.Component = (*GlobalLogsModel)(nil)

func NewGlobalLogsModel(ctx types.TUIContext) *GlobalLogsModel {
	vp := viewport.New(0, 0)
	ti := textinput.New()
	ti.Placeholder = "Search logs..."

	return &GlobalLogsModel{
		tuiCtx:   ctx,
		Viewport: vp,
		Search:   ti,
		AutoTail: true,
		Lines:    make([]string, 0, 1000), // 1000 line ring buffer
	}
}

func (m *GlobalLogsModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *GlobalLogsModel) AddLogLine(line string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Lines = append(m.Lines, shared.SanitizeANSI(line))
	if len(m.Lines) > 1000 {
		m.Lines = m.Lines[1:] // Ring buffer
	}

	m.refreshViewport()
}

func (m *GlobalLogsModel) RefreshFromDisk() {
	if m.tuiCtx.Core == nil {
		return
	}
	bytes, err := m.tuiCtx.Core.GetRequestsLogContent()
	if err == nil && len(bytes) > 0 {
		content := strings.TrimRight(string(bytes), "\n")

		m.mu.Lock()
		if content == m.lastFileContent {
			m.mu.Unlock()
			return
		}
		m.lastFileContent = content
		m.mu.Unlock()

		lines := strings.Split(content, "\n")
		// Keep last 1000 lines max
		if len(lines) > 1000 {
			lines = lines[len(lines)-1000:]
		}

		m.mu.Lock()
		m.Lines = lines
		m.refreshViewport()
		m.mu.Unlock()
	}
}

func parseLogLine(line string, width int) string {
	pathWidth := width - 60
	if pathWidth < 10 {
		pathWidth = 10
	}

	parts := strings.Split(line, " | ")
	if len(parts) == 4 {
		prefixParts := strings.SplitN(parts[0], " HTTP ", 2)
		if len(prefixParts) == 2 {
			timestamp := prefixParts[0]
			reqInfo := strings.SplitN(prefixParts[1], " ", 2)
			if len(reqInfo) == 2 {
				method := reqInfo[0]
				path := reqInfo[1]

				status := strings.TrimPrefix(parts[1], "status: ")
				bytesStr := strings.TrimPrefix(parts[2], "bytes: ")
				latency := parts[3]

				timeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Width(10)

				methodStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Secondary).Bold(true).Width(7)
				if method == "GET" {
					methodStyle = methodStyle.Foreground(types.DefaultTheme.Success)
				} else if method == "POST" {
					methodStyle = methodStyle.Foreground(types.DefaultTheme.Primary)
				} else if method == "DELETE" {
					methodStyle = methodStyle.Foreground(types.DefaultTheme.Error)
				} else if method == "PUT" || method == "PATCH" {
					methodStyle = methodStyle.Foreground(types.DefaultTheme.Warning)
				}

				pathStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text).Width(pathWidth)

				statusStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Success).Width(8)
				if strings.HasPrefix(status, "4") || strings.HasPrefix(status, "5") {
					statusStyle = statusStyle.Foreground(types.DefaultTheme.Error)
				}

				bytesStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Width(10).Align(lipgloss.Right)
				latencyStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Width(10).Align(lipgloss.Right)

				timeDisplay := timestamp
				if len(timestamp) > 10 {
					timeDisplay = timestamp[11:] // Strip date "YYYY/MM/DD "
				}

				if len(path) > pathWidth {
					path = path[:pathWidth-3] + "..."
				}

				return fmt.Sprintf("%s │ %s │ %s │ %s │ %s │ %s",
					timeStyle.Render(timeDisplay),
					methodStyle.Render(method),
					pathStyle.Render(path),
					statusStyle.Render(status),
					bytesStyle.Render(bytesStr+" B"),
					latencyStyle.Render(latency),
				)
			}
		}
	}
	return lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(line)
}

func (m *GlobalLogsModel) refreshViewport() {
	var filtered []string
	matchedLines := shared.FilterSubstring(m.Search.Value(), m.Lines)

	for _, l := range matchedLines {
		parsedLine := parseLogLine(l, m.Viewport.Width)
		filtered = append(filtered, parsedLine)
	}

	if len(filtered) == 0 {
		emptyMsg := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render("No logs found.")
		content := lipgloss.Place(m.Viewport.Width, m.Viewport.Height-2, lipgloss.Center, lipgloss.Center, emptyMsg)
		filtered = append(filtered, content)
	}

	oldYOffset := m.Viewport.YOffset
	m.Viewport.SetContent(strings.Join(filtered, "\n"))

	if m.AutoTail {
		m.Viewport.GotoBottom()
	} else {
		m.Viewport.YOffset = oldYOffset
	}
}

func (m *GlobalLogsModel) SetFocus(focus bool) {}
