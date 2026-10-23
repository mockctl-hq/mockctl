package types_test

import (
	"testing"
	"time"

	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsConfig_SerializationRoundTrip(t *testing.T) {
	orig := &types.SettingsConfig{
		ActiveTheme:             2, // Dracula
		UIAnimations:            false,
		GlobalCORS:              false,
		StrictOpenAPIValidation: false,
		RequestTimeout:          60 * time.Second,
		HistoryLimit:            5000,
		TabSize:                 4,
		LineWrapping:            true,
		AutoFormatOnSave:        false,
		FollowRedirects:         true,
	}

	data, err := orig.Serialize()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	loaded, err := types.DeserializeSettingsConfig(data)
	require.NoError(t, err)
	assert.Equal(t, orig, loaded)
}

func TestSettingsConfig_DeserializeFallback(t *testing.T) {
	// Empty data should return defaults
	loaded, err := types.DeserializeSettingsConfig(nil)
	require.NoError(t, err)
	assert.Equal(t, types.DefaultSettingsConfig(), loaded)

	// Partial JSON should populate specified fields and keep defaults for others
	partialJSON := []byte(`{"active_theme": 1, "history_limit": 500}`)
	loaded, err = types.DeserializeSettingsConfig(partialJSON)
	require.NoError(t, err)
	assert.Equal(t, 1, loaded.ActiveTheme)
	assert.Equal(t, 500, loaded.HistoryLimit)
	assert.Equal(t, true, loaded.UIAnimations, "Should retain default UIAnimations")
	assert.Equal(t, 10*time.Second, loaded.RequestTimeout, "Should retain default RequestTimeout")

	// Corrupted JSON should return default and error
	loaded, err = types.DeserializeSettingsConfig([]byte(`{invalid-json`))
	assert.Error(t, err)
	assert.Equal(t, types.DefaultSettingsConfig(), loaded)
}
