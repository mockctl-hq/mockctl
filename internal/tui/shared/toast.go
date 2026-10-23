package shared

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// ShowToastMsg is an alias to types.ShowToastMsg for backwards compatibility.
type ShowToastMsg = types.ShowToastMsg

// HideToastMsg is emitted when the 3-second toast timer expires.
type HideToastMsg struct{}

// ToastModel handles transient flash notifications.
type ToastModel struct {
	tuiCtx  types.TUIContext
	Message string
	IsError bool
	Visible bool
}

func NewToastModel(ctx types.TUIContext) *ToastModel {
	return &ToastModel{
		tuiCtx: ctx,
	}
}

func (m *ToastModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *ToastModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case types.ShowToastMsg:
		m.Message = msg.Message
		m.IsError = msg.IsError
		m.Visible = true
		return m, m.hideTimer()
	case HideToastMsg:
		m.Visible = false
		return m, nil
	}
	return m, nil
}

func (m *ToastModel) hideTimer() tea.Cmd {
	return tea.Tick(time.Second*3, func(_ time.Time) tea.Msg {
		return HideToastMsg{}
	})
}

func (m *ToastModel) View() string {
	if !m.Visible {
		return ""
	}
	
	style := lipgloss.NewStyle().
		Foreground(types.DefaultTheme.Text).
		Padding(0, 1)

	if m.IsError {
		style = style.Background(types.DefaultTheme.Error)
		return style.Render("❌ " + m.Message)
	}
	
	style = style.Background(types.DefaultTheme.Success)
	return style.Render("✅ " + m.Message)
}

func (m *ToastModel) SetFocus(focus bool) {}
