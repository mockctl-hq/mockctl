# 📐 PKS-025 — Software Design Document (Master SDD)

> **Project:** Mock:ctl
>
> **Document ID:** PKS-025
>
> **Version:** 1.0
>
> **Status:** Approved
>
> **Owner:** Upen Tudu
>
> **Authors:** Upen Tudu & Antigravity
>
> **Created:** 2026-08-15
>
> **Last Updated:** 2026-08-15
>
> **Category:** Engineering
>
> **Priority:** Critical

---

# 📖 Executive Summary

The Software Design Document (Master SDD) is the final engineering blueprint before implementation begins.

Building upon the Component Architecture (PKS-024), this document defines the concrete Go structs, interface method signatures, and critical data models required to build Mock:ctl.

By standardizing these code-level definitions beforehand, we ensure that multiple developers (or AI coding agents) can work on different modules concurrently without facing integration mismatches later.

---

# 🎯 Purpose

The objectives of this document are to:

- Translate architectural components into concrete Go code definitions.
- Define the primary structs and their fields.
- Define the exact method signatures for core interfaces.
- Eliminate ambiguity in cross-boundary communication.
- Establish the Composition Root and CLI entry points.
- Serve as the strict coding contract for the implementation phase.

---

# 📌 Scope

This document covers the core Go design for:

- Shared Domain Types (`internal/shared`)
- Storage Implementation (`internal/storage`)
- Configuration Models (`internal/config`)
- Project & Overrides (`internal/project`)
- Specification Models (`internal/spec`)
- Data Generators (`internal/data`)
- Mock Generation (`internal/generator`)
- Runtime Engine & Chaos (`internal/runtime`)
- Application Orchestrator (`internal/app`)
- Presentation Layer (`cmd/`)

This document does not contain internal business logic or private helper functions. It focuses entirely on public APIs, structs, and inter-package contracts.

---

# 🧱 1. Shared Domain Types (`internal/shared`)

The shared package contains primitives used across the entire system.

## 1.1 Custom Error Types

All components must return `DomainError` to allow consistent HTTP/CLI translation.

```go
package shared

type ErrorCode string

const (
    ErrCodeInvalidSpec  ErrorCode = "ERR_SPEC_INVALID"
    ErrCodeNotFound     ErrorCode = "ERR_NOT_FOUND"
    ErrCodeInternal     ErrorCode = "ERR_INTERNAL"
    ErrCodeValidation   ErrorCode = "ERR_VALIDATION"
)

type DomainError struct {
    Code       ErrorCode
    Message    string
    HTTPStatus int
    Err        error // Underlying error for stack traces
}

func (e *DomainError) Error() string {
    return e.Message
}
```

## 1.2 Logger Interface

To prevent direct dependencies on `log`, all components use this interface.

```go
package shared

type Logger interface {
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, err error, args ...any)
    Debug(msg string, args ...any)
}
```

---

# 💾 2. Storage Design (`internal/storage`)

Abstracts the OS file system to enable fully in-memory testing.

```go
package storage

import "context"

type FileSystem interface {
    ReadFile(ctx context.Context, path string) ([]byte, error)
    WriteFile(ctx context.Context, path string, data []byte) error
    Exists(ctx context.Context, path string) bool
}

// LocalFS is the concrete implementation used in production.
type LocalFS struct {}

func NewLocalFS() *LocalFS {
    return &LocalFS{}
}

// SystemStore handles persistent monetization and configuration data via bbolt.
type SystemStore interface {
    GetSetting(ctx context.Context, key string) (string, error)
    SetSetting(ctx context.Context, key string, value string) error
    SaveAuthToken(ctx context.Context, jwtToken string) error
    GetAuthToken(ctx context.Context) (string, error)
    LogTelemetry(ctx context.Context, event string, data []byte) error
    Close() error
}

// BoltSystemStore is the concrete implementation using go.etcd.io/bbolt
type BoltSystemStore struct {
    db *bbolt.DB
}
```

---

# ⚙️ 3. Configuration Design (`internal/config`)

Handles resolving the user's workspace config.

```go
package config

import "context"

type AppConfig struct {
    ProjectName string
    Port        int
    StrictParse bool
}

type ConfigLoader interface {
    Load(ctx context.Context, configPath string) (*AppConfig, error)
}
```

---

# 📁 4. Project & Overrides (`internal/project`)

Manages workspace settings and user-defined custom JSON overrides.

```go
package project

import "context"

// CustomPayload represents user-defined static mock responses
type CustomPayload struct {
    Path     string
    Method   string
    Response map[string]any
}

type WorkspaceContext struct {
    RootPath  string
    Overrides []CustomPayload
}

type ProjectManager interface {
    Initialize(ctx context.Context, path string) error
    LoadContext(ctx context.Context, path string) (*WorkspaceContext, error)
}
```

---

# 📑 5. Specification Design (`internal/spec`)

Parses OpenAPI files using `kin-openapi` and produces a read-only model.

```go
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

type SpecParser interface {
    ParseFile(ctx context.Context, path string) (*SpecModel, error)
}
```

---

# 🎲 6. Data Generation Design (`internal/data`)

Wraps `gofakeit` to generate realistic data payloads.

```go
package data

import "context"

type ValueProvider interface {
    GenerateString(format string) string
    GenerateInt(min, max int) int
    GenerateBoolean() bool
}

type PayloadBuilder interface {
    BuildFromSchema(ctx context.Context, schemaRef any) (map[string]any, error)
}
```

---

# 🚀 7. Mock Generation Design (`internal/generator`)

Combines `SpecModel`, custom overrides, and `ValueProvider` into a runtime blueprint.

```go
package generator

import (
    "context"
    "github.com/mockctl/mockctl/internal/spec"
    "github.com/mockctl/mockctl/internal/project"
    "github.com/mockctl/mockctl/internal/data"
    "github.com/mockctl/mockctl/internal/shared"
)

type RuntimeDefinition struct {
    Endpoints map[string]EndpointHandler
}

// ResponseTemplate holds the payload structure for a specific HTTP status
type ResponseTemplate struct {
    Headers map[string]string
    Body    map[string]any
}

type EndpointHandler struct {
    Method      string
    Path        string
    PathParams  []string // e.g., ["id"] for /users/{id}
    QueryParams []string // Expected query parameters
    Responses   map[int]ResponseTemplate // Maps HTTP status (e.g. 200, 404) to its template
}

type MockGenerator struct {
    logger   shared.Logger
    provider data.ValueProvider
}

func NewMockGenerator(l shared.Logger, p data.ValueProvider) *MockGenerator {
    return &MockGenerator{logger: l, provider: p}
}

// Generate merges the OpenAPI spec with user Overrides to create the final routes
func (g *MockGenerator) Generate(ctx context.Context, model *spec.SpecModel, ctx *project.WorkspaceContext) (*RuntimeDefinition, error)
```

---

# 🧠 8. Runtime Engine & Chaos Design (`internal/runtime`)

The stateful engine that processes active HTTP requests and simulates errors.

```go
package runtime

import (
    "context"
    "net/http"
    "sync"
    "github.com/mockctl/mockctl/internal/generator"
    "github.com/mockctl/mockctl/internal/shared"
)

// ChaosEvaluator determines if a request should fail or delay
type ChaosEvaluator interface {
    Evaluate(ctx context.Context, r *http.Request) *shared.DomainError
}

// StateStore handles full CRUD memory persistence for dynamic simulation
type StateStore interface {
    Insert(ctx context.Context, collection string, data map[string]any) error
    Get(ctx context.Context, collection, id string) (map[string]any, error)
    List(ctx context.Context, collection string) ([]map[string]any, error)
    Update(ctx context.Context, collection, id string, data map[string]any) error
    Delete(ctx context.Context, collection, id string) error
}

// MemoryStateStore is the thread-safe implementation
type MemoryStateStore struct {
    mu    sync.RWMutex
    store map[string]map[string]any
}

type RuntimeEngine struct {
    logger     shared.Logger
    definition *generator.RuntimeDefinition
    state      StateStore
    chaos      ChaosEvaluator
}

func NewRuntimeEngine(l shared.Logger, def *generator.RuntimeDefinition, store StateStore, chaos ChaosEvaluator) *RuntimeEngine {
    return &RuntimeEngine{
        logger:     l,
        definition: def,
        state:      store,
        chaos:      chaos,
    }
}

// ServeHTTP makes RuntimeEngine compatible with net/http routers (like Chi)
func (e *RuntimeEngine) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

---

# 🧩 9. WebAssembly Plugin Host (`internal/plugin`)

Loads and executes WASM plugins using `wazero` to support future extensibility (EDL-046, EDL-047).

```go
package plugin

import "context"

// PluginHost manages the lifecycle of loaded WebAssembly modules
type PluginHost interface {
    LoadPlugin(ctx context.Context, path string) error
    ExecuteMiddleware(ctx context.Context, pluginName string, payload []byte) ([]byte, error)
    Close(ctx context.Context) error
}
```

---

# 🧬 10. Application Orchestrator (`internal/app`)

Acts as the Composition Root. It wires all isolated components together so `cmd/` can execute them.

```go
package app

import (
    "context"
    "github.com/mockctl/mockctl/internal/shared"
    "github.com/mockctl/mockctl/internal/storage"
    "github.com/mockctl/mockctl/internal/project"
)

type App struct {
    logger         shared.Logger
    fs             storage.FileSystem
    projectManager project.ProjectManager
}

// NewApp performs strict Dependency Injection for the entire system
func NewApp(logger shared.Logger, fs storage.FileSystem) *App {
    // Wiring occurs here
    pm := project.NewProjectManager(logger, fs)
    return &App{
        logger:         logger,
        fs:             fs,
        projectManager: pm,
    }
}

// StartServer orchestrates reading the spec, generating the mock, and starting the HTTP server
func (a *App) StartServer(ctx context.Context, projectPath string) error
```

---

# ⌨️ 11. Presentation Layer (`cmd/`)

The CLI interface built using Cobra (`spf13/cobra`). It is extremely thin and delegates all work to `internal/app`.

```go
package cmd

import (
    "context"
    "github.com/spf13/cobra"
    "github.com/mockctl/mockctl/internal/app"
    "github.com/mockctl/mockctl/internal/shared"
    "github.com/mockctl/mockctl/internal/storage"
)

func Execute() error {
    rootCmd := &cobra.Command{
        Use:   "mockctl",
        Short: "Mock:ctl - Developer-first backend simulation",
    }

    startCmd := &cobra.Command{
        Use:   "start [path]",
        Short: "Start a mock server for a specific project",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := cmd.Context()
            
            logger := shared.NewConsoleLogger()
            fs := storage.NewLocalFS()
            
            application := app.NewApp(logger, fs)
            
            path := "."
            if len(args) > 0 {
                path = args[0]
            }
            
            return application.StartServer(ctx, path)
        },
    }

    rootCmd.AddCommand(startCmd)
    return rootCmd.ExecuteContext(context.Background())
}
```

---

# 📱 12. Desktop & Android FFI Bindings (`cmd/mockctl-ffi`)

Exports C-compatible functions using `cgo` so the Go Backend can be embedded in Flutter (EDL-040, EDL-044).

```go
package main

import "C"
import (
    "context"
    "github.com/mockctl/mockctl/internal/app"
    "github.com/mockctl/mockctl/internal/shared"
    "github.com/mockctl/mockctl/internal/storage"
)

//export StartMockServer
func StartMockServer(cProjectPath *C.char) int {
    projectPath := C.GoString(cProjectPath)
    
    logger := shared.NewConsoleLogger() // Or FFI-bridged logger
    fs := storage.NewLocalFS()
    
    application := app.NewApp(logger, fs)
    
    // Run in background and return status
    err := application.StartServer(context.Background(), projectPath)
    if err != nil {
        return 1 // Error code
    }
    return 0 // Success code
}

func main() {} // Required for c-shared build mode
```

---

# 📌 Conclusion

The Software Design Document (Master SDD) establishes the final coding blueprints for Mock:ctl's internal Go architecture. 
By defining the exact structs, method signatures, FFI bindings, and WASM plugin hooks beforehand, we have eliminated ambiguity from the implementation phase. Developers can now map these strict interfaces directly into Go code, confident that the components will integrate seamlessly.

---

# 🔗 Related Documents

**Foundation**

- PKS-000 — Repository Blueprint
- PKS-002 — Documentation Style Guide

**Engineering**

- PKS-020 — System Architecture
- PKS-021 — Technology Stack
- PKS-022 — Repository & Module Architecture
- PKS-023 — Data Flow Architecture
- PKS-024 — Component Architecture

**Next Document**

- PKS-026 — Database Design (Memory State Architecture)

---

# 📜 Revision History

| Version | Date | Description |
|----------|------------|----------------------------------------------|
| 1.0 | 2026-08-15 | Initial approved release |

---

# ✅ Approval Checklist

- Executive summary completed
- Shared domain models (`DomainError`, `Logger`) defined
- Storage interface mapped
- Configuration struct mapped
- Project overrides struct defined
- Specification model mapped
- Data generation interface mapped
- Mock generator mapped (Multi-status & Path variables supported)
- Runtime engine mapped (StateStore Full CRUD)
- Chaos Evaluator mapped
- WebAssembly Plugin Host mapped
- Application Orchestrator (Composition Root) defined
- Cobra CLI commands (`cmd/`) mapped
- Embedded Go Core FFI bindings mapped
- Thread safety annotations added
- Constructor injection patterns verified
- Formatting follows PKS style guide

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** Completed

**Architecture Status:** ✅ Established

**Next Document:** **PKS-026 — Database Design**
