package overlays

import (
	"os"
	"path/filepath"
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
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return nil
		}
		cleanHome := filepath.Clean(home)
		homeSepCount := strings.Count(cleanHome, string(filepath.Separator))

		var files []string
		_ = filepath.WalkDir(cleanHome, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			depth := strings.Count(path, string(filepath.Separator)) - homeSepCount
			if depth > 5 {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") && path != cleanHome {
				return filepath.SkipDir
			}
			if !d.IsDir() {
				ext := strings.ToLower(filepath.Ext(d.Name()))
				if ext == ".yaml" || ext == ".yml" {
					files = append(files, path)
				}
			}
			return nil
		})

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
