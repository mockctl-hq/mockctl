package overlays

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

func (m *ImportSpecModel) updateSearch() {
	query := strings.TrimSpace(m.Input.Value())
	if query == "" {
		m.Matches = nil
		m.CursorIndex = 0
		return
	}

	matches := shared.FilterSubstring(query, m.AllFiles)
	if len(matches) > 10 {
		matches = matches[:10]
	}
	m.Matches = matches
	if m.CursorIndex >= len(m.Matches) {
		m.CursorIndex = len(m.Matches) - 1
	}
	if m.CursorIndex < 0 {
		m.CursorIndex = 0
	}
}

func (m *ImportSpecModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	case FilesLoadedMsg:
		m.AllFiles = msg.Files
		return m, nil
	case types.SettingsChangedMsg:
		m.Input.PromptStyle = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Background(types.DefaultTheme.Bg)
		m.Input.TextStyle = lipgloss.NewStyle().Foreground(types.DefaultTheme.Text).Background(types.DefaultTheme.Bg)
		m.Input.PlaceholderStyle = lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Background(types.DefaultTheme.Bg)
		m.Input.Cursor.Style = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Background(types.DefaultTheme.Bg)
		return m, nil
	}

	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Escape) {
			m.Active = false
			return m, nil
		}

		if msg.String() == "down" && len(m.Matches) > 0 {
			m.CursorIndex++
			if m.CursorIndex >= len(m.Matches) {
				m.CursorIndex = len(m.Matches) - 1
			}
			return m, nil
		}

		if msg.String() == "up" && len(m.Matches) > 0 {
			m.CursorIndex--
			if m.CursorIndex < 0 {
				m.CursorIndex = 0
			}
			return m, nil
		}

		if msg.String() == "enter" {
			var path string

			// If we have matches and we are focused on one
			if len(m.Matches) > 0 {
				path = m.Matches[m.CursorIndex]
			} else {
				path = strings.TrimSpace(m.Input.Value())
			}

			if path == "" {
				m.ErrorMsg = "Path cannot be empty"
				return m, nil
			}

			// Use the full path as requested by the user, but replace home dir with ~
			filename := path
			if home, err := os.UserHomeDir(); err == nil && home != "" && strings.HasPrefix(filename, home) {
				filename = "~" + strings.TrimPrefix(filename, home)
			}

			if filename == "" {
				filename = "api.yaml"
			}

			// Read file
			file, err := os.Open(path)
			if err != nil {
				m.ErrorMsg = "Failed to open file: " + err.Error()
				return m, nil
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				m.ErrorMsg = "Failed to read file: " + err.Error()
				return m, nil
			}

			// Save to core
			if m.tuiCtx.Core != nil && m.ProjectName != "" {
				err = m.tuiCtx.Core.UpdateProjectSpec(context.Background(), m.ProjectName, filename, content)
				if err != nil {
					m.ErrorMsg = "Engine rejected spec: " + err.Error()
					return m, nil
				}

				// Success
				m.Active = false

				// Return a message that the project was updated, so it reloads in the editor
				// and also refresh the sidebar endpoints tree.
				return m, tea.Batch(
					func() tea.Msg { return types.ProjectStructureChangedMsg{ProjectName: m.ProjectName} },
					func() tea.Msg { return types.ProjectSelectedMsg{ProjectName: m.ProjectName, FileName: filename} },
				)
			}
		}
	}

	prevVal := m.Input.Value()
	m.Input, cmd = m.Input.Update(msg)

	if m.Input.Value() != prevVal {
		m.updateSearch()
	}

	return m, cmd
}
