package tester

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace/tester/protocols/http"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// ApiTesterModel acts as the bottom panel container managing protocol tabs.
type ApiTesterModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	HttpTester *http.HttpTesterModel
	IsFocused  bool
}

var _ types.Component = (*ApiTesterModel)(nil)

func NewApiTesterModel(ctx types.TUIContext) *ApiTesterModel {
	return &ApiTesterModel{
		tuiCtx:     ctx,
		HttpTester: http.NewHttpTesterModel(ctx),
	}
}

func (m *ApiTesterModel) Init(ctx types.TUIContext) tea.Cmd {
	return m.HttpTester.Init(ctx)
}

func (m *ApiTesterModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		m.HttpTester.Update(tea.WindowSizeMsg{
			Width:  msg.Width - 2, // minus border
			Height: msg.Height - 3, // minus protocol tab (1) and borders (2)
		})
	}

	if _, isWindowSize := msg.(tea.WindowSizeMsg); !isWindowSize {
		_, cmd = m.HttpTester.Update(msg)
	}

	return m, cmd
}

func (m *ApiTesterModel) View() string {
	protocolTab := lipgloss.NewStyle().
		Background(types.DefaultTheme.Primary).
		Foreground(types.DefaultTheme.Bg).
		Padding(0, 1).
		Render(" HTTP/REST ")
		
	border := types.GlobalStyles.UnfocusedBorder
	if m.IsFocused {
		border = types.GlobalStyles.FocusedBorder
	}
	
	testerStr := m.HttpTester.View()
	return border.Render(lipgloss.JoinVertical(lipgloss.Left, protocolTab, testerStr))
}

func (m *ApiTesterModel) SetFocus(focus bool) {
	m.IsFocused = focus
	m.HttpTester.SetFocus(focus)
}
