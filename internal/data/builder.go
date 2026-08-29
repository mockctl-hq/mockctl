package data

import (
	"context"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/mockctl-hq/mockctl/internal/shared"
	"github.com/mockctl-hq/mockctl/internal/spec"
)

type PayloadBuilder interface {
	BuildFromSchema(ctx context.Context, schemaRef any) (any, error)
}

type FakeValueProvider struct{}

func NewFakeValueProvider() *FakeValueProvider {
	return &FakeValueProvider{}
}

func (p *FakeValueProvider) GenerateUUID(ctx context.Context) string {
	return gofakeit.UUID()
}

func (p *FakeValueProvider) GenerateString(ctx context.Context, format string) string {
	switch format {
	case "uuid":
		return gofakeit.UUID()
	case "email":
		return gofakeit.Email()
	case "date-time":
		return gofakeit.Date().Format("2006-01-02T15:04:05Z07:00")
	case "ipv4":
		return gofakeit.IPv4Address()
	case "password":
		return gofakeit.Password(true, true, true, true, false, 12)
	default:
		return gofakeit.Word()
	}
}

func (p *FakeValueProvider) GenerateInt(ctx context.Context, min, max int) int {
	if min >= max {
		return min
	}
	return gofakeit.Number(min, max)
}

func (p *FakeValueProvider) GenerateBoolean(ctx context.Context) bool {
	return gofakeit.Bool()
}

type DefaultPayloadBuilder struct {
	provider shared.ValueProvider
}

func NewDefaultPayloadBuilder(provider shared.ValueProvider) *DefaultPayloadBuilder {
	return &DefaultPayloadBuilder{provider: provider}
}

func (b *DefaultPayloadBuilder) BuildFromSchema(ctx context.Context, schemaRef any) (any, error) {
	if schemaRef == nil {
		return nil, nil // Body-less response like 204
	}

	normSchema, ok := schemaRef.(*spec.NormalizedSchema)
	if !ok {
		return nil, nil
	}

	return b.build(ctx, normSchema)
}

func (b *DefaultPayloadBuilder) build(ctx context.Context, s *spec.NormalizedSchema) (any, error) {
	if s == nil {
		return nil, nil
	}

	// Priority 1: Example
	if s.Example != nil {
		return s.Example, nil
	}

	// Priority 2: Default
	if s.Default != nil {
		return s.Default, nil
	}

	// Priority 3: Polymorphic merging (simplistic pick first)
	if len(s.OneOf) > 0 {
		return b.build(ctx, s.OneOf[0])
	}
	if len(s.AnyOf) > 0 {
		return b.build(ctx, s.AnyOf[0])
	}

	switch s.Type {
	case "string":
		return b.provider.GenerateString(ctx, s.Format), nil
	case "integer", "number":
		return b.provider.GenerateInt(ctx, 1, 1000), nil
	case "boolean":
		return b.provider.GenerateBoolean(ctx), nil
	case "array":
		if s.Items != nil {
			item, err := b.build(ctx, s.Items)
			if err != nil {
				return nil, err
			}
			return []any{item}, nil
		}
		return []any{}, nil
	case "object":
		fallthrough
	default: // Treat empty type as object usually
		if len(s.Properties) > 0 {
			obj := make(map[string]any)
			for k, v := range s.Properties {
				val, err := b.build(ctx, v)
				if err != nil {
					return nil, err
				}
				obj[k] = val
			}
			// Merge allOf properties
			for _, allOfSchema := range s.AllOf {
				if allOfSchema != nil {
					allOfVal, err := b.build(ctx, allOfSchema)
					if err == nil {
						if allOfMap, ok := allOfVal.(map[string]any); ok {
							for k, v := range allOfMap {
								obj[k] = v
							}
						}
					}
				}
			}
			return obj, nil
		}
		return map[string]any{}, nil
	}
}
