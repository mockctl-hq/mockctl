package body

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

var pathParamRegex = regexp.MustCompile(`\{[^}]+\}`)

type resetCopyNoticeMsg struct{}
type copyResultMsg struct{ err error }

type CodeSnippetsModel struct {
	tuiCtx         types.TUIContext
	ActiveProject  string
	ActiveEndpoint string
	IsFocused      bool
	ActiveLanguage int // 0: cURL, 1: Fetch
	Size           tea.WindowSizeMsg
	Copied         bool
	DraftBody      string
	DraftHeaders   string
}

func NewCodeSnippetsModel(ctx types.TUIContext) *CodeSnippetsModel {
	return &CodeSnippetsModel{
		tuiCtx: ctx,
	}
}

func (m *CodeSnippetsModel) Init(ctx types.TUIContext) tea.Cmd {
	m.tuiCtx = ctx
	return nil
}

func (m *CodeSnippetsModel) Update(msg tea.Msg) (types.Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Size = msg
	case resetCopyNoticeMsg:
		m.Copied = false
		return m, nil
	case copyResultMsg:
		if msg.err != nil {
			m.Copied = false
			return m, func() tea.Msg {
				return types.ShowToastMsg{Message: "Failed to copy: " + msg.err.Error(), IsError: true}
			}
		}
		m.Copied = true
		toastCmd := func() tea.Msg {
			return types.ShowToastMsg{Message: "Snippet copied to clipboard!", IsError: false}
		}
		resetCmd := tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return resetCopyNoticeMsg{}
		})
		return m, tea.Batch(toastCmd, resetCmd)
	case types.TesterDraftChangedMsg:
		m.DraftBody = msg.Body
		m.DraftHeaders = msg.Headers
		return m, nil
	case types.EndpointSelectedMsg:
		m.ActiveProject = msg.ProjectName
		m.ActiveEndpoint = msg.EndpointPath
		m.DraftBody = ""
		m.DraftHeaders = ""
		m.Copied = false
		return m, nil
	case types.ProjectSelectedMsg:
		if msg.ProjectName != "" {
			m.ActiveProject = msg.ProjectName
			m.Copied = false
		}
		return m, nil
	}

	if !m.IsFocused || m.ActiveEndpoint == "" {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, types.GlobalKeys.Copy) || msg.String() == "y" || msg.String() == "enter":
			code := m.GenerateSnippet()
			return m, func() tea.Msg {
				err := shared.WriteToClipboard(code)
				return copyResultMsg{err: err}
			}

		case msg.String() == "left" || msg.String() == "h" || key.Matches(msg, types.GlobalKeys.Up) || msg.String() == "shift+tab":
			m.ActiveLanguage = (m.ActiveLanguage - 1 + 2) % 2
			m.Copied = false
			return m, nil

		case msg.String() == "right" || msg.String() == "l" || key.Matches(msg, types.GlobalKeys.Down) || msg.String() == "tab":
			m.ActiveLanguage = (m.ActiveLanguage + 1) % 2
			m.Copied = false
			return m, nil
		}
	}
	return m, nil
}

func (m *CodeSnippetsModel) GenerateSnippet() string {
	parts := strings.SplitN(strings.TrimSpace(m.ActiveEndpoint), " ", 2)
	method := "GET"
	path := "/"
	if len(parts) == 2 {
		method = strings.ToUpper(strings.TrimSpace(parts[0]))
		path = strings.TrimSpace(parts[1])
	} else if len(parts) == 1 && parts[0] != "" {
		path = strings.TrimSpace(parts[0])
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	path = pathParamRegex.ReplaceAllString(path, "1")

	fullPath := path
	if m.ActiveProject != "" {
		proj := strings.Trim(m.ActiveProject, "/")
		fullPath = "/" + proj + path
	}

	port := m.tuiCtx.ActivePort
	if port <= 0 {
		port = 8080
	}
	host := fmt.Sprintf("http://localhost:%d", port)

	bodyPayload := "{}"
	if strings.TrimSpace(m.DraftBody) != "" {
		bodyPayload = strings.TrimSpace(m.DraftBody)
	}

	if m.ActiveLanguage == 0 {
		if method == "POST" || method == "PUT" || method == "PATCH" {
			return fmt.Sprintf("curl -X %s \\\n  %s%s \\\n  -H 'Content-Type: application/json' \\\n  -H 'Accept: application/json' \\\n  -d '%s'", method, host, fullPath, bodyPayload)
		}
		return fmt.Sprintf("curl -X %s \\\n  %s%s \\\n  -H 'Accept: application/json'", method, host, fullPath)
	}

	if method == "POST" || method == "PUT" || method == "PATCH" {
		return fmt.Sprintf("fetch('%s%s', {\n  method: '%s',\n  headers: {\n    'Content-Type': 'application/json',\n    'Accept': 'application/json'\n  },\n  body: JSON.stringify(%s)\n})\n  .then(res => res.json())\n  .then(data => console.log(data))", host, fullPath, method, bodyPayload)
	}
	return fmt.Sprintf("fetch('%s%s', {\n  method: '%s',\n  headers: {\n    'Accept': 'application/json'\n  }\n})\n  .then(res => res.json())\n  .then(data => console.log(data))", host, fullPath, method)
}

func (m *CodeSnippetsModel) View() string {
	if m.ActiveEndpoint == "" {
		return lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Padding(1, 2).Render("Select an endpoint to generate code snippets.")
	}

	style := lipgloss.NewStyle().Padding(1, 2)
	activeStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Bold(true)
	inactiveStyle := lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim)

	curlTab := " [ cURL ] "
	fetchTab := " [ Fetch ] "

	if m.ActiveLanguage == 0 {
		curlTab = activeStyle.Render(curlTab)
		fetchTab = inactiveStyle.Render(fetchTab)
	} else {
		curlTab = inactiveStyle.Render(curlTab)
		fetchTab = activeStyle.Render(fetchTab)
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, curlTab, fetchTab)

	code := m.GenerateSnippet()

	codeWidth := m.Size.Width - 6
	if codeWidth < 0 {
		codeWidth = 0
	}

	borderColor := types.DefaultTheme.Dim
	if m.IsFocused {
		borderColor = types.DefaultTheme.Primary
	}

	codeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1).
		Width(codeWidth).
		Foreground(types.DefaultTheme.Text).
		Render(code)

	var footer string
	if m.Copied {
		footer = lipgloss.NewStyle().Foreground(types.DefaultTheme.Success).Bold(true).Render("✓ Copied to clipboard!")
	}

	elements := []string{tabs, "", codeBox}
	if footer != "" {
		elements = append(elements, "", footer)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, elements...)

	return style.Render(content)
}

func (m *CodeSnippetsModel) SetFocus(focus bool) {
	m.IsFocused = focus
	if !focus {
		m.Copied = false
	}
}
