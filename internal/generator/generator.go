package generator

import (
	"context"
	"fmt"

	"github.com/mockctl-hq/mockctl/internal/data"
	"github.com/mockctl-hq/mockctl/internal/project"
	"github.com/mockctl-hq/mockctl/internal/shared"
	"github.com/mockctl-hq/mockctl/internal/spec"
)

type MockGenerator struct {
	logger   shared.Logger
	provider shared.ValueProvider
}

func NewMockGenerator(l shared.Logger, p shared.ValueProvider) *MockGenerator {
	return &MockGenerator{logger: l, provider: p}
}

type OverrideMerger interface {
	MergeOverrides(def *RuntimeDefinition, ctx *project.WorkspaceContext) error
}

func (g *MockGenerator) Generate(ctx context.Context, model *spec.SpecModel, wsCtx *project.WorkspaceContext) (*RuntimeDefinition, error) {
	def := &RuntimeDefinition{
		Endpoints: make(map[string]EndpointHandler),
	}

	builder := data.NewDefaultPayloadBuilder(g.provider)

	for _, route := range model.Routes {
		id := route.Method + " " + route.Path
		handler, exists := def.Endpoints[id]
		if !exists {
			handler = EndpointHandler{
				Method:    route.Method,
				Path:      route.Path,
				Responses: make(map[int]ResponseTemplate),
			}
		}

		// Apply override if it exists
		var body any
		overridden := false
		if wsCtx != nil && wsCtx.Overrides != nil {
			if pathOverrides, ok := wsCtx.Overrides[route.Path]; ok {
				if overrideBody, ok := pathOverrides[route.Status]; ok {
					body = overrideBody
					overridden = true
				}
			}
		}

		if !overridden {
			if route.SchemaRef != nil {
				builtBody, err := builder.BuildFromSchema(ctx, route.SchemaRef)
				if err != nil {
					// We just log and continue, we don't want to crash the whole generation
					g.logger.Error(fmt.Sprintf("Failed to build payload for path %s", route.Path), err)
					continue
				}
				body = builtBody
			}
		}

		handler.Responses[route.Status] = ResponseTemplate{
			Headers: map[string]string{"Content-Type": route.ContentType},
			Body:    body,
		}

		def.Endpoints[id] = handler
	}

	return def, nil
}
