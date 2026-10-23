package types

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color palette for the TUI.
type Theme struct {
	Name      string
	Primary   lipgloss.AdaptiveColor
	Secondary lipgloss.AdaptiveColor
	Success   lipgloss.AdaptiveColor
	Warning   lipgloss.AdaptiveColor
	Error     lipgloss.AdaptiveColor
	Dim       lipgloss.AdaptiveColor
	Text      lipgloss.AdaptiveColor
	Bg        lipgloss.AdaptiveColor
}

// Built-in theme palettes
var (
	ThemeDefault = Theme{
		Name:      "Default (Adaptive)",
		Primary:   lipgloss.AdaptiveColor{Light: "#fe6802", Dark: "#fe6802"},
		Secondary: lipgloss.AdaptiveColor{Light: "#5F00AF", Dark: "#AF87FF"},
		Success:   lipgloss.AdaptiveColor{Light: "#008700", Dark: "#5FFF87"},
		Warning:   lipgloss.AdaptiveColor{Light: "#AF5F00", Dark: "#FFD700"},
		Error:     lipgloss.AdaptiveColor{Light: "#D70000", Dark: "#FF5F5F"},
		Dim:       lipgloss.AdaptiveColor{Light: "#808080", Dark: "#585858"},
		Text:      lipgloss.AdaptiveColor{Light: "#1C1C1C", Dark: "#EEEEEE"},
		Bg:        lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#121212"},
	}

	ThemeOceanDark = Theme{
		Name:      "Ocean Dark",
		Primary:   lipgloss.AdaptiveColor{Light: "#0077b6", Dark: "#00b4d8"},
		Secondary: lipgloss.AdaptiveColor{Light: "#023e8a", Dark: "#90e0ef"},
		Success:   lipgloss.AdaptiveColor{Light: "#2d6a4f", Dark: "#52b788"},
		Warning:   lipgloss.AdaptiveColor{Light: "#d4a373", Dark: "#f4a261"},
		Error:     lipgloss.AdaptiveColor{Light: "#b00020", Dark: "#e76f51"},
		Dim:       lipgloss.AdaptiveColor{Light: "#718096", Dark: "#4a5568"},
		Text:      lipgloss.AdaptiveColor{Light: "#0f172a", Dark: "#e2e8f0"},
		Bg:        lipgloss.AdaptiveColor{Light: "#e0f2fe", Dark: "#0a1128"},
	}

	ThemeDracula = Theme{
		Name:      "Dracula",
		Primary:   lipgloss.AdaptiveColor{Light: "#7c3aed", Dark: "#bd93f9"},
		Secondary: lipgloss.AdaptiveColor{Light: "#db2777", Dark: "#ff79c6"},
		Success:   lipgloss.AdaptiveColor{Light: "#16a34a", Dark: "#50fa7b"},
		Warning:   lipgloss.AdaptiveColor{Light: "#ca8a04", Dark: "#f1fa8c"},
		Error:     lipgloss.AdaptiveColor{Light: "#dc2626", Dark: "#ff5555"},
		Dim:       lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#6272a4"},
		Text:      lipgloss.AdaptiveColor{Light: "#282a36", Dark: "#f8f8f2"},
		Bg:        lipgloss.AdaptiveColor{Light: "#f8f8f2", Dark: "#282a36"},
	}

	ThemeMonokai = Theme{
		Name:      "Monokai",
		Primary:   lipgloss.AdaptiveColor{Light: "#4d7c0f", Dark: "#a6e22e"},
		Secondary: lipgloss.AdaptiveColor{Light: "#0284c7", Dark: "#66d9ef"},
		Success:   lipgloss.AdaptiveColor{Light: "#15803d", Dark: "#a6e22e"},
		Warning:   lipgloss.AdaptiveColor{Light: "#c2410c", Dark: "#fd971f"},
		Error:     lipgloss.AdaptiveColor{Light: "#be123c", Dark: "#f92672"},
		Dim:       lipgloss.AdaptiveColor{Light: "#6b7280", Dark: "#75715e"},
		Text:      lipgloss.AdaptiveColor{Light: "#272822", Dark: "#f8f8f2"},
		Bg:        lipgloss.AdaptiveColor{Light: "#f8f8f2", Dark: "#272822"},
	}
)

// DefaultTheme provides the currently active color palette across the TUI.
var DefaultTheme = ThemeDefault

// ThemeList contains the list of all available themes.
var ThemeList = []Theme{
	ThemeDefault,
	ThemeOceanDark,
	ThemeDracula,
	ThemeMonokai,
}

// ApplyTheme activates the specified theme index and recompiles global styles.
func ApplyTheme(themeIndex int) {
	if themeIndex >= 0 && themeIndex < len(ThemeList) {
		DefaultTheme = ThemeList[themeIndex]
	} else {
		DefaultTheme = ThemeDefault
	}
	InitTheme()
}

// Styles holds pre-compiled lipgloss styles for performance.
type Styles struct {
	FocusedBorder   lipgloss.Style
	UnfocusedBorder lipgloss.Style
	Base            lipgloss.Style
}

// GlobalStyles is the instantiated map of styles.
var GlobalStyles = Styles{}

// InitTheme initializes the global styles, taking NO_COLOR into account.
func InitTheme() {
	// PKS-008: Check NO_COLOR to disable styling if requested.
	if os.Getenv("NO_COLOR") != "" {
		// Monochrome mode
		GlobalStyles.FocusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
		GlobalStyles.UnfocusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
		GlobalStyles.Base = lipgloss.NewStyle()
		return
	}

	// Light Mode Protection
	lipgloss.SetHasDarkBackground(lipgloss.HasDarkBackground())

	GlobalStyles.FocusedBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DefaultTheme.Primary)

	GlobalStyles.UnfocusedBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(DefaultTheme.Dim)

	GlobalStyles.Base = lipgloss.NewStyle().
		Foreground(DefaultTheme.Text)
}
