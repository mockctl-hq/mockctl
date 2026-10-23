package types

import (
	"encoding/json"
	"time"
)

// SettingsConfig is the single source of truth for all user-configurable TUI and runtime settings.
// It is held as a pointer in TUIContext so any component can read live values,
// and SettingsChangedMsg broadcasts updates to components that need to react.
type SettingsConfig struct {
	// --- Theme ---
	// ActiveTheme: 0=Default (Adaptive), 1=Ocean Dark, 2=Dracula, 3=Monokai
	ActiveTheme int `json:"active_theme"`

	// UIAnimations: when false, components skip background animated tickers (e.g., logo colon blink).
	// Applied immediately when settings change or on launch.
	UIAnimations bool `json:"ui_animations"`

	// --- Network ---
	// GlobalCORS: whether to add CORS headers to all mock responses.
	GlobalCORS bool `json:"global_cors"`

	// StrictOpenAPIValidation: when true, the engine validates request bodies against the spec.
	StrictOpenAPIValidation bool `json:"strict_openapi_validation"`

	// RequestTimeout is the HTTP timeout used by the API Tester's http.Client.
	// Applied immediately when settings change.
	RequestTimeout time.Duration `json:"request_timeout"`

	// HistoryLimit is the max number of telemetry events fetched for the History overlay.
	// Applied immediately when settings change.
	HistoryLimit int `json:"history_limit"`

	// --- Editor ---
	// TabSize is the number of spaces per indent level in the YAML editor.
	TabSize int `json:"tab_size"`

	// LineWrapping: when true, long YAML lines wrap instead of scrolling horizontally.
	LineWrapping bool `json:"line_wrapping"`

	// AutoFormatOnSave: when true, YAML is formatted before saving.
	AutoFormatOnSave bool `json:"auto_format_on_save"`

	// --- Tester ---
	// FollowRedirects: when true, the API Tester http.Client follows HTTP redirects.
	// Applied immediately when settings change.
	FollowRedirects bool `json:"follow_redirects"`
}

// DefaultSettingsConfig returns the default configuration matching the previous hardcoded values.
func DefaultSettingsConfig() *SettingsConfig {
	return &SettingsConfig{
		ActiveTheme:             0,
		UIAnimations:            true,
		GlobalCORS:              true,
		StrictOpenAPIValidation: true,
		RequestTimeout:          10 * time.Second,
		HistoryLimit:            1000,
		TabSize:                 2,
		LineWrapping:            false,
		AutoFormatOnSave:        true,
		FollowRedirects:         false,
	}
}

// Serialize converts the config into JSON bytes for persistence.
func (c *SettingsConfig) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// DeserializeSettingsConfig parses JSON bytes into a SettingsConfig, falling back to defaults for missing fields.
func DeserializeSettingsConfig(data []byte) (*SettingsConfig, error) {
	cfg := DefaultSettingsConfig()
	if len(data) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultSettingsConfig(), err
	}
	return cfg, nil
}
