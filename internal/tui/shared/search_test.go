package shared_test

import (
	"testing"

	"github.com/mockctl-hq/mockctl/internal/tui/shared"
	"github.com/stretchr/testify/assert"
)

func TestFilterSubstring(t *testing.T) {
	items := []string{
		"GET /api/v1/users",
		"POST /api/v1/users",
		"GET /api/v1/products",
		"DELETE /api/v1/orders",
	}

	// Empty query returns everything
	assert.Equal(t, items, shared.FilterSubstring("", items))
	assert.Equal(t, items, shared.FilterSubstring("   ", items))

	// Case-insensitive query
	users := shared.FilterSubstring("USERS", items)
	assert.Equal(t, []string{"GET /api/v1/users", "POST /api/v1/users"}, users)

	// Substring match
	products := shared.FilterSubstring("prod", items)
	assert.Equal(t, []string{"GET /api/v1/products"}, products)

	// No match returns empty slice
	none := shared.FilterSubstring("nonexistent", items)
	assert.Empty(t, none)
}

func TestFilterFuzzy(t *testing.T) {
	items := []string{
		"alpha/beta/gamma.yaml",
		"openapi/spec/v1.yml",
		"routes/handlers.go",
	}

	assert.Equal(t, items, shared.FilterFuzzy("", items))

	matched := shared.FilterFuzzy("spec", items)
	assert.Equal(t, []string{"openapi/spec/v1.yml"}, matched)

	fuzzyMatched := shared.FilterFuzzy("abg", items)
	assert.Equal(t, []string{"alpha/beta/gamma.yaml"}, fuzzyMatched)
}

func TestSearchModel_Filter(t *testing.T) {
	sm := shared.NewSearchModel()
	items := []string{"apple", "banana", "cherry"}

	sm.Input.SetValue("ap")
	assert.Equal(t, []string{"apple"}, sm.Filter(items))
	assert.Equal(t, []string{"apple"}, sm.FilterSubstring(items))
}
