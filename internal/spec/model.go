package spec

import "context"

// SpecModel is the strictly read-only schema tree.
type SpecModel struct {
	Title   string
	Version string
	Routes  []RouteDef
}

type RouteDef struct {
	Method      string
	Path        string
	Status      int
	ContentType string
	SchemaRef   any // Normalized schema representation
}

// NormalizedSchema translates kin-openapi AST into safe Go structs
type NormalizedSchema struct {
	Type       string
	Format     string
	Properties map[string]*NormalizedSchema
	Items      *NormalizedSchema
	Required   []string
	Enum       []any
	Example    any // Priority 1 for Data Generation
	Default    any // Priority 2 for Data Generation
	AllOf      []*NormalizedSchema
	OneOf      []*NormalizedSchema
	AnyOf      []*NormalizedSchema
}

type SpecParser interface {
	ParseFile(ctx context.Context, path string) (*SpecModel, error)
}

type RouteExtractor interface {
	// Iterates over the parsed model to yield clean route definitions
	ExtractRoutes(model *SpecModel) ([]RouteDef, error)
}
