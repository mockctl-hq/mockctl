package types

import (
	"github.com/mockctl-hq/mockctl/internal/runtime"
)

// ShowToastMsg is emitted to trigger a global transient notification in the header.
type ShowToastMsg struct {
	Message string
	IsError bool
}

// DaemonCrashedMsg is emitted when the background daemon fails to bind or crashes.
type DaemonCrashedMsg struct {
	Error string
}

// ProjectStructureChangedMsg is emitted when a project's spec changes (e.g., endpoint added/renamed),
// passing the updated DTO to the Tree Component.
type ProjectStructureChangedMsg struct {
	ProjectName string
	// We'll pass the parsed SpecModel DTO here eventually.
}

// ProjectSelectedMsg is emitted when a project node is clicked in the Project Tree.
type ProjectSelectedMsg struct {
	ProjectName string
	FileName    string
}

// EndpointSelectedMsg is emitted when an endpoint is clicked in the Project Tree.
type EndpointSelectedMsg struct {
	ProjectName  string
	EndpointPath string // e.g., "GET /users"
}

// TesterToggleMsg is emitted to expand/collapse the bottom API Tester panel.
type TesterToggleMsg struct{}

// SaveYamlMsg is emitted when the user presses Ctrl+S.
type SaveYamlMsg struct{}

// FileExternallyModifiedMsg is emitted when the app detects an external disk modification.
type FileExternallyModifiedMsg struct{}

// BlinkMsg is emitted every 500ms to toggle the colon visibility (Daemon Heartbeat).
type BlinkMsg struct{}

// TelemetryMsg is emitted when the backend EventBroker pushes a new real-time event.
type TelemetryMsg struct {
	Payload runtime.Event
}

// SettingsChangedMsg is emitted whenever the user saves a settings change.
// Components that depend on settings (e.g., HTTP Client, History) must listen for this
// and update their internal state accordingly.
type SettingsChangedMsg struct {
	Config *SettingsConfig
}

// TesterDraftChangedMsg is emitted when the user modifies body or headers in the API Tester,
// enabling reactive live code snippet generation.
type TesterDraftChangedMsg struct {
	Method  string
	Path    string
	Body    string
	Headers string
}
