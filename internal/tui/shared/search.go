package shared

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/sahilm/fuzzy"
)

// SearchModel provides an inline fuzzy search component.
type SearchModel struct {
	Input     textinput.Model
	IsFocused bool
	Active    bool
}

func NewSearchModel() *SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.Prompt = "🔍 "
	ti.CharLimit = 100

	return &SearchModel{
		Input: ti,
	}
}

func (m *SearchModel) Init(ctx types.TUIContext) tea.Cmd {
	return textinput.Blink
}

func (m *SearchModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	if !m.IsFocused || !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, types.GlobalKeys.Escape) {
			m.Active = false
			m.Input.SetValue("")
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m *SearchModel) View() string {
	if !m.Active {
		return ""
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(types.DefaultTheme.Dim).
		Width(24)

	if m.IsFocused {
		style = style.BorderForeground(types.DefaultTheme.Primary)
	}

	return style.Render(m.Input.View())
}

func (m *SearchModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if focus {
		m.Input.Focus()
	} else {
		m.Input.Blur()
	}
}

// Filter filters a slice of strings using the sahilm/fuzzy package.
func (m *SearchModel) Filter(items []string) []string {
	return FilterFuzzy(m.Input.Value(), items)
}

// FilterSubstring filters a slice of strings using case-insensitive substring matching.
func (m *SearchModel) FilterSubstring(items []string) []string {
	return FilterSubstring(m.Input.Value(), items)
}

// FilterSubstring performs a case-insensitive substring search over a slice of strings.
func FilterSubstring(query string, items []string) []string {
	if strings.TrimSpace(query) == "" {
		return items
	}
	term := strings.ToLower(strings.TrimSpace(query))
	var result []string
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), term) {
			result = append(result, item)
		}
	}
	return result
}

// FilterFuzzy performs a fuzzy match search using sahilm/fuzzy.
func FilterFuzzy(query string, items []string) []string {
	if strings.TrimSpace(query) == "" {
		return items
	}
	matches := fuzzy.Find(strings.TrimSpace(query), items)
	var result []string
	for _, match := range matches {
		result = append(result, match.Str)
	}
	return result
}
