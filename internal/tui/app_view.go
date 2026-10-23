package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// View renders the entire TUI grid layout.
func (m *RootModel) View() string {
	// 1. Termux Shields & Defensive Bounds
	if m.Size.Width == 0 || m.Size.Height == 0 {
		return "Initializing..." // Wait for first WindowSizeMsg
	}

	// Death Spiral Prevention & Minimum Size Obstruction Modal
	if m.Size.Width < 120 || m.Size.Height < 24 {
		m.ScreenWarning.Size = m.Size
		return m.ScreenWarning.View()
	}

	// 2. Active Overlay Preemption
	if m.ActiveOverlay != types.OverlayNone {
		var overlayStr string
		switch m.ActiveOverlay {
		case types.OverlaySettings:
			overlayStr = m.Settings.View()
		case types.OverlayHistory:
			overlayStr = m.History.View()
		case types.OverlayGlobalLog:
			overlayStr = m.GlobalLog.View()
		case types.OverlayProfiler:
			overlayStr = m.Profiler.View()
		case types.OverlayStateInspector:
			overlayStr = m.StateInspector.View()
		case types.OverlayImportSpec:
			overlayStr = m.ImportSpec.View()
		case types.OverlayDaemonCrashed:
			overlayStr = m.DaemonCrashed.View()
		case types.OverlayDeleteProject:
			overlayStr = m.DeleteProject.View()
		}
		// Render overlay centered over an empty/obscured background
		return lipgloss.Place(m.Size.Width, m.Size.Height, lipgloss.Center, lipgloss.Center, overlayStr)
	}

	// 3. VS Code Grid Layout (Phase 2 Setup)
	// Helper to determine border style based on focus
	borderStyle := func(target types.FocusRingEnum) lipgloss.Style {
		if m.FocusRing == target {
			return types.GlobalStyles.FocusedBorder
		}
		return types.GlobalStyles.UnfocusedBorder
	}

	headerStr := m.Header.View()

	m.Footer.SetFocusContext(
		m.FocusRing,
		m.ActiveOverlay,
		m.LeftSidebar.Search.Active,
		m.LeftSidebar.Scaffold.Active,
		m.RightSidebar.Icons.ActiveIndex,
		m.RightSidebar.IsIconsFocused,
	)
	footerStr := m.Footer.View()

	// The remaining height for the middle sections
	middleHeight := m.Size.Height - lipgloss.Height(headerStr) - lipgloss.Height(footerStr)
	if middleHeight < 0 {
		middleHeight = 0
	}

	m.LeftSidebar.IsCollapsed = !m.LeftSidebarVisible
	m.RightSidebar.IsCollapsed = !m.RightSidebarVisible
	m.LeftSidebar.Size.Height = middleHeight
	m.RightSidebar.Size.Height = middleHeight

	var middleParts []string
	leftStr := m.LeftSidebar.View()
	middleParts = append(middleParts, leftStr)
	leftWidth := lipgloss.Width(leftStr)

	rightStr := m.RightSidebar.View()
	rightWidth := lipgloss.Width(rightStr)

	// Calculate center workspace width
	centerWidth := m.Size.Width - leftWidth - rightWidth
	if centerWidth < 0 {
		centerWidth = 0
	}

	centerWorkspaceStr := borderStyle(types.FocusWorkspace).
		Width(centerWidth).
		Height(middleHeight).
		Border(lipgloss.NormalBorder(), false, false, false, false).
		Render(m.Workspace.View())
	middleParts = append(middleParts, centerWorkspaceStr)
	middleParts = append(middleParts, rightStr)

	// Horizontal stitch
	middleSection := lipgloss.JoinHorizontal(lipgloss.Top, middleParts...)

	// Vertical stitch
	return lipgloss.JoinVertical(lipgloss.Left, headerStr, middleSection, footerStr)
}
