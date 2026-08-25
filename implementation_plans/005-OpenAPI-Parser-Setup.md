# Implementation Plan 005: OpenAPI Schema Parser & Mock Generator

**Status:** 🟢 Approved  
**Focus:** Implementing the `kin-openapi` parsing layer (`internal/spec`) and the `MockGenerator` (`internal/generator`) strictly according to **PKS-024**, **PKS-025**, and high-level security standards.

---

## 1. Objective

To build the parsing and generation engine that reads a user's `openapi.yaml` and produces the `RuntimeDefinition` used by the `RuntimeEngine`. This implementation strictly enforces the **Anti-Corruption Layer (PKS-024)**, ensuring `kin-openapi` types never leak outside `internal/spec/`.

---

## 2. Architectural & Security Constraints (PKS-024, PKS-025, PKS-028)
1. **Strict Isolation:** `kin-openapi` MUST only be imported inside `internal/spec/`. No other package is allowed to import it.
2. **Version Lock (Critical):** The implementation MUST use `github.com/getkin/kin-openapi@v0.9.0` (or compatible) because versions `>v0.9.0` require Go 1.25+, which breaks the project's Go 1.24.0 compiler requirement.
3. **Read-Only Model:** The `SpecModel` is a strictly read-only schema tree.
4. **Stateless Components:** Components like `SpecParser`, `MockGenerator`, and `DataGeneration` must be completely stateless after initialization.
5. **No Runtime Validation Coupling:** The runtime engine must remain parser-agnostic. All validation logic must rely purely on the normalized internal definitions.
6. **🛡️ SSRF & LFI Prevention:** The parser must strictly disable remote `$ref` fetching (HTTP/HTTPS) and ensure local file `$ref` resolutions cannot escape the Workspace Directory (No Path Traversal).
7. **🛡️ Infinite Recursion Protection:** The `SpecParser` must implement cycle detection or a maximum recursion depth (e.g., 10 levels) when normalizing schemas to prevent Stack Overflow / OOM crashes caused by circular `$ref`s.
8. **🛡️ YAML Bomb Protection:** Enforce a strict file size limit (e.g., 50MB) before loading the specification into memory to prevent Memory Exhaustion DoS.
9. **🛡️ ReDoS Protection:** The `ValueProvider` must implement timeouts or safe-limits when processing OpenAPI `pattern` (regex) fields for data generation to prevent CPU hanging on malicious regular expressions.

---

## 3. Step-by-Step Implementation

### Step 1: The Anti-Corruption Parsing Layer (`internal/spec`)
**Files:** `internal/spec/model.go` & `internal/spec/parser.go`

- **Purpose:** To read an OpenAPI v3 file securely, applying size limits, safe `$ref` resolution, extracting both global and operation-level base paths, and mapping it to the internal `SpecModel`.
- **Model Definition (With Polymorphic & Example Support):**
  ```go
  type SpecModel struct {
      Title   string
      Version string
      Routes  []RouteDef
  }

  type RouteDef struct {
      Method      string
      Path        string // Includes global OR operation-level base path prepended
      Status      int
      ContentType string
      SchemaRef   any // Normalized schema representation (e.g. NormalizedSchema)
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
      AllOf      []*NormalizedSchema // Polymorphic schemas
      OneOf      []*NormalizedSchema
      AnyOf      []*NormalizedSchema
  }
  ```
- **Interfaces:**
  ```go
  type SpecParser interface {
      ParseFile(ctx context.Context, path string) (*SpecModel, error)
  }
  
  type RouteExtractor interface {
      // Iterates over the parsed model to yield clean route definitions, accounting for operation-level `servers`
      ExtractRoutes(model *SpecModel) ([]RouteDef, error)
  }
  ```
- **Implementation Details:**
  - Enforce `v0.9.0` in `go.mod`.
  - Disable external URL resolution in `openapi3.Loader`.
  - Check file size before loading.
  - When mapping schemas, explicitly extract `Example`, `Default`, and polymorphic slices (`AllOf`, `OneOf`, `AnyOf`) into `NormalizedSchema`.
  - Resolve global `servers` and override with path/operation-level `servers` if present.

### Step 2: The Data Generation Design (`internal/data`)
**File:** `internal/data/builder.go`

- **Purpose:** To generate realistic fake payloads deterministically using `gofakeit`, prioritizing user-defined examples.
- **Interfaces:**
  ```go
  type ValueProvider interface {
      GenerateString(format string) string // MUST map to gofakeit (e.g. format="uuid" -> gofakeit.UUID())
      GenerateInt(min, max int) int
      GenerateBoolean() bool
  }

  type PayloadBuilder interface {
      BuildFromSchema(ctx context.Context, schemaRef any) (any, error)
  }
  ```
- **Implementation Details:**
  - **Priority Logic:** The builder MUST check for the presence of `Example`, then `Default` inside the `NormalizedSchema`. Only if both are `nil` should it use `ValueProvider` to generate a random fake value.
  - **Semantic Format Mapping:** `GenerateString` MUST strictly map OpenAPI `format` strings (`uuid`, `date-time`, `email`, `ipv4`) to their respective `gofakeit` functions to ensure frontend validation passes.
  - **Nil Schema Handling:** Gracefully return empty maps/nil for body-less schemas (e.g., 204 No Content).
  - **Polymorphic Merging:** The builder must merge properties from `AllOf`, and pick the first valid option from `OneOf` or `AnyOf`.

### Step 3: The Mock Generation Engine (`internal/generator`)
**File:** `internal/generator/generator.go`

- **Purpose:** Combines `SpecModel`, custom overrides, and `ValueProvider` into a runtime blueprint.
- **Implementation:**
  ```go
  type MockGenerator struct {
      logger   shared.Logger
      provider data.ValueProvider
  }

  func NewMockGenerator(l shared.Logger, p data.ValueProvider) *MockGenerator {
      return &MockGenerator{logger: l, provider: p}
  }

  type OverrideMerger interface {
      // Applies user-defined custom payloads from WorkspaceContext
      MergeOverrides(def *RuntimeDefinition, ctx *project.WorkspaceContext) error
  }
  ```
- **Logic:**
  1. `Generate(ctx, model, wsCtx)` iterates through `model.Routes`.
  2. If a route has a nil schema (e.g. 204 No Content), assign an empty template safely.
  3. Uses `data.PayloadBuilder` to build default payloads.
  4. Uses `OverrideMerger` to apply static JSON replacements from `WorkspaceContext`.
  5. Returns the executable `RuntimeDefinition`.

### Step 4: Unit Testing Strategy
**Files:** `internal/spec/parser_test.go` & `internal/generator/generator_test.go`

- **Parser Tests:** 
  - Ensure OpenAPI `example`, `default`, and polymorphic fields map correctly to `NormalizedSchema`.
  - Verify that operation-level `servers` override global `servers` in the extracted `Path`.
  - Security assertions (100MB dummy file rejection, remote `$ref` rejection, infinite loop panic prevention).
- **Generator Tests:** 
  - Verify `PayloadBuilder` returns semantic data (e.g., valid UUIDs for `format: uuid`) via `gofakeit`.
  - Verify `PayloadBuilder` prioritizes the exact string from `Example` over random generation.
  - Verify 204 No Content routes process safely without Nil Pointer Panics.

---

## 4. Directory & Package Blueprint

```text
internal/
├── spec/
│   ├── model.go             // Read-only structs (SpecModel, RouteDef, NormalizedSchema)
│   ├── parser.go            // SpecParser & RouteExtractor implementation
│   └── parser_test.go
├── generator/
│   ├── generator.go         // MockGenerator & OverrideMerger logic
│   └── generator_test.go
└── data/
    └── builder.go           // ValueProvider & PayloadBuilder interface stubs
```
