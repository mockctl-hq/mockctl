package http_test

import (
	"context"
	"testing"

	"github.com/mockctl-hq/mockctl/internal/app"
	"github.com/mockctl-hq/mockctl/internal/tui/components/workspace/tester/protocols/http"
	"github.com/mockctl-hq/mockctl/internal/tui/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpTesterModel_DraftsPersistenceAcrossEndpoints(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	ctx := context.Background()
	core, err := app.NewApp(ctx)
	require.NoError(t, err)
	defer core.Shutdown(ctx)

	tuiCtx := types.TUIContext{Ctx: ctx, Core: core, ActivePort: 8080}
	tester := http.NewHttpTesterModel(tuiCtx)
	_ = tester.Init(tuiCtx)

	// 1. Select first endpoint
	tester.Update(types.EndpointSelectedMsg{
		ProjectName:  "my-proj",
		EndpointPath: "POST /users",
	})
	assert.Equal(t, "POST", tester.Input.Method)
	assert.Equal(t, "/my-proj/users", tester.Input.PathInput.Value())

	// 2. Modify body & headers for this draft
	tester.Body.Input.SetValue(`{"name":"Alice"}`)
	tester.Headers.Input.SetValue("Authorization: Bearer token123")

	// 3. Switch to another endpoint
	tester.Update(types.EndpointSelectedMsg{
		ProjectName:  "my-proj",
		EndpointPath: "GET /products",
	})
	assert.Equal(t, "GET", tester.Input.Method)
	assert.Equal(t, "/my-proj/products", tester.Input.PathInput.Value())
	assert.Equal(t, "", tester.Body.Input.Value(), "New endpoint draft should start empty")

	// 4. Switch back to first endpoint -> verify draft was preserved!
	tester.Update(types.EndpointSelectedMsg{
		ProjectName:  "my-proj",
		EndpointPath: "POST /users",
	})
	assert.Equal(t, "POST", tester.Input.Method)
	assert.Equal(t, "/my-proj/users", tester.Input.PathInput.Value())
	assert.Equal(t, `{"name":"Alice"}`, tester.Body.Input.Value(), "Draft body must be preserved")
	assert.Equal(t, "Authorization: Bearer token123", tester.Headers.Input.Value(), "Draft headers must be preserved")
}
