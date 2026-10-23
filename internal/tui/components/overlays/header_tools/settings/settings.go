package settings

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type SettingsModel struct {
	tuiCtx      types.TUIContext
	Size        tea.WindowSizeMsg
	Active      bool
	IsFocused   bool
	ActiveTab   int // 0: Theme, 1: Network, 2: Editor, 3: Tester
	FocusPane   int // 0: Sidebar, 1: Content

	ThemeTab   *ThemeTabModel
	NetworkTab *NetworkTabModel
	EditorTab  *EditorTabModel
	TesterTab  *TesterTabModel
}

func NewSettingsModel(ctx types.TUIContext) *SettingsModel {
	m := &SettingsModel{
		tuiCtx:     ctx,
		ThemeTab:   NewThemeTabModel(ctx),
		NetworkTab: NewNetworkTabModel(ctx),
		EditorTab:  NewEditorTabModel(ctx),
		TesterTab:  NewTesterTabModel(ctx),
	}
	if ctx.Settings != nil {
		m.syncFromConfig(ctx.Settings)
	}
	return m
}

// syncFromConfig reads live config values into the tab models so the UI reflects current state.
func (m *SettingsModel) syncFromConfig(cfg *types.SettingsConfig) {
	m.ThemeTab.ActiveTheme = cfg.ActiveTheme
	m.ThemeTab.Animations = cfg.UIAnimations
	m.NetworkTab.GlobalCORS = cfg.GlobalCORS
	m.NetworkTab.StrictValidate = cfg.StrictOpenAPIValidation
	m.NetworkTab.TimeoutIndex = timeoutToIndex(cfg.RequestTimeout)
	m.NetworkTab.HistoryBuffer = historyLimitToIndex(cfg.HistoryLimit)
	m.EditorTab.TabSizeIndex = tabSizeToIndex(cfg.TabSize)
	m.EditorTab.LineWrapping = cfg.LineWrapping
	m.EditorTab.AutoFormatSave = cfg.AutoFormatOnSave
	m.TesterTab.FollowRedirects = cfg.FollowRedirects
}

// buildConfig reads the current tab UI state back into a SettingsConfig.
func (m *SettingsModel) buildConfig() *types.SettingsConfig {
	return &types.SettingsConfig{
		ActiveTheme:             m.ThemeTab.ActiveTheme,
		UIAnimations:            m.ThemeTab.Animations,
		GlobalCORS:              m.NetworkTab.GlobalCORS,
		StrictOpenAPIValidation: m.NetworkTab.StrictValidate,
		RequestTimeout:          indexToTimeout(m.NetworkTab.TimeoutIndex),
		HistoryLimit:            indexToHistoryLimit(m.NetworkTab.HistoryBuffer),
		TabSize:                 indexToTabSize(m.EditorTab.TabSizeIndex),
		LineWrapping:            m.EditorTab.LineWrapping,
		AutoFormatOnSave:        m.EditorTab.AutoFormatSave,
		FollowRedirects:         m.TesterTab.FollowRedirects,
	}
}

func (m *SettingsModel) Init(ctx types.TUIContext) tea.Cmd {
	return tea.Batch(
		m.ThemeTab.Init(ctx),
		m.NetworkTab.Init(ctx),
		m.EditorTab.Init(ctx),
		m.TesterTab.Init(ctx),
	)
}

func (m *SettingsModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	// Always track window size regardless of Active state.
	// Without this, m.Size stays {0,0} and the overlay renders tiny on first open.
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		m.Size = sizeMsg
	}

	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Already handled above
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Escape) || msg.String() == "q" {
			if m.FocusPane == 1 {
				// If focused on content, Esc returns focus to sidebar
				m.FocusPane = 0
				return m, nil
			}
			// If focused on sidebar, Esc/q closes settings — emit SettingsChangedMsg
			m.Active = false
			cfg := m.buildConfig()
			// Update the live config pointer so other components see the change immediately
			if m.tuiCtx.Settings != nil {
				*m.tuiCtx.Settings = *cfg
			}
			return m, func() tea.Msg {
				return types.SettingsChangedMsg{Config: cfg}
			}
		}
		
		if msg.String() == "left" || msg.String() == "h" {
			m.FocusPane = 0
			return m, nil
		} else if msg.String() == "right" || msg.String() == "l" || msg.String() == "enter" {
			m.FocusPane = 1
			// If enter was pressed, don't pass it down immediately so we just focus
			if msg.String() == "enter" {
				return m, nil
			}
		}
		
		if m.FocusPane == 0 {
			// Sidebar Navigation
			if msg.String() == "up" || msg.String() == "k" || key.Matches(msg, types.GlobalKeys.TabBackward) {
				if m.ActiveTab > 0 {
					m.ActiveTab--
				} else {
					m.ActiveTab = 3
				}
			} else if msg.String() == "down" || msg.String() == "j" || key.Matches(msg, types.GlobalKeys.TabForward) {
				if m.ActiveTab < 3 {
					m.ActiveTab++
				} else {
					m.ActiveTab = 0
				}
			}
			return m, nil
		}
	}

	// Content Navigation
	var cmd tea.Cmd
	if m.FocusPane == 1 {
		switch m.ActiveTab {
		case 0:
			_, cmd = m.ThemeTab.Update(msg)
		case 1:
			_, cmd = m.NetworkTab.Update(msg)
		case 2:
			_, cmd = m.EditorTab.Update(msg)
		case 3:
			_, cmd = m.TesterTab.Update(msg)
		}
	}

	return m, cmd
}

func (m *SettingsModel) View() string {
	if !m.Active {
		return ""
	}

	dims := shared.ComputeOverlayDimensions(m.Size, 30, 10)

	var instructionText string
	if m.FocusPane == 0 {
		instructionText = "Use ↑/↓ to navigate tabs | Enter/→ to enter tab | Esc/q to close"
	} else {
		instructionText = "Use ↑/↓ to navigate options | Space to toggle | Esc/← to go back"
	}

	// Sidebar
	sidebarWidth := 20
	activeTabStyle := lipgloss.NewStyle().
		Background(types.DefaultTheme.Primary).
		Foreground(types.DefaultTheme.Bg).
		Bold(true).
		PaddingLeft(2).
		Width(sidebarWidth)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(types.DefaultTheme.Text).
		PaddingLeft(2).
		Width(sidebarWidth)

	tabs := []string{"Theme", "Network", "Editor", "Tester"}
	var renderedTabs []string

	// Add some top padding to sidebar
	renderedTabs = append(renderedTabs, "")
	for i, tab := range tabs {
		if i == m.ActiveTab {
			renderedTabs = append(renderedTabs, activeTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, inactiveTabStyle.Render(tab))
		}
		renderedTabs = append(renderedTabs, "") // space between items
	}
	sidebarContent := lipgloss.JoinVertical(lipgloss.Left, renderedTabs...)
	sidebar := lipgloss.NewStyle().Width(sidebarWidth).Height(dims.MiddleHeight).Render(sidebarContent)

	// Main Content
	var activeView string
	switch m.ActiveTab {
	case 0:
		activeView = m.ThemeTab.View()
	case 1:
		activeView = m.NetworkTab.View()
	case 2:
		activeView = m.EditorTab.View()
	case 3:
		activeView = m.TesterTab.View()
	}

	mainContentWidth := dims.InnerWidth - sidebarWidth - 3 // 3 for vertical divider
	mainContent := lipgloss.NewStyle().Width(mainContentWidth).Height(dims.MiddleHeight).PaddingLeft(2).Render("\n" + activeView)

	// Vertical Divider
	var vDividerLines []string
	for i := 0; i < dims.MiddleHeight; i++ {
		vDividerLines = append(vDividerLines, lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render("│"))
	}
	vDivider := lipgloss.JoinVertical(lipgloss.Left, vDividerLines...)

	middle := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, vDivider, mainContent)

	return shared.RenderOverlayModal(
		m.Size,
		dims.InnerWidth,
		dims.InnerHeight,
		"Settings (Mock:ctl)",
		"",
		middle,
		instructionText,
	)
}

func (m *SettingsModel) SetFocus(focus bool) {
	m.IsFocused = focus
}
