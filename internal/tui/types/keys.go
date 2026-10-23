package types

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the global keybindings for the TUI.
type KeyMap struct {
	Quit            key.Binding
	QuitContext     key.Binding
	TabForward      key.Binding
	TabBackward     key.Binding
	Search          key.Binding
	AutoTail        key.Binding
	Format          key.Binding
	CollapseTree    key.Binding
	ToggleLeft      key.Binding
	ToggleRight     key.Binding
	ToggleApiTester key.Binding
	Save            key.Binding
	NextTab         key.Binding
	PrevTab         key.Binding
	CloseTab        key.Binding
	ImportSpec      key.Binding
	ExecuteTest     key.Binding
	ResetState      key.Binding
	Escape          key.Binding
	Copy            key.Binding
	Delete          key.Binding
	Restart         key.Binding
	ToggleHelp      key.Binding
	Up              key.Binding
	Down            key.Binding
	Settings        key.Binding
	History         key.Binding
	GlobalLog       key.Binding
	Profiler        key.Binding
	StateInspector  key.Binding
}

// GlobalKeys is the instantiated map of all keyboard shortcuts.
var GlobalKeys = KeyMap{
	Quit:            key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "quit")),
	QuitContext:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit / copy")),
	TabForward:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
	TabBackward:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
	Search:          key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
	AutoTail:        key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "toggle auto-tail")),
	Format:          key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "format code")),
	CollapseTree:    key.NewBinding(key.WithKeys("shift+left"), key.WithHelp("shift+left", "collapse tree")),
	ToggleLeft:      key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "toggle left panel")),
	ToggleRight:     key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "toggle right panel")),
	ToggleApiTester: key.NewBinding(key.WithKeys("ctrl+\\"), key.WithHelp("ctrl+\\", "toggle api tester")),
	Save:            key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save changes")),
	NextTab:         key.NewBinding(key.WithKeys("ctrl+right", "alt+right"), key.WithHelp("ctrl+right", "next tab")),
	PrevTab:         key.NewBinding(key.WithKeys("ctrl+left", "alt+left"), key.WithHelp("ctrl+left", "prev tab")),
	CloseTab:        key.NewBinding(key.WithKeys("ctrl+w"), key.WithHelp("ctrl+w", "close tab")),
	ImportSpec:      key.NewBinding(key.WithKeys("alt+o"), key.WithHelp("alt+o", "import spec")),
	ExecuteTest:     key.NewBinding(key.WithKeys("ctrl+r", "ctrl+enter"), key.WithHelp("ctrl+r", "execute test")),
	ResetState:      key.NewBinding(key.WithKeys("ctrl+backspace"), key.WithHelp("ctrl+bksp", "reset memory state")),
	Escape:          key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel / back")),
	Copy:            key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy to clipboard")),
	Delete:          key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete project")),
	Restart:         key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "restart daemon")),
	ToggleHelp:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	Up:              key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:            key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Settings:        key.NewBinding(key.WithKeys("alt+s"), key.WithHelp("alt+s", "settings")),
	History:         key.NewBinding(key.WithKeys("alt+h"), key.WithHelp("alt+h", "history")),
	GlobalLog:       key.NewBinding(key.WithKeys("alt+l"), key.WithHelp("alt+l", "logs")),
	Profiler:        key.NewBinding(key.WithKeys("alt+p"), key.WithHelp("alt+p", "profiler")),
	StateInspector:  key.NewBinding(key.WithKeys("alt+x", "alt+i"), key.WithHelp("alt+x/alt+i", "inspector")),
}
