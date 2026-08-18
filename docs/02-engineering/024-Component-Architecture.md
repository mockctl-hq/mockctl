# 🧱 PKS-024 — Component Architecture

> **Project:** Mock:ctl
>
> **Document ID:** PKS-024
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
> **Priority:** High

---

# 📖 Executive Summary

The Component Architecture bridges the gap between the high-level repository structure (PKS-022) and the detailed software design (PKS-025).

While PKS-022 defined the directories and boundaries, this document defines the specific internal interfaces, primary Data Transfer Objects (DTOs), concurrency rules, and dependency injection strategies that live inside those boundaries.

By establishing strict interfaces and principles for each component, Mock:ctl ensures that its subsystems remain loosely coupled, highly testable, thread-safe, and capable of independent evolution without monolithic entanglement.

---

# 🎯 Purpose

The objectives of this document are to:

- Define the primary software components within each repository module.
- Establish the exact Go interfaces that components use to communicate.
- Define the structural boundaries for DTOs flowing between modules.
- Enforce Go-idiomatic principles (Narrow Interfaces, Context, Concurrency).
- Formalize the boundaries that isolate third-party dependencies.
- Standardize error handling and logging architectures.
- Enforce unit-testing contracts via isolated component mocks.

---

# 📌 Scope

This document covers:

- Core Component Definitions (including Shared, Config, and Presentation).
- Key Go Interfaces and Cross-Component DTOs.
- Dependency Injection Architecture.
- Error Handling and Logging Architecture.
- The Narrow Interface Principle.
- Lifecycle, Context Management, and Thread-Safety Rules.
- Testing and Mocking Contracts.
- Future Plugin Extension Interfaces.

This document does not define the granular method signatures or implementation logic, which will be covered in the Master Software Design Document (PKS-025).

---

# 📐 Go Component Design Principles

Mock:ctl strictly adheres to the following Go-specific component design principles.

**1. The Narrow Interface Principle**
Components must not expose massive, monolithic interfaces (e.g., a 20-method `Database` interface). 
Instead, interfaces must be small, role-specific, and defined by the *consumer*, not the *provider* (e.g., `StateReader`, `StateWriter`).

**2. Accept Interfaces, Return Structs**
Component constructors should accept interfaces for their dependencies but return concrete structs (or pointers to structs). This avoids unnecessary interface abstraction overhead.

**3. No Global State**
`init()` functions and global variables are strictly forbidden for component wiring. All components must be instantiated and injected explicitly.

**4. Context Management**
Any component function that involves I/O, delays, or deep execution trees must accept `ctx context.Context` as its first parameter. This ensures the HTTP server or CLI can cancel running processes gracefully.

**5. Thread Safety**
All components intended to be used by the HTTP server (e.g., `RuntimeEngine`, `StateStore`) must be designed for concurrent execution. Mutable internal state must be strictly guarded by `sync.RWMutex` or channel-based synchronization.

---

# 🏗️ Core Component Definitions

Every module in Mock:ctl is built around one or more core components. 

A component is defined as a cohesive set of Go interfaces and their implementations.

---

## 1️⃣ Configuration Component (`internal/config`)

The Configuration Component manages the resolution of application parameters.

**Key Interfaces:**
- `ConfigLoader`: Responsible for reading configuration from sources (file, env, flags).
- `ConfigValidator`: Ensures the loaded configuration is semantically valid.

**Primary DTO:**
- `AppConfig`: The unified struct containing all resolved settings.

---

## 2️⃣ Project Component (`internal/project`)

The Project Component manages the lifecycle of a Mock:ctl workspace.

**Key Interfaces:**
- `WorkspaceInitializer`: Sets up the required `.mockctl` directory structure.
- `ProjectReader`: Reads existing project state and overrides.

**Primary DTO:**
- `WorkspaceContext`: Holds the active project's file paths and metadata.

---

## 3️⃣ Specification Component (`internal/spec`)

The Specification Component encapsulates OpenAPI parsing.

**Key Interfaces:**
- `SpecParser`: Parses raw data into an internal model.
- `RouteExtractor`: Iterates over the parsed model to yield clean route definitions.

**Primary DTO:**
- `SpecModel`: The strictly read-only, normalized internal schema tree.

**Isolation Boundary:**
This component acts as an anti-corruption layer around `kin-openapi`. No package outside of `internal/spec/` is permitted to import or use `kin-openapi` types.

---

## 4️⃣ Data Generation Component (`internal/data`)

The Data Generation Component provides realistic fake data for mock responses.

**Key Interfaces:**
- `ValueProvider`: Generates a specific fake primitive (e.g., `GenerateEmail()`).
- `PayloadBuilder`: Constructs complex JSON structures based on an OpenAPI schema tree.

**Isolation Boundary:**
This component wraps `gofakeit`. No package outside of `internal/data/` is permitted to import or use `gofakeit`.

---

## 5️⃣ Mock Generation Component (`internal/generator`)

The Mock Generation Component translates specifications into executable runtime routes.

**Key Interfaces:**
- `MockGenerator`: Takes an internal specification model and returns executable routes.
- `OverrideMerger`: Applies user-defined custom payloads over auto-generated mocks.

**Primary DTO:**
- `RuntimeDefinition`: The structured blueprint containing compiled routes, methods, and templates.

---

## 6️⃣ Runtime Component (`internal/runtime`)

The Runtime Component acts as the dynamic engine simulating the backend.

**Key Interfaces:**
- `RuntimeEngine`: The core executor that processes incoming requests.
- `StateReader` / `StateWriter`: Focused interfaces for the in-memory database used during stateful simulations.
- `ChaosEvaluator`: Determines if a request should intentionally fail or delay.

**Primary DTO:**
- `StateSnapshot`: A serializable representation of the current in-memory database.

---

## 7️⃣ Presentation & HTTP Component (`cmd/` & `net/http`)

The Presentation layer acts as the primary entry point and transport layer.

**Key Interfaces:**
- `HTTPServer`: Wraps the `net/http` server and `Chi` router for graceful startup/shutdown.
- `CLICommand`: Represents an executable terminal command.

**Isolation Boundary:**
The `Chi` router strictly resides here. The internal `RuntimeEngine` receives standardized Go `http.Request` and `http.ResponseWriter` objects, remaining unaware of the specific routing framework.

---

## 8️⃣ Storage Component (`internal/storage`)

The Storage Component abstracts all file system and persistent configuration database interactions.

**Key Interfaces:**
- `FileReader` / `FileWriter`: Narrow interfaces for disk I/O.
- `PathResolver`: Safely constructs and validates absolute file paths.
- `SystemStore`: Handles permanent embedded storage for Monetization (Licenses, Settings) and Telemetry using `bbolt`.

---

# 🛡️ Shared Contracts: Errors & Logging (`internal/shared`)

The `internal/shared/` module defines the foundational contracts used across all components.

**Error Handling Architecture:**
Mock:ctl utilizes a custom `DomainError` struct that implements the standard Go `error` interface.
- It includes: `Code`, `Message`, `HTTPStatus`, and an inner `Err` for stack tracing.
- Components must always return `DomainError`. This allows the Presentation layer to consistently translate internal failures into appropriate 500s/400s or CLI messages without duplicating logic.

**Logging Boundary:**
Direct calls to the standard library `log` package (e.g., `log.Println`) are strictly prohibited inside core components.
- Components must accept a `shared.Logger` interface (e.g., methods for `Info()`, `Debug()`, `Error()`).
- The actual implementation (which handles terminal colors or JSON formatting) is injected at startup.

---

# 🔄 Lifecycle & Concurrency Model

**Stateless Components:**
Components like `SpecParser`, `MockGenerator`, and `DataGeneration` must be completely stateless after initialization. They must safely support simultaneous access from multiple goroutines without requiring mutex locks.

**Stateful Components:**
Components like `StateStore` (in `internal/runtime/`) hold mutable memory. These components *must* wrap their internal maps or slices with a `sync.RWMutex` to prevent data races during concurrent HTTP traffic.

**Graceful Shutdown:**
The HTTP Server Component listens for OS termination signals (SIGINT/SIGTERM) and propagates cancellation via `context.Context` to all active subsystems to ensure safe state snapshots before exit.

---

# 🧪 Testing & Mocking Contracts

Every core component interface defined in this architecture must be designed to be completely mockable.

Mock:ctl enforces the following testing rules:
1. No unit test may touch the physical disk (use injected `StorageComponent` mocks).
2. No unit test may open a real network port (use Go's `httptest` package).
3. Auto-generated mock implementations (e.g., via `moq` or `gomock`) must be generated for all critical interfaces (like `ValueProvider`, `StateStore`) to allow other components to be tested in total isolation.

---

# 💉 Dependency Injection Strategy

The `internal/app/` package acts as the application's **Composition Root**. 

It is responsible for:
1. Instantiating the `StorageComponent` and `Logger`.
2. Resolving the `ConfigurationComponent`.
3. Wiring dependencies into the `ProjectManager`, `RuntimeEngine`, etc.
4. Starting the CLI or HTTP Server.

```go
// Example of strict constructor injection in Mock:ctl
func NewMockGenerator(logger shared.Logger, provider data.ValueProvider, overrides spec.OverrideReader) *Generator {
    return &Generator{
        logger:    logger,
        provider:  provider,
        overrides: overrides,
    }
}
```

---

# 🔌 Plugin Architecture (Future)

To support the future Data Flow extension point (PKS-023), the component architecture reserves a space for a Plugin Manager.

**Key Future Interfaces:**
- `PluginHost`: Manages the lifecycle of WASM plugins (via `wazero`).
- `MiddlewareInterceptor`: An interface that allows plugins to mutate a request before the `RuntimeEngine` processes it, or mutate a response before it is sent to the client.

This ensures the `RuntimeEngine` remains unaware of whether it is interacting with a plugin or native Go code.

---

# 📌 Conclusion

The Component Architecture successfully bridges the gap between high-level modules and low-level code by defining explicit, narrow Go interfaces and cross-boundary DTOs.
By strictly adhering to dependency injection and prohibiting global state, this design guarantees that Mock:ctl will remain highly testable, scalable, and resilient to future changes.

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

**Next Document**

- PKS-025 — Software Design Document (Master SDD)

---

# 📜 Revision History

| Version | Date | Description |
|----------|------------|----------------------------------------------|
| 1.0 | 2026-08-15 | Initial approved release |

---

# ✅ Approval Checklist

- Executive summary completed
- Go component design principles defined
- Concurrency and context rules established
- Project, Spec, Data, Generator, Runtime, Storage components defined
- Presentation and HTTP Server components defined
- Error handling and Logging architecture formalized
- Lifecycle and graceful shutdown rules defined
- Testing and mocking contracts enforced
- Third-party isolation strategies documented
- Dependency injection architecture established
- Future plugin extension interfaces defined
- Formatting follows PKS style guide

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** Completed

**Architecture Status:** ✅ Established

**Next Document:** **PKS-025 — Software Design Document (Master SDD)**
