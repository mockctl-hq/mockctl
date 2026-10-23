package types_test

import (
	"testing"

	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
)

func TestApplyTheme_Palettes(t *testing.T) {
	// Reset to default when test ends
	defer types.ApplyTheme(0)

	// Test Default Theme
	types.ApplyTheme(0)
	assert.Equal(t, "Default (Adaptive)", types.DefaultTheme.Name)
	assert.Equal(t, types.ThemeDefault.Primary, types.DefaultTheme.Primary)

	// Test Ocean Dark Theme
	types.ApplyTheme(1)
	assert.Equal(t, "Ocean Dark", types.DefaultTheme.Name)
	assert.Equal(t, types.ThemeOceanDark.Primary, types.DefaultTheme.Primary)

	// Test Dracula Theme
	types.ApplyTheme(2)
	assert.Equal(t, "Dracula", types.DefaultTheme.Name)
	assert.Equal(t, types.ThemeDracula.Primary, types.DefaultTheme.Primary)

	// Test Monokai Theme
	types.ApplyTheme(3)
	assert.Equal(t, "Monokai", types.DefaultTheme.Name)
	assert.Equal(t, types.ThemeMonokai.Primary, types.DefaultTheme.Primary)

	// Test Out of bounds fallback
	types.ApplyTheme(999)
	assert.Equal(t, "Default (Adaptive)", types.DefaultTheme.Name)

	types.ApplyTheme(-1)
	assert.Equal(t, "Default (Adaptive)", types.DefaultTheme.Name)
}

func TestApplyTheme_RecompilesStyles(t *testing.T) {
	defer types.ApplyTheme(0)

	// Apply Dracula
	types.ApplyTheme(2)
	assert.Equal(t, "Dracula", types.DefaultTheme.Name)

	// GlobalStyles should be compiled and non-empty
	borderView := types.GlobalStyles.FocusedBorder.Render("test")
	assert.Contains(t, borderView, "test")

	unfocusedView := types.GlobalStyles.UnfocusedBorder.Render("unfocused")
	assert.Contains(t, unfocusedView, "unfocused")
}
