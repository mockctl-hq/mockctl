package profiler

import (
	"fmt"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type TickMsg time.Time

type GoProfilerModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	HeapProg progress.Model
	MemStats runtime.MemStats
}

var _ types.Component = (*GoProfilerModel)(nil)

func NewGoProfilerModel(ctx types.TUIContext) *GoProfilerModel {
	p := progress.New(progress.WithDefaultGradient())
	return &GoProfilerModel{
		tuiCtx:   ctx,
		HeapProg: p,
	}
}

func (m *GoProfilerModel) Init(ctx types.TUIContext) tea.Cmd {
	return m.tickCmd()
}

func (m *GoProfilerModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m *GoProfilerModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		innerWidth := msg.Width - 6
		if innerWidth < 10 {
			innerWidth = 10
		}
		progWidth := innerWidth / 2
		if progWidth > 60 {
			progWidth = 60
		}
		if progWidth < 20 {
			progWidth = 20
		}
		m.HeapProg.Width = progWidth
	case TickMsg:
		runtime.ReadMemStats(&m.MemStats)
		return m, m.tickCmd()
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Escape) || key.Matches(msg, types.GlobalKeys.Quit) || msg.String() == "q" {
			// Instruct root to close overlay
			return m, func() tea.Msg { return types.OverlayStateEnum(types.OverlayNone) }
		}
	}
	return m, cmd
}

func (m *GoProfilerModel) View() string {
	dims := shared.ComputeOverlayDimensions(m.Size, 10, 5)

	// Convert bytes to MB
	allocMB := float64(m.MemStats.Alloc) / 1024 / 1024
	sysMB := float64(m.MemStats.Sys) / 1024 / 1024

	heapPercent := allocMB / sysMB
	if sysMB == 0 {
		heapPercent = 0
	}

	statsStr := fmt.Sprintf("Goroutines: %d\nHeap Alloc: %.2f MB\nSys Memory: %.2f MB\nGC Pauses: %d",
		runtime.NumGoroutine(), allocMB, sysMB, m.MemStats.NumGC)

	progStr := m.HeapProg.ViewAs(heapPercent)

	statsBlock := lipgloss.JoinVertical(lipgloss.Left,
		statsStr,
		"",
		"Heap Utilization:",
		progStr,
	)

	centeredStats := lipgloss.Place(dims.InnerWidth, dims.MiddleHeight, lipgloss.Center, lipgloss.Center, statsBlock)

	return shared.RenderOverlayModal(
		m.Size,
		dims.InnerWidth,
		dims.InnerHeight,
		"Go Runtime Memory Profiler",
		"",
		centeredStats,
		"Press Esc or 'q' to close",
	)
}

func (m *GoProfilerModel) SetFocus(focus bool) {
	// overlays always focused when active
}
