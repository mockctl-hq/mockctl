package workspace

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// YamlEditorModel manages the YAML syntax highlighting and viewport scrolling.
type YamlEditorModel struct {
	tuiCtx    types.TUIContext
	Size      tea.WindowSizeMsg
	IsFocused bool

	Viewport viewport.Model

	// File State
	RawContent         string
	HighlightedContent string
	activeProject      string

	// Editor Settings
	LineWrapping     bool
	TabSize          int
	AutoFormatOnSave bool
}

// Ensure interface compliance
var _ types.Component = (*YamlEditorModel)(nil)

func NewYamlEditorModel(ctx types.TUIContext) *YamlEditorModel {
	vp := viewport.New(0, 0)
	vp.YPosition = 0

	lineWrapping := false
	tabSize := 2
	autoFormat := true
	if ctx.Settings != nil {
		lineWrapping = ctx.Settings.LineWrapping
		if ctx.Settings.TabSize > 0 {
			tabSize = ctx.Settings.TabSize
		}
		autoFormat = ctx.Settings.AutoFormatOnSave
	}

	return &YamlEditorModel{
		tuiCtx:           ctx,
		Viewport:         vp,
		LineWrapping:     lineWrapping,
		TabSize:          tabSize,
		AutoFormatOnSave: autoFormat,
	}
}

func (m *YamlEditorModel) Init(ctx types.TUIContext) tea.Cmd {
	m.tuiCtx = ctx
	return nil
}

func (m *YamlEditorModel) SetFocus(focus bool) {
	m.IsFocused = focus
}

func (m *YamlEditorModel) LineWrappingEnabled() bool {
	return m.LineWrapping
}

func (m *YamlEditorModel) CurrentTabSize() int {
	return m.TabSize
}

func (m *YamlEditorModel) AutoFormatEnabled() bool {
	return m.AutoFormatOnSave
}

