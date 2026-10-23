package footer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type Shortcut struct {
	Key  string
	Desc string
}

type FooterModel struct {
	tuiCtx             types.TUIContext
	Size               tea.WindowSizeMsg
	FocusRing          types.FocusRingEnum
	ActiveOverlay      types.OverlayStateEnum
	LeftSearchActive   bool
	LeftScaffoldActive bool
	RightActiveTab     int
	RightIconsFocused  bool
}

var _ types.Component = (*FooterModel)(nil)

func NewFooterModel(ctx types.TUIContext) *FooterModel {
	return &FooterModel{
		tuiCtx:    ctx,
		FocusRing: types.FocusWorkspace,
	}
}

func (m *FooterModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *FooterModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	}
	return m, nil
}

func (m *FooterModel) SetFocusContext(
	focus types.FocusRingEnum,
	overlay types.OverlayStateEnum,
	leftSearch bool,
	leftScaffold bool,
	rightTab int,
	rightIconsFocused bool,
) {
	m.FocusRing = focus
	m.ActiveOverlay = overlay
	m.LeftSearchActive = leftSearch
	m.LeftScaffoldActive = leftScaffold
	m.RightActiveTab = rightTab
	m.RightIconsFocused = rightIconsFocused
}

func (m *FooterModel) getShortcuts() (string, []Shortcut) {
	if m.ActiveOverlay != types.OverlayNone {
		switch m.ActiveOverlay {
		case types.OverlaySettings:
			return "[Settings]", []Shortcut{
				{"↑/↓", "Nav"},
				{"Space/Enter", "Toggle"},
				{"Tab", "Switch Tab"},
				{"Esc", "Close"},
			}
		case types.OverlayHistory:
			return "[History]", []Shortcut{
				{"↑/↓", "Select"},
				{"/", "Search"},
				{"Esc", "Close"},
			}
		case types.OverlayGlobalLog:
			return "[Logs]", []Shortcut{
				{"↑/↓", "Scroll"},
				{"f", "Auto-Tail"},
				{"/", "Search"},
				{"Esc", "Close"},
			}
		case types.OverlayProfiler:
			return "[Profiler]", []Shortcut{
				{"↑/↓", "Scroll"},
				{"Esc", "Close"},
			}
		case types.OverlayStateInspector:
			return "[Inspector]", []Shortcut{
				{"↑/↓", "Scroll"},
				{"Esc", "Close"},
			}
		case types.OverlayImportSpec:
			return "[Import Spec]", []Shortcut{
				{"↑/↓", "Select"},
				{"Enter", "Import"},
				{"Esc", "Cancel"},
			}
		case types.OverlayDaemonCrashed:
			return "[Daemon Crashed]", []Shortcut{
				{"R", "Retry"},
				{"Ctrl+Q", "Quit"},
			}
		case types.OverlayDeleteProject:
			return "[Delete Project]", []Shortcut{
				{"Type Name", "Confirm"},
				{"Enter", "Delete"},
				{"Esc", "Cancel"},
			}
		default:
			return "[Overlay]", []Shortcut{
				{"Esc", "Close"},
			}
		}
	}

	switch m.FocusRing {
	case types.FocusLeftPanel:
		if m.LeftSearchActive {
			return "[Search]", []Shortcut{
				{"Type", "Filter"},
				{"Enter", "Done"},
				{"Esc", "Cancel"},
			}
		}
		if m.LeftScaffoldActive {
			return "[New Project]", []Shortcut{
				{"Type", "Name"},
				{"Enter", "Create"},
				{"Esc", "Cancel"},
			}
		}
		return "[Left Panel]", []Shortcut{
			{"↑/↓", "Nav"},
			{"Enter", "Open"},
			{"+", "New"},
			{"d", "Delete"},
			{"/", "Search"},
			{"Shift+←", "Collapse"},
			{"Tab", "Next Panel"},
			{"Ctrl+Q", "Quit"},
		}

	case types.FocusWorkspace:
		return "[Workspace]", []Shortcut{
			{"Ctrl+S", "Save"},
			{"Ctrl+F", "Format"},
			{"Ctrl+\\", "Tester"},
			{"Ctrl+W", "Close"},
			{"Tab", "Next Panel"},
			{"Ctrl+Q", "Quit"},
		}

	case types.FocusApiTester:
		return "[API Tester]", []Shortcut{
			{"Enter", "Send"},
			{"Ctrl+R", "Execute"},
			{"Tab", "Next Field"},
			{"Esc", "Cancel"},
			{"Ctrl+\\", "Hide"},
			{"Ctrl+Q", "Quit"},
		}

	case types.FocusRightPanel:
		if m.RightIconsFocused {
			return "[Right Panel]", []Shortcut{
				{"↑/↓", "Select Tab"},
				{"Enter/←", "Open Tab"},
				{"Tab", "Next Panel"},
				{"Ctrl+Q", "Quit"},
			}
		}
		switch m.RightActiveTab {
		case 0:
			return "[Chaos Controls]", []Shortcut{
				{"Space", "Toggle"},
				{"Tab", "Next Field"},
				{"0-9", "Edit"},
				{"Esc", "Tabs"},
				{"Ctrl+Q", "Quit"},
			}
		case 1:
			return "[Code Snippets]", []Shortcut{
				{"↑/↓", "Select Lang"},
				{"Enter/Ctrl+C", "Copy"},
				{"Esc", "Tabs"},
				{"Ctrl+Q", "Quit"},
			}
		case 2:
			return "[Static Overrides]", []Shortcut{
				{"Enter/Ctrl+S", "Save"},
				{"Ctrl+D", "Clear"},
				{"Esc", "Tabs"},
				{"Ctrl+Q", "Quit"},
			}
		default:
			return "[Right Panel]", []Shortcut{
				{"Esc", "Tabs"},
				{"Tab", "Next Panel"},
				{"Ctrl+Q", "Quit"},
			}
		}

	default:
		return "[Help]", []Shortcut{
			{"Ctrl+S", "Save"},
			{"Ctrl+\\", "Tester"},
			{"Tab", "Next Panel"},
			{"Ctrl+Q", "Quit"},
		}
	}
}

func (m *FooterModel) renderShortcuts(badge string, shortcuts []Shortcut) string {
	badgeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)
	sep := dimStyle.Render(" • ")

	prefix := " "
	if badge != "" {
		prefix += badgeStyle.Render(badge) + "  "
	}

	availWidth := m.Size.Width - lipgloss.Width(prefix) - 2
	if m.Size.Width <= 0 {
		availWidth = 80 - lipgloss.Width(prefix) - 2
	}
	if availWidth <= 0 {
		return prefix
	}

	var renderedItems []string
	currentWidth := 0
	sepWidth := lipgloss.Width(sep)

	for i, sc := range shortcuts {
		itemStr := dimStyle.Render(sc.Key + ": " + sc.Desc)
		itemWidth := lipgloss.Width(itemStr)

		cost := itemWidth
		if i > 0 {
			cost += sepWidth
		}

		if currentWidth+cost > availWidth {
			break
		}

		currentWidth += cost
		renderedItems = append(renderedItems, itemStr)
	}

	return prefix + strings.Join(renderedItems, sep)
}

func (m *FooterModel) View() string {
	badge, shortcuts := m.getShortcuts()
	line := m.renderShortcuts(badge, shortcuts)

	return lipgloss.NewStyle().
		Width(m.Size.Width).
		Height(1). // 1 text line + 1 top border (lipgloss adds border to the height)
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(types.DefaultTheme.Dim).
		Render(line)
}

func (m *FooterModel) SetFocus(focus bool) {}
