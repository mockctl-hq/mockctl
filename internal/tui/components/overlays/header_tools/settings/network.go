package settings

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type NetworkTabModel struct {
	tuiCtx         types.TUIContext
	GlobalCORS     bool
	StrictValidate bool
	TimeoutIndex   int // 0: 5s, 1: 15s, 2: 30s, 3: 60s
	HistoryBuffer  int // 0: 100, 1: 1000, 2: 5000
	ActiveField    int // 0: CORS, 1: Strict, 2: Timeout, 3: History
}

func NewNetworkTabModel(ctx types.TUIContext) *NetworkTabModel {
	return &NetworkTabModel{
		tuiCtx:         ctx,
		GlobalCORS:     true,
		StrictValidate: true,
		TimeoutIndex:   1,
		HistoryBuffer:  1,
	}
}

func (m *NetworkTabModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *NetworkTabModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Up) {
			if m.ActiveField > 0 {
				m.ActiveField--
			} else {
				m.ActiveField = 3
			}
		} else if key.Matches(msg, types.GlobalKeys.Down) {
			if m.ActiveField < 3 {
				m.ActiveField++
			} else {
				m.ActiveField = 0
			}
		} else if msg.String() == " " || msg.String() == "enter" {
			if m.ActiveField == 0 {
				m.GlobalCORS = !m.GlobalCORS
			} else if m.ActiveField == 1 {
				m.StrictValidate = !m.StrictValidate
			} else if m.ActiveField == 2 {
				m.TimeoutIndex = (m.TimeoutIndex + 1) % 4
			} else if m.ActiveField == 3 {
				m.HistoryBuffer = (m.HistoryBuffer + 1) % 3
			}
		}
	}
	return m, nil
}

func (m *NetworkTabModel) View() string {
	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text)
	dimStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	corsStr := "[ ] Enable Global CORS"
	if m.GlobalCORS {
		corsStr = "[x] Enable Global CORS"
	}
	if m.ActiveField == 0 {
		corsStr = activeStyle.Render(corsStr)
	} else {
		corsStr = inactiveStyle.Render(corsStr)
	}

	strictStr := "[ ] Strict OpenAPI Validation"
	if m.StrictValidate {
		strictStr = "[x] Strict OpenAPI Validation"
	}
	if m.ActiveField == 1 {
		strictStr = activeStyle.Render(strictStr)
	} else {
		strictStr = inactiveStyle.Render(strictStr)
	}

	timeouts := []string{"5s", "15s", "30s", "60s"}
	timeStr := fmt.Sprintf("HTTP Timeout: [%s]", timeouts[m.TimeoutIndex])
	if m.ActiveField == 2 {
		timeStr = activeStyle.Render(timeStr)
	} else {
		timeStr = inactiveStyle.Render(timeStr)
	}

	buffers := []string{"100 events", "1000 events", "5000 events"}
	bufStr := fmt.Sprintf("History Ring Buffer: [%s]", buffers[m.HistoryBuffer])
	if m.ActiveField == 3 {
		bufStr = activeStyle.Render(bufStr)
	} else {
		bufStr = inactiveStyle.Render(bufStr)
	}
	
	helpStr := dimStyle.Render("Press Space/Enter to toggle")

	return lipgloss.JoinVertical(lipgloss.Left, corsStr, "", strictStr, "", timeStr, "", bufStr, "", helpStr)
}

func (m *NetworkTabModel) SetFocus(focus bool) {}
