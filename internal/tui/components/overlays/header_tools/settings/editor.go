package settings

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type EditorTabModel struct {
	tuiCtx         types.TUIContext
	TabSizeIndex   int // 0: 2, 1: 4, 2: 8
	LineWrapping   bool
	AutoFormatSave bool
	ActiveField    int // 0: TabSize, 1: Wrapping, 2: AutoFormat
}

func NewEditorTabModel(ctx types.TUIContext) *EditorTabModel {
	return &EditorTabModel{
		tuiCtx:         ctx,
		TabSizeIndex:   0, // 2 spaces
		LineWrapping:   false,
		AutoFormatSave: true,
	}
}

func (m *EditorTabModel) Init(ctx types.TUIContext) tea.Cmd {
	return nil
}

func (m *EditorTabModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Up) {
			if m.ActiveField > 0 {
				m.ActiveField--
			} else {
				m.ActiveField = 2
			}
		} else if key.Matches(msg, types.GlobalKeys.Down) {
			if m.ActiveField < 2 {
				m.ActiveField++
			} else {
				m.ActiveField = 0
			}
		} else if msg.String() == " " || msg.String() == "enter" {
			if m.ActiveField == 0 {
				m.TabSizeIndex = (m.TabSizeIndex + 1) % 3
			} else if m.ActiveField == 1 {
				m.LineWrapping = !m.LineWrapping
			} else if m.ActiveField == 2 {
				m.AutoFormatSave = !m.AutoFormatSave
			}
		}
	}
	return m, nil
}

func (m *EditorTabModel) View() string {
	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Text)
	dimStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	sizes := []string{"2 spaces", "4 spaces", "8 spaces"}
	tabStr := fmt.Sprintf("Tab Size: [%s]", sizes[m.TabSizeIndex])
	if m.ActiveField == 0 {
		tabStr = activeStyle.Render(tabStr)
	} else {
		tabStr = inactiveStyle.Render(tabStr)
	}

	wrapStr := "[ ] Word Wrap"
	if m.LineWrapping {
		wrapStr = "[x] Word Wrap"
	}
	if m.ActiveField == 1 {
		wrapStr = activeStyle.Render(wrapStr)
	} else {
		wrapStr = inactiveStyle.Render(wrapStr)
	}

	fmtStr := "[ ] Auto-format on Save"
	if m.AutoFormatSave {
		fmtStr = "[x] Auto-format on Save"
	}
	if m.ActiveField == 2 {
		fmtStr = activeStyle.Render(fmtStr)
	} else {
		fmtStr = inactiveStyle.Render(fmtStr)
	}

	helpStr := dimStyle.Render("Press Space/Enter to toggle")

	return lipgloss.JoinVertical(lipgloss.Left, tabStr, "", wrapStr, "", fmtStr, "", helpStr)
}

func (m *EditorTabModel) SetFocus(focus bool) {}
