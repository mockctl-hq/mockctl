package workspace

import (
	"context"
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

// HighlightMsg is returned from async highlighting.
type HighlightMsg struct {
	Content string
}

func (m *YamlEditorModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height
		m.updateViewportContent()

	case types.SettingsChangedMsg:
		if msg.Config != nil {
			changed := false
			if m.LineWrapping != msg.Config.LineWrapping {
				m.LineWrapping = msg.Config.LineWrapping
				changed = true
			}
			newTabSize := msg.Config.TabSize
			if newTabSize <= 0 {
				newTabSize = 2
			}
			if m.TabSize != newTabSize {
				m.TabSize = newTabSize
				changed = true
			}
			m.AutoFormatOnSave = msg.Config.AutoFormatOnSave
			if changed {
				if m.LineWrapping {
					m.Viewport.SetXOffset(0)
				}
				m.updateViewportContent()
			}
		}
		return m, nil

	case types.ProjectSelectedMsg:
		m.activeProject = msg.ProjectName
		if msg.ProjectName == "" {
			return m, m.SetContent("")
		}
		if m.tuiCtx.Core != nil {
			projects, err := m.tuiCtx.Core.GetProjects(context.Background())
			if err == nil {
				if proj, ok := projects[msg.ProjectName]; ok {
					return m, m.SetContent(proj.OpenAPISpec)
				}
			}
		}
		return m, nil

	case HighlightMsg:
		m.HighlightedContent = msg.Content
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		if m.IsFocused && key.Matches(msg, types.GlobalKeys.Format) {
			if m.RawContent != "" {
				if formatted, err := FormatPayload(m.RawContent, true, m.TabSize); err == nil {
					_ = m.SetContent(formatted)
					if m.tuiCtx.Core != nil && m.activeProject != "" {
						_ = m.tuiCtx.Core.UpdateProjectSpec(context.Background(), m.activeProject, "api.yaml", []byte(formatted))
					}
				}
			}
			return m, nil
		}
	}

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *YamlEditorModel) SetContent(content string) tea.Cmd {
	m.RawContent = content
	
	// Syntax Highlighting (Async for large files, sync for small)
	if len(content) < 100*1024 && content != "" {
		var buf strings.Builder
		err := quick.Highlight(&buf, content, "yaml", "terminal256", "monokai")
		if err == nil {
			m.HighlightedContent = buf.String()
		} else {
			m.HighlightedContent = content
		}
	} else {
		m.HighlightedContent = content
	}
	m.updateViewportContent()
	return nil
}

func (m *YamlEditorModel) updateViewportContent() {
	if m.HighlightedContent == "" {
		m.Viewport.SetContent("")
		return
	}

	lines := strings.Split(m.HighlightedContent, "\n")
	tabSpaces := strings.Repeat(" ", m.TabSize)
	var numbered strings.Builder
	availWidth := m.Viewport.Width - 7

	for i, line := range lines {
		if m.TabSize > 0 && strings.Contains(line, "\t") {
			line = strings.ReplaceAll(line, "\t", tabSpaces)
		}

		if m.LineWrapping && availWidth > 10 {
			wrapped := ansi.Wrap(line, availWidth, "")
			sublines := strings.Split(wrapped, "\n")
			for j, sub := range sublines {
				if j == 0 {
					num := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(fmt.Sprintf("%4d │ ", i+1))
					numbered.WriteString(num + sub + "\n")
				} else {
					cont := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render("     │ ")
					numbered.WriteString(cont + sub + "\n")
				}
			}
		} else {
			num := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Render(fmt.Sprintf("%4d │ ", i+1))
			numbered.WriteString(num + line + "\n")
		}
	}
	m.Viewport.SetContent(numbered.String())
}
