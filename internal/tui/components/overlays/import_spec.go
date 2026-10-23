package overlays

import (
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
)

type ImportSpecModel struct {
	tuiCtx types.TUIContext
	Size   tea.WindowSizeMsg

	Active      bool
	Input       textinput.Model
	ProjectName string
	ErrorMsg    string

	// Search state
	IsLoading   bool
	AllFiles    []string
	Matches     []string
	CursorIndex int
	SearchMode  bool

	loadOnce sync.Once
}

func NewImportSpecModel(ctx types.TUIContext) *ImportSpecModel {
	ti := textinput.New()
	ti.Placeholder = "Type path or search query..."
	ti.Focus()
	ti.CharLimit = 256
	ti.PromptStyle = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Background(types.DefaultTheme.Bg)
	ti.TextStyle = lipgloss.NewStyle().Foreground(types.DefaultTheme.Text).Background(types.DefaultTheme.Bg)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(types.DefaultTheme.Dim).Background(types.DefaultTheme.Bg)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(types.DefaultTheme.Primary).Background(types.DefaultTheme.Bg)

	m := &ImportSpecModel{
		tuiCtx: ctx,
		Input:  ti,
	}

	return m
}

func (m *ImportSpecModel) Init(ctx types.TUIContext) tea.Cmd {
	return tea.Batch(textinput.Blink, m.loadFilesCmd())
}

func (m *ImportSpecModel) loadFilesCmd() tea.Cmd {
	return func() tea.Msg {
		home, _ := os.UserHomeDir()
		if home == "" {
			return nil
		}
		// Fast find for yaml files up to 5 levels deep
		cmd := exec.Command("find", home, "-maxdepth", "5", "-type", "f", "-name", "*.yaml", "-o", "-name", "*.yml")
		out, err := cmd.Output()
		if err != nil {
			return nil
		}

		lines := strings.Split(string(out), "\n")
		var files []string
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				files = append(files, l)
			}
		}
		return FilesLoadedMsg{Files: files}
	}
}

type FilesLoadedMsg struct {
	Files []string
}

func (m *ImportSpecModel) SetProject(name string) {
	m.ProjectName = name
	m.Input.SetValue("")
	m.ErrorMsg = ""
	m.SearchMode = false
	m.Matches = nil
	m.CursorIndex = 0
}

func (m *ImportSpecModel) SetFocus(focus bool) {
	m.Active = focus
	if focus {
		m.Input.Focus()
	} else {
		m.Input.Blur()
	}
}
