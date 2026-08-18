# PKS-022 — Repository & Module Architecture

> **Project:** Mock:ctl
>
> **Document ID:** PKS-022
>
> **Version:** 1.0
>
> **Status:** Approved
>
> **Owner:** Upen Tudu
>
> **Authors:** Upen Tudu & ChatGPT
>
> **Created:** 2026-08-09
>
> **Last Updated:** 2026-08-09
>
> **Category:** Engineering
>
> **Priority:** Critical

---

# 📖 Executive Summary

This document defines the official repository structure and module architecture of Mock:ctl.

It translates the high-level architecture established by **PKS-020 — System Architecture** and the technology decisions established by **PKS-021 — Technology Stack** and the approved **Engineering Decision Log** into a concrete repository and module organization.

Mock:ctl follows an **Internal-First Modular Repository Architecture** built around a single Go module.

The repository is organized to provide:

- Clear responsibility ownership.
- Explicit module boundaries.
- Predictable dependency direction.
- Separation between presentation, application, domain, and infrastructure concerns.
- Isolation of external technology dependencies.
- Platform-independent shared business logic.
- A stable foundation for CLI, Desktop, Android, and future interfaces.
- A repository structure that is understandable to both human developers and AI coding agents.
- Incremental evolution without unnecessary architectural rewrites.

The repository structure defined here is an architectural contract.

It is not merely a filesystem convention.

---

# 🎯 Purpose

The purpose of this document is to define:

- The official Mock:ctl repository structure.
- The responsibility of each major repository directory.
- The responsibility of each core Go module.
- The relationship between repository modules and the system architecture.
- Module ownership boundaries.
- Dependency direction.
- Module communication boundaries.
- External dependency isolation.
- Platform abstraction boundaries.
- Shared-code rules.
- Repository evolution rules.
- AI-assisted development constraints.

This document establishes the repository foundation required by the remaining engineering documents.

---

# 📌 Scope

This document defines:

- Repository structure.
- Root-level directory responsibilities.
- Go module organization.
- `cmd/` responsibilities.
- `internal/` responsibilities.
- Core internal module responsibilities.
- System-subsystem to repository-module mapping.
- Architectural layer mapping.
- Dependency direction.
- Module communication rules.
- External dependency boundaries.
- Configuration boundaries.
- Storage boundaries.
- Platform boundaries.
- Plugin-readiness boundaries.
- Test-directory responsibilities.
- Repository evolution rules.
- AI-assisted implementation rules.

This document does **not** define:

- Detailed data flows.
- Detailed component interactions.
- Database schema.
- API contracts.
- Coding conventions.
- Testing methodology.
- Deployment procedures.

Those concerns are defined by subsequent engineering documents.

---

# 🏛 Repository Architecture Philosophy

Mock:ctl follows an **Internal-First Modular Repository Architecture**.

The repository structure must make architectural responsibility visible.

The primary principles are:

1. **Clear responsibility over convenience.**
2. **Internal modules over unnecessary public packages.**
3. **Explicit dependencies over hidden dependencies.**
4. **Stable boundaries over implementation leakage.**
5. **Focused modules over large multi-purpose modules.**
6. **Shared business logic over duplicated platform implementations.**
7. **Deterministic behavior over implicit behavior.**
8. **Documentation before implementation.**
9. **AI-readable structure over clever abstractions.**
10. **Incremental evolution over unnecessary rewrites.**

The repository should communicate the architecture to a developer or AI coding agent through its structure, naming, and module boundaries.

---

# 🧭 Architectural Foundation

PKS-020 establishes the following high-level architecture:

```text
Presentation
     ↓
Application Core
     ↓
Domain Services
     ↓
Infrastructure Services
     ↓
File System / Operating System / Runtime
```

The repository structure translates these architectural layers into concrete Go packages.

The primary relationship is:
```text
┌───────────────────────────────────────┐
│              cmd/                     │
│          Presentation Layer            │
└──────────────────┬────────────────────┘
                   │
                   ▼
┌───────────────────────────────────────┐
│            internal/app/              │
│          Application Core             │
└──────────────────┬────────────────────┘
                   │
          ┌────────┼─────────┐
          ▼        ▼         ▼
┌────────────┐ ┌──────────┐ ┌────────────┐
│  project   │ │   spec   │ │ generator  │
│  Project   │ │   Spec   │ │    Mock    │
│  Manager   │ │  Engine  │ │ Generation │
└────────────┘ └──────────┘ └────────────┘
          │        │         │
          └────────┼─────────┘
                   │
                   ▼
          ┌────────────────────┐
          │      data/         │
          │ Data Generation    │
          └────────────────────┘
                   │
                   ▼
          Infrastructure Services
          ┌────────┼────────────┐
          ▼        ▼            ▼
      runtime/  config/     storage/
```

This structure reflects the subsystem organization established by PKS-020.


---

## 📂 Official Repository Structure

The official repository structure is:

```text
mockctl/
├── cmd/
├── internal/
│   ├── app/
│   ├── project/
│   ├── spec/
│   ├── generator/
│   ├── runtime/
│   ├── config/
│   ├── storage/
│   ├── data/
│   └── shared/
├── docs/
├── scripts/
├── test/
├── assets/
└── go.mod
```

Additional root-level files required by the Go toolchain, Git, CI/CD, release automation, or project governance may exist.

New top-level directories must not be introduced without architectural justification and documentation alignment.


---

## 📦 Repository Directory Responsibilities

**`cmd/`**

**Architectural Role**

Presentation Layer / Application Entry Points.

**Responsibility**

The `cmd/` directory contains executable entry points for Mock:ctl.

It represents the boundary through which users and platform-specific application shells interact with the shared application core.

**Responsibilities**

- Define executable entry points.
- Initialize the application.
- Initialize CLI commands.
- Connect presentation concerns to the Application Core.
- Handle process-level startup and shutdown.
- Translate user-facing commands into application operations.

**Does Not Own**

- Business logic.
- Project domain rules.
- OpenAPI parsing.
- Mock generation rules.
- Data-generation rules.
- Storage implementation.
- Runtime internals.

The CLI must remain a thin presentation layer.


---

**`internal/`**

**Architectural Role**

Private Application Implementation.

**Responsibility**

The `internal/` directory contains Mock:ctl's private implementation.

Core application logic should remain inside `internal/` unless a future Engineering Decision explicitly establishes a public Go package.

The internal-first structure protects implementation details and prevents accidental exposure of internal APIs.

**Responsibilities**

- Application orchestration.
- Project management.
- Specification processing.
- Mock generation.
- Data generation.
- Runtime execution.
- Configuration management.
- Storage.
- Shared internal abstractions.


---

**`internal/app/`**

**Architectural Role**

Application Core.

**Responsibility**

The Application Core coordinates application-level workflows.

**Responsibilities**

- Orchestrate use cases.
- Coordinate Mock:ctl subsystems.
- Route application operations.
- Manage application-level sequencing.
- Coordinate lifecycle operations.
- Provide application-facing interfaces to presentation layers.


**Does Not Own**

- CLI parsing implementation.
- OpenAPI parser implementation.
- Storage implementation.
- Fake-data provider implementation.
- HTTP router implementation.
- Platform-specific UI behavior.

The Application Core coordinates subsystem responsibilities. It must not become a replacement for those subsystems.


---

**`internal/project/`**

**Architectural Role**

Project Manager.

**Responsibility**

The `project/` module owns the lifecycle and domain representation of a Mock:ctl project.

**Responsibilities**

- Project creation.
- Project opening.
- Project validation.
- Project metadata.
- Project lifecycle operations.
- Project-level domain rules.

**Does Not Own**

- CLI command parsing.
- OpenAPI parser implementation.
- Runtime server implementation.
- Generic filesystem infrastructure.
- HTTP routing.

Project-specific behavior belongs in the Project Manager rather than in presentation or infrastructure modules.

---

**`internal/spec/`**

**Architectural Role**

Specification Engine.

**Responsibility**

The `spec/` module owns the processing of API specifications.

**Responsibilities**

- OpenAPI parsing.
- Specification validation.
- Specification normalization.
- Internal specification representation.
- Component resolution.
- Reference resolution.
- Parser adapter integration.

**External Dependency Boundary**

The external OpenAPI parser must remain isolated behind an internal abstraction.

The rest of Mock:ctl should consume the internal specification representation rather than directly depending on external parser-library types.

The approved OpenAPI parser is `kin-openapi`.

**Does Not Own**

- Mock response generation.
- Runtime server execution.
- CLI presentation.
- Persistent storage implementation.

---

**`internal/generator/`**

**Architectural Role**

Mock Generation Engine.

**Responsibility**

The `generator/` module transforms validated internal API specifications into executable mock definitions.

**Responsibilities**

- Endpoint generation.
- Response template generation.
- Mock handler preparation.
- Runtime definition preparation.
- Generation rules.
- Generated mock configuration.

**Does Not Own**

- OpenAPI parsing.
- HTTP server startup.
- CLI command parsing.
- Persistent storage implementation.

The generator consumes the internal specification representation rather than external parser-library objects.


---

**`internal/runtime/`**

**Architectural Role**

Runtime Engine.

**Responsibility**

The runtime/ module executes the generated mock backend.

It belongs to the Infrastructure Services layer defined by PKS-020.

**Responsibilities**

- Runtime startup.
- Runtime shutdown.
- Request routing.
- Handler execution.
- Runtime state management.
- HTTP middleware integration.
- Mock response delivery.

**Technology Boundary**

The approved HTTP runtime foundation consists of:

- Go net/http.
- Chi.

HTTP-specific implementation details remain inside the Runtime Engine boundary.

**Does Not Own**

- CLI command implementation.
- OpenAPI parsing.
- Project lifecycle management.
- Configuration-file parsing.
- Project persistence rules.

The Runtime Engine must not bypass the Application Core to manipulate Project Manager internals.


---

**`internal/config/`**

**Architectural Role**

Configuration Manager.

**Responsibility**

The `config/` module owns configuration loading, validation, merging, and resolution.

It belongs to the Infrastructure Services layer defined by PKS-020.

**Responsibilities**

- Load configuration.
- Save configuration.
- Apply defaults.
- Resolve configuration precedence.
- Validate configuration.
- Provide resolved configuration to consuming modules.


**Configuration Precedence**

Configuration resolution follows the approved layered model:

```text
Built-in Defaults
        ↓
Global Configuration
        ↓
Project Configuration
        ↓
Environment Overrides
        ↓
Command-Line Arguments
```

Higher-priority values override lower-priority values.

Configuration resolution must remain deterministic.

**Does Not Own**

- Business logic.
- CLI command behavior.
- Runtime business rules.
- Project lifecycle rules.

Individual modules must not independently implement configuration precedence.


---

**`internal/storage/`**

**Architectural Role**

Storage Layer.

**Responsibility**

The `storage/` module owns persistent project storage and filesystem interaction.

It belongs to the Infrastructure Services layer defined by PKS-020.

**Responsibilities**

- Read files.
- Write files.
- Manage directories.
- Serialize and deserialize persisted data.
- Perform project-boundary filesystem operations.
- Provide storage abstractions to higher layers.

**Architectural Rule**

Storage must remain independent from business logic.

Higher-level modules should interact with storage through appropriate interfaces rather than spreading filesystem implementation details throughout the application.

**Does Not Own**

- Project business rules.
- CLI behavior.
- Runtime business logic.
- Specification interpretation.
- Mock-generation rules.


---

**`internal/data/`**

**Architectural Role**

Data Generation Engine.

**Responsibility**

The `data/` module owns realistic contextual fake-data generation required by the mock simulation system.

It belongs to the Domain Services layer defined by PKS-020.

**Responsibilities**

- Generate contextual fake values.
- Generate realistic entities.
- Support deterministic generation.
- Manage fake-data providers.
- Provide data-generation abstractions to consuming modules.

**External Dependency Boundary**

The approved fake-data technology is `gofakeit`.

The external library must remain isolated behind the internal data-generation abstraction.

Business logic should not directly depend on `gofakeit` implementation details.

**Determinism**

The data-generation subsystem must support deterministic behavior:

Same Input + Same Seed
          ↓
     Same Output

Random behavior may be supported where explicitly required, but deterministic behavior is the default.


---

**`internal/shared/`**

**Architectural Role**

Shared Internal Infrastructure.

**Responsibility**

The `shared/` module contains genuinely reusable internal abstractions that do not have stronger ownership within another subsystem.

**Valid Uses**

Examples may include:

- Common internal interfaces.
- Shared internal error types.
- Cross-cutting internal primitives.
- Small abstractions genuinely required by multiple independent modules.


**Strict Rule**

`shared/` must not become a general-purpose dumping ground.

A piece of functionality belongs in `shared/` only when:

1. It is genuinely shared.
2. It has no stronger ownership in another module.
3. Reusing it improves architectural clarity.
4. Moving it into `shared/` does not hide responsibility.


If functionality clearly belongs to `project`, `spec`, `generator`, `runtime`, `config`, `storage`, or `data`, it must remain there.


---

### 📚 **`docs/`**

**Architectural Role**

Project Documentation.

**Responsibility**

The `docs/` directory contains the Project Knowledge System and project documentation.

Documentation is part of the project's engineering system and serves as an authoritative source of truth for implementation.

**Responsibilities**

- Foundation documentation.
- Product documentation.
- Engineering documentation.
- AI-related documentation.
- Future approved documentation categories.

Implementation should follow approved documentation rather than defining architecture after implementation.


---

### 🛠 **`scripts/`**

**Architectural Role**

Development and Automation Support.

**Responsibility**

The `scripts/` directory contains project-controlled scripts used to support development, automation, or repetitive repository operations.

Scripts supplement the official development toolchain and must not create a competing source of truth for project behavior.

Detailed scripting conventions belong to the appropriate engineering documentation.


---

### 🧪 **`test/`**

**Architectural Role**

Repository-Level Test Support.

**Responsibility**

The `test/` directory contains repository-level test resources that do not naturally belong beside a specific implementation package.

Examples include:

- Integration-test fixtures.
- End-to-end test resources.
- Shared test assets.
- Repository-level golden-test resources where appropriate.

Package-specific unit tests should remain alongside the package they test.

Detailed testing methodology is defined by `PKS-029 — Testing Strategy`.


---

### 🎨 **`assets/`**

**Architectural Role**

Project-Controlled Non-Source Assets.

**Responsibility**

The `assets/` directory contains non-source assets required by project tooling, documentation, examples, or application resources.

Assets must remain clearly owned and must not become a location for executable business logic.

Detailed asset usage may be further defined by later engineering documentation when required.


---

## 📦 Go Module Architecture

Mock:ctl uses a **single Go module**.

The repository contains one authoritative go.mod at the repository root:

```text
mockctl/
└── go.mod
```

Additional Go modules must not be introduced without a new approved Engineering Decision.

The single-module architecture preserves:

- Consistent dependency management.
- Simple local development.
- Straightforward AI-assisted navigation.
- Unified tooling.
- Reduced repository complexity.


---

## 🧩 Core Module Structure

The core internal repository modules are:

```text
internal/
├── app/
├── project/
├── spec/
├── generator/
├── runtime/
├── config/
├── storage/
├── data/
└── shared/
```
Each module has one primary architectural responsibility.

The directory name should communicate the responsibility clearly enough that a developer or AI coding agent can determine the likely ownership of new functionality before inspecting implementation details.


---

# 🔗 Module Responsibility & Dependency Architecture

The repository structure is only useful when module boundaries are enforced consistently.

Each module must have a clearly defined responsibility, and dependencies must follow the architectural direction established by PKS-020.

The following rules are mandatory.

---

## 🧭 Architectural Layer Mapping

Mock:ctl maps its repository modules to the system architecture as follows:

| Architectural Layer | Repository Module | Primary Responsibility |
|---|---|---|
| Presentation | `cmd/` | CLI and executable entry points |
| Application Core | `internal/app/` | Application orchestration |
| Domain Services | `internal/project/` | Project lifecycle and project domain |
| Domain Services | `internal/spec/` | Specification processing |
| Domain Services | `internal/generator/` | Mock generation |
| Domain Services | `internal/data/` | Data generation |
| Infrastructure Services | `internal/runtime/` | Runtime execution |
| Infrastructure Services | `internal/config/` | Configuration management |
| Infrastructure Services | `internal/storage/` | Persistence and filesystem access |
| Shared Internal | `internal/shared/` | Genuinely cross-cutting internal abstractions |

This mapping must remain aligned with PKS-020.

A module must not silently change architectural ownership merely because a particular implementation makes that convenient.

---

# 🔄 Dependency Direction

The preferred dependency direction is:

```text
cmd/
  ↓
internal/app/
  ↓
Domain Services
  ↓
Infrastructure Services
  ↓
Operating System / File System / External Runtime
```

Expanded:

```text
┌───────────────┐
                         │     cmd/      │
                         │ Presentation  │
                         └───────┬───────┘
                                 │
                                 ▼
                         ┌───────────────┐
                         │  internal/    │
                         │     app/      │
                         │ Application   │
                         │     Core      │
                         └───────┬───────┘
                                 │
               ┌─────────────────┼─────────────────┐
               │                 │                 │
               ▼                 ▼                 ▼
        ┌────────────┐    ┌────────────┐    ┌────────────┐
        │  project/  │    │   spec/    │    │ generator/ │
        │   Domain   │    │   Domain   │    │   Domain   │
        └────────────┘    └─────┬──────┘    └─────┬──────┘
                                │                 │
                                ▼                 │
                         ┌────────────┐           │
                         │   data/    │◄──────────┘
                         │   Domain   │
                         └────────────┘

                         Infrastructure
               ┌────────────────┼────────────────┐
               ▼                ▼                ▼
        ┌────────────┐   ┌────────────┐   ┌────────────┐
        │ runtime/   │   │  config/   │   │ storage/   │
        │            │   │            │   │            │
        └────────────┘   └────────────┘   └────────────┘
```
This diagram represents architectural responsibility and preferred dependency direction.

It does not imply that every module must directly depend on every module below it.


---

## 🚦 Application Core Boundary

internal/app/ is the primary application orchestration boundary.

Application workflows should pass through the Application Core rather than allowing presentation code to directly orchestrate internal subsystems.

Preferred:

```text
CLI
 ↓
Application Core
 ↓
Subsystems
```

Not preferred:

```text
CLI
 ├── Project Manager
 ├── Specification Engine
 ├── Runtime Engine
 └── Storage
```

The Application Core exists to coordinate these operations.

However, `internal/app/` must not become a monolithic implementation layer.

It should orchestrate responsibilities owned by other modules rather than absorbing those responsibilities.


---

## 🧱 Module Ownership Rules

Each module has one primary owner responsibility.

| Module | Owns |
|---|---|
| `cmd/` | Executable and presentation entry points |
| `app/` | Application orchestration |
| `project/` | Project lifecycle and project domain |
| `spec/` | Specification processing |
| `generator/` | Mock generation |
| `data/` | Fake-data generation |
| `runtime/` | Runtime execution |
| `config/` | Configuration resolution |
| `storage/` | Persistence and filesystem access |
| `shared/` | Genuinely cross-cutting internal abstractions |


New functionality must first be assigned to an existing ownership boundary before a new package is considered.


---

## 🚫 No Responsibility Leakage

A module must not implement another module's primary responsibility merely because it has access to the required data.

Examples of prohibited responsibility leakage:

```text
cmd/
 └── implements mock-generation rules

runtime/
 └── parses OpenAPI documents

storage/
 └── applies project business rules

config/
 └── decides runtime behavior

generator/
 └── directly parses external OpenAPI structures

project/
 └── starts the HTTP server
```

The correct approach is to call the module that owns the responsibility.


---

## 🔐 Application Core as the Coordination Boundary

The Application Core may coordinate multiple modules.

For example:

```text
User Command
     ↓
Application Core
     ↓
Project Manager
     ↓
Specification Engine
     ↓
Mock Generation Engine
     ↓
Data Generation Engine
     ↓
Runtime Engine
```

Each module performs its own responsibility.

The Application Core determines the application workflow.

It does not duplicate the implementation of those modules.


---

## 🔁 Module Communication Rules

Modules should communicate through explicit interfaces and well-defined data structures where abstraction is beneficial.

The following rules apply:

1. Prefer explicit dependencies.


2. Avoid hidden global state.


3. Avoid circular dependencies.


4. Avoid direct access to another module's internal implementation.


5. Prefer domain-owned types over third-party types at architectural boundaries.


6. Keep infrastructure details behind infrastructure boundaries.


7. Keep presentation details outside domain modules.


8. Keep application orchestration outside infrastructure modules.


9. Avoid unnecessary abstraction layers.


10. Introduce an interface when it represents a meaningful architectural boundary, not merely to satisfy a pattern.




---

## 🔄 Circular Dependency Rule

Circular dependencies between modules are prohibited.

Invalid:

```text
project/
   ↓
spec/
   ↓
project/
```

Invalid:

```text
generator/
   ↓
runtime/
   ↓
generator/
```

Invalid:

```text
config/
   ↓
app/
   ↓
config/
```

When a circular dependency appears, the architecture must be reconsidered.

Possible solutions include:

- Moving a shared abstraction to the correct owner.
- Introducing a narrow interface.
- Moving orchestration into `internal/app/`.
- Introducing a stable shared type where genuinely appropriate.
- Separating responsibilities that were incorrectly coupled.


The solution must preserve ownership clarity.


---

## 🚫 Runtime → Project Manager Bypass

The Runtime Engine must not directly manipulate Project Manager internals.

Forbidden:

```text
runtime/
    ↓
project/
```

when the purpose is to bypass the Application Core and directly alter project lifecycle or project state.

The Runtime Engine may consume runtime definitions or state explicitly provided to it through an approved application boundary.

Preferred:

```text
Application Core
      ↓
Project Manager
      ↓
Runtime Definition
      ↓
Runtime Engine
```

The Runtime Engine must remain focused on executing the runtime rather than managing the project itself.


---

## 📜 Specification Boundary

The Specification Engine owns the conversion from external API specifications into Mock:ctl's internal representation.

The preferred flow is:

```text
OpenAPI Document
      ↓
External Parser
      ↓
Parser Adapter
      ↓
Internal Specification Model
      ↓
Application / Domain Modules
```

The rest of the application should not become coupled to kin-openapi types.

This creates a stable internal specification boundary.

If the parser implementation changes in the future, the impact should remain concentrated within the Specification Engine.


---

## 🧩 Parser Abstraction

The external parser is an implementation detail.

The architecture must therefore avoid:

```text
generator/
   ↓
kin-openapi types

runtime/
   ↓
kin-openapi types

project/
   ↓
kin-openapi types
```

Preferred:

```text
spec/
 ├── parser adapter
 ├── validation
 └── internal specification model
             ↓
      Other application modules
```

The internal specification model is the architectural contract.

The external parser is not.


---

## 🎲 Data Generation Boundary

The Data Generation Engine owns fake-data generation.

The approved implementation technology is gofakeit.

The architecture must prevent the rest of the application from becoming tightly coupled to the external library.

Preferred:

```text
generator/
     ↓
data/
     ↓
internal data-generation abstraction
     ↓
gofakeit
```

Not:

```text
generator/
     ↓
gofakeit
```

The purpose of this boundary is to protect the application from external-library implementation details and preserve future replacement flexibility.


---

## 🌐 HTTP Runtime Boundary

The Runtime Engine owns HTTP runtime behavior.

The approved HTTP foundation is:

```text
Go net/http
     +
Chi
```

The HTTP implementation must remain inside the Runtime Engine boundary.

Preferred:

```text
Application Core
      ↓
Runtime Engine
      ↓
HTTP Abstraction
      ↓
Chi / net/http
```

Other domain modules should not directly manipulate HTTP router internals.


---

## ⚙️ Configuration Boundary

Configuration is resolved centrally by `internal/config/`.

The application must not allow every subsystem to implement its own interpretation of configuration precedence.

The approved precedence model is:

```text
Built-in Defaults
        ↓
Global Configuration
        ↓
Project Configuration
        ↓
Environment Overrides
        ↓
Command-Line Arguments
```

Higher-priority configuration overrides lower-priority configuration.

The resolved result should be passed to consuming components through explicit application boundaries.


---

## 💾 Storage Boundary

Filesystem and persistence operations belong to `internal/storage/`.

Business modules should not scatter raw filesystem operations throughout their implementation.

Avoid:

```text
project/
 ├── os.ReadFile(...)
 ├── os.WriteFile(...)
 └── filepath operations
```

when those operations represent project persistence responsibilities.

Prefer:

```text
project/
     ↓
storage abstraction
     ↓
storage/
     ↓
filesystem
```

This keeps persistence concerns isolated from domain rules.


---

## 🧠 Shared Module Rules

`internal/shared/` exists only for genuinely shared internal functionality.

Before placing code in `shared/`, answer:

1. Is this used by multiple independent modules?


2. Does it have no stronger ownership elsewhere?


3. Is the abstraction stable enough to be shared?


4. Does moving it here improve clarity rather than hide responsibility?



If the answer is no, the code belongs somewhere else.


---

## 🚫 Shared Package Anti-Pattern

The following pattern is prohibited:

```text
internal/shared/
├── helpers.go
├── utils.go
├── common.go
├── misc.go
├── constants.go
└── everything_else.go
```

Generic convenience is not sufficient architectural justification.

Avoid creating shared/ utilities simply because moving code into the correct owner requires more thought.


---

## 🖥️ Platform Architecture Boundary

Mock:ctl's shared application logic must remain platform-independent.

The approved platform architecture supports:

```text
Shared Go Core
                      │
          ┌───────────┴───────────┐
          │                       │
          ▼                       ▼
      Desktop                  Android
      Flutter                  Flutter
          │                       │
        FFI                     FFI
          │                       │
          └───────────┬───────────┘
                      │
                 Shared Core
```

Desktop and Android presentation layers must remain thin clients over the shared Go core.

Platform-specific UI behavior belongs to the platform layer.

Shared application behavior belongs to the Go core.


---

## 🖥️ Desktop Boundary

The desktop application uses Flutter as the presentation framework and embeds the shared Go core.

The architectural relationship is:

```text
Flutter Desktop
      ↓
Dart FFI
      ↓
Shared Go Core
```

Desktop UI code must not duplicate core Mock:ctl business logic.

The desktop layer is responsible for presentation and platform-specific concerns.


---

## 📱 Android Boundary

Android follows the same shared-core principle:

```text
Flutter Android
      ↓
Dart FFI
      ↓
Shared Go Core
```

Android-specific UI and platform integration remain outside the shared domain implementation.

The Go core remains platform-independent.


---

## 🔌 Plugin Readiness Boundary

Mock:ctl is designed to remain plugin-ready without making plugins a Version 1.0 implementation dependency.

Therefore:

Core modules must not assume plugin implementations exist.

Plugin-specific logic must not leak into core modules.

The repository must not introduce speculative plugin directories without an approved architectural requirement.

Future plugin implementation must integrate through an explicit plugin boundary.

Plugin execution must not require rewriting the core architecture.


The architecture is therefore:

```text
Core Application
      │
      ├── Built-in Capabilities
      │
      └── Future Plugin Boundary
                │
                ▼
          Plugin Runtime
```

The exact plugin repository layout is intentionally deferred until the plugin architecture requires it.


---

## 🧪 Test Placement Rules

Testing follows two repository placement rules.

Package-Level Tests

Tests that directly validate a package should normally remain beside that package.

Example:

```text
internal/spec/
├── parser.go
├── model.go
└── parser_test.go
```

Repository-Level Tests

Tests requiring repository-wide resources, integration fixtures, or end-to-end infrastructure may use:

```text
test/
├── fixtures/
├── integration/
└── e2e/
```

The exact structure will be refined by PKS-029 — Testing Strategy.

This document defines ownership, not the complete testing methodology.


---

## 🧱 Package Creation Rules

A new package should only be introduced when at least one of the following is true:

- It represents a distinct architectural responsibility.
- It establishes a meaningful dependency boundary.
- It isolates a replaceable implementation.
- It significantly improves maintainability.
- It prevents responsibility leakage.


Do not create packages merely to reduce file size.

Do not create packages merely because a pattern suggests doing so.

Do not create packages solely for aesthetic organization.


---

## 🏗️ Module Evolution Rule

Existing modules should be extended before new modules are created when the new functionality clearly belongs to an existing responsibility.

For example:

Specification validation

belongs inside: `internal/spec/`

rather than creating: `internal/validation/`

unless validation becomes a distinct architectural responsibility.

Similarly:

Runtime middleware

belongs within: `internal/runtime/`

unless a future architectural decision establishes an independent middleware subsystem.


---

## 🔒 Internal Package Visibility

The `internal/` boundary is intentional.

Mock:ctl's core implementation should not be treated as a reusable public Go library by default.

A package should become externally importable only when a separate approved architectural decision establishes that requirement.

This protects the project from prematurely committing to public APIs.


---

## 🔗 External Dependency Isolation

External libraries must be isolated whenever they represent a replaceable implementation detail or an architectural boundary.

Current important boundaries include:

```text
OpenAPI
    ↓
kin-openapi
    ↓
spec abstraction

Fake Data
    ↓
gofakeit
    ↓
data abstraction

HTTP Runtime
    ↓
Chi / net/http
    ↓
runtime abstraction
```

External library types should not unnecessarily propagate across the entire application.

The goal is not to eliminate third-party dependencies.

The goal is to contain their architectural impact.


---

## 📋 Dependency Decision Rule

Before adding a new dependency, the implementation must establish:

1. Why the dependency is required.


2. Which module owns the dependency.


3. Whether the dependency should be hidden behind an abstraction.


4. Whether the dependency introduces platform coupling.


5. Whether the dependency duplicates an existing capability.


6. Whether the dependency conflicts with an approved Engineering Decision.



New dependencies must not silently bypass the Technology Stack or Engineering Decision Log.


---

# 🤖 AI-Assisted Development Rules

Mock:ctl is explicitly designed to support AI-assisted development.

The repository architecture must therefore remain understandable to both human developers and AI coding agents.

AI agents must treat this document and the other approved Project Knowledge System documents as architectural constraints rather than optional suggestions.

---

## 📖 Documentation-First Implementation

Before implementing a feature, an AI coding agent must determine:

1. Which approved requirement defines the feature.
2. Which architectural document defines its structure.
3. Which module owns the responsibility.
4. Which existing interfaces or abstractions should be reused.
5. Which Engineering Decisions constrain the implementation.
6. Whether the feature requires a new architectural decision.

Implementation must not redefine an approved architectural decision without explicit approval.

---

# 🧭 Module Selection Rule for AI Agents

When an AI agent receives an implementation request, it must first identify the correct module owner.

Use the following decision model:

```text
Does the functionality belong to presentation?
        │
        ├── Yes → cmd/
        │
        └── No
             ↓
Does it orchestrate an application workflow?
        │
        ├── Yes → internal/app/
        │
        └── No
             ↓
Does it belong to project lifecycle/domain?
        │
        ├── Yes → internal/project/
        │
        └── No
             ↓
Does it process API specifications?
        │
        ├── Yes → internal/spec/
        │
        └── No
             ↓
Does it generate mocks?
        │
        ├── Yes → internal/generator/
        │
        └── No
             ↓
Does it generate fake data?
        │
        ├── Yes → internal/data/
        │
        └── No
             ↓
Does it execute the runtime?
        │
        ├── Yes → internal/runtime/
        │
        └── No
             ↓
Does it manage configuration?
        │
        ├── Yes → internal/config/
        │
        └── No
             ↓
Does it manage persistence/filesystem access?
        │
        ├── Yes → internal/storage/
        │
        └── No
             ↓
Is it genuinely cross-cutting?
        │
        ├── Yes → internal/shared/
        │
        └── No → Re-evaluate the architecture
```

The purpose of this rule is to prevent AI agents from creating arbitrary packages when an existing architectural boundary already owns the responsibility.


---

## 🚫 AI Must Not Invent Architecture

AI coding agents must not introduce:

- New top-level directories.
- New architectural layers.
- New major modules.
- New public Go packages.
- Alternative dependency structures.
- Duplicate implementations of existing subsystems.
- Unapproved frameworks.
- Unapproved infrastructure technologies.


unless the required change has been explicitly approved through the project's documentation and decision process.

If implementation genuinely requires an architectural change, the agent must stop and surface the conflict rather than silently changing the architecture.


---

## 🧩 Existing Module First

Before creating a new module, an AI agent must inspect existing modules and determine whether the requested functionality already belongs to one of them.

For example:

OpenAPI validation

belongs to: `internal/spec/`

not: `internal/validation/`

Likewise:

HTTP middleware

belongs to: `internal/runtime/`

not automatically: `internal/middleware/`

The existence of a new concept does not automatically justify the creation of a new package.


---

## 🔍 AI Repository Navigation

AI agents should use the repository structure as the first architectural signal.

Before modifying code, the agent should inspect:

```text
docs/
internal/
cmd/
test/
assets/
scripts/
go.mod
```

and identify relevant existing implementations before creating new files.

The agent should prefer modifying an existing implementation over creating a parallel implementation.


---

## 🧱 AI Implementation Boundary

AI-generated code must respect module ownership.

For example:

Specification Logic: `internal/spec/`

Mock Generation: `internal/generator/`

Fake Data: `internal/data/`

Runtime: `internal/runtime/`

Configuration: `internal/config/`

Persistence: `internal/storage/`

An AI agent must not place implementation code in a convenient location that violates these ownership boundaries.


---

## 🔗 AI Dependency Rules

When adding or modifying dependencies, an AI agent must verify:

```text
New Dependency
      ↓
Owning Module
      ↓
Architectural Boundary
      ↓
Approved Technology Stack
      ↓
Approved Engineering Decision
```

If the dependency is not consistent with the approved Technology Stack or Engineering Decision Log, the agent must not silently introduce it.


---

## 🧠 AI Abstraction Rules

AI-generated code must avoid unnecessary abstractions.

Do not create:

- Interfaces with only one meaningless implementation.
- Generic factories without an architectural requirement.
- Wrapper types that provide no boundary.
- Utility packages that only relocate code.
- Generic service layers that duplicate the Application Core.
- Framework-shaped abstractions that are not required by Mock:ctl.


An abstraction is justified when it provides a meaningful architectural boundary, replaceability, testability, or ownership separation.


---

## 🔄 Refactoring Rules

AI agents may refactor implementation when required to preserve the architecture.

However, refactoring must not silently alter:

- Module ownership.
- Public behavior.
- Approved interfaces.
- Configuration precedence.
- Runtime boundaries.
- Repository structure.
- Technology choices.
- Architectural decisions.


Large architectural refactors require explicit approval.


---

## 🧪 AI Verification Before Completion

Before declaring an implementation complete, an AI coding agent should verify:

- ✓ Correct module ownership
- ✓ Correct dependency direction
- ✓ No circular dependencies
- ✓ No responsibility leakage
- ✓ No unnecessary new packages
- ✓ No unauthorized dependencies
- ✓ No direct external-library leakage across boundaries
- ✓ Existing architecture remains intact
- ✓ Existing tests remain valid
- ✓ New behavior has appropriate tests
- ✓ Documentation remains consistent

The implementation is not considered complete merely because the code compiles.


---

## 🗺️ Repository-to-System Mapping

The following mapping is the authoritative repository interpretation of the system architecture:

```text
Mock:ctl
│
├── cmd/
│   └── Presentation / Entry Points
│
├── internal/
│   │
│   ├── app/
│   │   └── Application Core
│   │
│   ├── project/
│   │   └── Project Manager
│   │
│   ├── spec/
│   │   └── Specification Engine
│   │
│   ├── generator/
│   │   └── Mock Generation Engine
│   │
│   ├── data/
│   │   └── Data Generation Engine
│   │
│   ├── runtime/
│   │   └── Runtime Engine
│   │
│   ├── config/
│   │   └── Configuration Manager
│   │
│   ├── storage/
│   │   └── Storage Layer
│   │
│   └── shared/
│       └── Shared Internal Abstractions
│
├── docs/
│   └── Project Knowledge System
│
├── scripts/
│   └── Development / Automation Support
│
├── test/
│   └── Repository-Level Test Resources
│
├── assets/
│   └── Project-Controlled Non-Source Assets
│
└── go.mod
    └── Single Go Module
```

This mapping should remain stable unless a later approved architecture document explicitly changes it.


---

## 📊 Responsibility Matrix

| Concern | Owning Module | Architectural Layer |
|---|---|---|
| CLI entry point | `cmd/` | Presentation |
| Application workflow | `internal/app/` | Application Core |
| Project lifecycle | `internal/project/` | Domain |
| OpenAPI processing | `internal/spec/` | Domain |
| Mock generation | `internal/generator/` | Domain |
| Fake-data generation | `internal/data/` | Domain |
| HTTP runtime | `internal/runtime/` | Infrastructure |
| Configuration | `internal/config/` | Infrastructure |
| Persistence | `internal/storage/` | Infrastructure |
| Cross-cutting internal abstractions | `internal/shared/` | Shared Internal |
| Documentation | `docs/` | Project Knowledge |
| Automation scripts | `scripts/` | Development Support |
| Repository-level test resources | `test/` | Testing Support |
| Non-source project assets | `assets/` | Project Resources |



---

## 🔍 Traceability to Engineering Decisions

The repository architecture is based on the following approved Engineering Decisions.

| Decision | Architectural Impact |
|---|---|
| EDL-001 | Go is the primary programming language |
| EDL-002 | Go is the primary application language |
| EDL-004 | Establishes the repository directory structure |
| EDL-005 | Establishes the internal-first repository architecture |
| EDL-006 | Establishes the single Go module |
| EDL-007 | Establishes standard Go package conventions |
| EDL-008 | Establishes the core module structure |
| EDL-009 | Establishes kin-openapi as the OpenAPI parser |
| EDL-010 | Establishes gofakeit as the fake-data provider |
| EDL-011 | Establishes the HTTP foundation |
| EDL-040 | Establishes Embedded Go Core (Dart FFI) for desktop |
| EDL-041 | Establishes Flutter as the desktop frontend |
| EDL-042 | Establishes the shared application core architecture |
| EDL-043 | Establishes Flutter as the Android framework |
| EDL-044 | Establishes Embedded Go Core for Android |
| EDL-045 | Establishes single shared backend strategy |
| EDL-046 | Establishes WebAssembly (WASM) plugin format |
| EDL-047 | Establishes wazero plugin runtime |
| EDL-048 | Establishes plugin distribution strategy |
| EDL-049 | Establishes local-first cloud support strategy |


These references establish traceability between repository architecture and engineering decisions.

Where a later Engineering Decision supersedes or modifies an earlier decision, the later approved decision takes precedence.


---

## 🔗 Relationship With PKS-020

PKS-020 — System Architecture defines the system-level architecture.

PKS-022 — Repository & Module Architecture translates that architecture into repository and module boundaries.

The relationship is:

```text
PKS-020
System Architecture
       ↓
PKS-022
Repository & Module Architecture
       ↓
PKS-023
Data Flow Architecture
       ↓
PKS-024
Component Architecture
       ↓
PKS-025
Software Design Document
```

PKS-022 must not redefine the system architecture established by PKS-020.

It provides the implementation-oriented repository interpretation of that architecture.


---

## 🔗 Relationship With PKS-021

PKS-021 — Technology Stack defines which technologies Mock:ctl uses.

PKS-022 defines where those technologies are allowed to live within the repository.

Examples:

```text
kin-openapi
    ↓
internal/spec/

gofakeit
    ↓
internal/data/

Chi / net/http
    ↓
internal/runtime/

Flutter
    ↓
Platform Presentation Layer

Shared Go Core
    ↓
Shared Application Logic
```

PKS-022 must not independently select replacement technologies.

Technology selection remains governed by PKS-021 and the Engineering Decision Log.


---

## 🔗 Relationship With PKS-023

PKS-023 will define how data and requests move through the architecture.

PKS-022 provides the module boundaries through which those flows operate.

Therefore PKS-023 must use the module names and ownership boundaries defined here.

Examples:

```text
cmd/
internal/app/
internal/spec/
internal/generator/
internal/data/
internal/runtime/
internal/config/
internal/storage/
```

must retain the same architectural meaning across subsequent documentation.


---

## 🔗 Relationship With PKS-024

PKS-024 will define component-level architecture.

Components must be placed within the module that owns their responsibility.

A component should not cross module boundaries merely because its implementation is convenient there.

PKS-024 must therefore treat PKS-022 as the repository-level ownership foundation.


---

## 🔗 Relationship With PKS-025

PKS-025 — Software Design Document (Master SDD) will consolidate detailed software design.

PKS-025 must inherit:

- Repository boundaries.
- Module ownership.
- Dependency direction.
- External dependency boundaries.
- Platform boundaries.
- Shared-core rules.

defined by PKS-022.

PKS-025 may refine implementation details but must not silently contradict this document.


---

## 🔗 Relationship With PKS-026

PKS-026 — Database Design will define persistent data structures and database-specific architecture if required by the approved product design.

Any database or persistence implementation must respect:

```text
Application / Domain
        ↓
Storage Boundary
        ↓
Persistence Implementation
```

Database implementation details must not leak into unrelated domain modules.

If a future persistence strategy changes, the repository architecture should preserve the storage/ responsibility boundary wherever practical.


---

## 🔗 Relationship With PKS-027

PKS-027 — API Design will define API contracts and interfaces.

API-facing implementation must respect the distinction between:

- API contract.
- Application orchestration.
- Domain behavior.
- Runtime implementation.


The repository architecture does not allow API concerns to become a substitute for the Application Core.


---

## 🔗 Relationship With PKS-028

PKS-028 — Coding Standards will define implementation-level coding conventions.

PKS-028 must operate within the architectural boundaries established by PKS-022.

Coding style must not override module ownership.

A stylistically clean implementation can still be architecturally incorrect.


---

## 🔗 Relationship With PKS-029

PKS-029 — Testing Strategy will define the testing methodology.

PKS-029 must use the repository boundaries defined here to determine:

- Unit-test ownership.
- Integration-test boundaries.
- End-to-end test placement.
- Test fixtures.
- Test resources.

This document establishes placement principles, while PKS-029 defines the detailed strategy.


---

## 🔗 Relationship With PKS-030

PKS-030 — Deployment Architecture will define how Mock:ctl is packaged, distributed, and deployed.

Deployment-specific files or automation may extend the repository structure where necessary.

However, deployment concerns must not introduce unnecessary coupling into the application modules.


---

## 📐 Repository Stability Rules

The following rules apply to future repository changes.

Rule 1 — Preserve Ownership

Existing module ownership should remain stable.

Rule 2 — Prefer Extension Over Fragmentation

Extend an existing module when the responsibility already belongs there.

Rule 3 — Avoid Premature Generalization

Do not create abstractions for hypothetical future requirements.

Rule 4 — Protect Boundaries

Do not bypass established architectural boundaries for implementation convenience.

Rule 5 — Document Architectural Changes

Any significant repository or module architecture change must be reflected in the appropriate PKS document.

Rule 6 — Update Traceability

When an Engineering Decision changes repository architecture, the affected documentation must be updated.

Rule 7 — Keep AI Guidance Consistent

Repository structure and module ownership must remain understandable to AI coding agents.


---

## 🚨 Architectural Violation Examples

The following are considered architectural violations.

Violation 1 — Business Logic in CLI

```texr
cmd/
 └── implements generation algorithm
```

Violation 2 — Parser Leakage

```text
generator/
 └── directly consumes kin-openapi objects
```

Violation 3 — Runtime Owning Project Lifecycle

```text
runtime/
 └── creates / deletes / manages projects
```

Violation 4 — Storage Owning Business Rules

```text
storage/
 └── decides mock-generation behavior
```

Violation 5 — Configuration Duplication

```text
project/
 └── custom configuration precedence

runtime/
 └── different configuration precedence
```

Violation 6 — Shared Package Dumping Ground

```text
shared/
 └── unrelated application features
```

Violation 7 — Unapproved Top-Level Module

```text
mockctl/
├── services/
├── managers/
├── engines/
└── ...
```

when these directories duplicate responsibilities already defined by internal/.


---

## ✅ Architecture Validation Checklist

Before approving a repository architecture change, verify:

- [ ] Does the change have a clearly defined owner?
- [ ] Does it fit an existing module?
- [ ] If not, is a new module genuinely justified?
- [ ] Does dependency direction remain valid?
- [ ] Does the change introduce circular dependencies?
- [ ] Does it leak external-library types?
- [ ] Does it bypass the Application Core?
- [ ] Does it duplicate an existing responsibility?
- [ ] Does it introduce platform coupling?
- [ ] Does it affect the shared Go core?
- [ ] Does it conflict with PKS-020?
- [ ] Does it conflict with PKS-021?
- [ ] Does it conflict with an approved EDL?
- [ ] Does the relevant documentation require an update?
- [ ] Can an AI coding agent understand the new boundary?

A change should not be considered architecturally complete until these questions have been addressed.


---

## 🧾 Repository Architecture Contract

The following statements are considered mandatory architectural rules:

1. Mock:ctl uses an internal-first repository architecture.


2. Mock:ctl uses a single Go module.


3. `cmd/` contains executable entry points and presentation concerns.


4. `internal/app/` owns application orchestration.


5. `internal/project/` owns project lifecycle and project domain behavior.


6. `internal/spec/` owns specification processing.


7. `internal/generator/` owns mock generation.


8. `internal/data/` owns fake-data generation.


9. `internal/runtime/` owns runtime execution.


10. `internal/config/` owns configuration resolution.


11. `internal/storage/` owns persistence and filesystem access.


12. `internal/shared/` contains only genuinely shared internal abstractions.


13. External technology dependencies must remain appropriately isolated.


14. The Runtime Engine must not bypass the Application Core to manipulate Project Manager internals.


15. Domain modules must not become dependent on presentation concerns.


16. Infrastructure modules must not absorb domain responsibilities.


17. Platform presentation layers must not duplicate shared core business logic.


18. New top-level architecture must not be introduced without approval.


19. New modules must have explicit architectural justification.


20. AI coding agents must follow these boundaries.


21. Significant architectural changes must be reflected in the appropriate documentation.


22. PKS-022 must remain consistent with PKS-020, PKS-021, and the approved Engineering Decision Log.




---

## 📝 Implementation Guidance

This document intentionally defines architecture and ownership, not every future implementation detail.

A developer or AI coding agent should use the following sequence:

```text
Requirement
    ↓
Relevant PKS Document
    ↓
Architectural Responsibility
    ↓
Owning Module
    ↓
Existing Implementation
    ↓
Minimal Required Change
    ↓
Tests
    ↓
Documentation Update if Required
```

The goal is to prevent implementation from becoming the source of architecture.

Architecture is established first.

Implementation follows it.


---

## 🔄 Change Management

Any change to the repository or module architecture must be evaluated against:

PKS-020 — System Architecture.

PKS-021 — Technology Stack.

Engineering Decision Log.

Existing module ownership.

Future dependent documents.


A change that affects module ownership, dependency direction, repository structure, or platform boundaries is an architectural change.

Architectural changes must not be introduced merely as implementation details.


---

# 🔗 Related Documents

This document is derived from and must remain consistent with:

- PKS-000 — Repository Blueprint.
- PKS-001 — Documentation Philosophy.
- PKS-002 — Documentation Style Guide.
- PKS-003 — Project Knowledge System.
- PKS-004 — Documentation Index.
- PKS-010 — Vision.
- PKS-011 — User Personas.
- PKS-012 — User Journey.
- PKS-013 — Functional Requirements.
- PKS-014 — Non-Functional Requirements.
- PKS-015 — MVP Definition.
- PKS-016 — Product Requirements Document.
- PKS-017 — Product Roadmap.
- PKS-020 — System Architecture.
- PKS-021 — Technology Stack.
- Engineering Decision Log.



---

## 📌 Document Authority

PKS-022 is the authoritative document for:

- Repository structure.
- Repository-level module organization.
- Module ownership.
- Repository dependency boundaries.
- Internal package responsibilities.
- Repository-level AI development constraints.


If another document conflicts with this document regarding repository/module architecture, the conflict must be resolved through the project's documentation and decision process rather than by silently choosing one interpretation.


---

## 🏁 Final Summary

Mock:ctl uses a deliberately simple repository architecture built around a single Go module and an internal-first package structure.

The architecture separates:

```text
Presentation
    ↓
Application Core
    ↓
Domain Services
    ↓
Infrastructure Services
```

while keeping external technologies behind appropriate boundaries.

The core repository modules are:

```text
cmd/
internal/app/
internal/project/
internal/spec/
internal/generator/
internal/data/
internal/runtime/
internal/config/
internal/storage/
internal/shared/
```

This structure provides a stable foundation for the next engineering documents and implementation work.

The repository is intentionally designed to be:

- Modular.
- Understandable.
- Maintainable.
- Testable.
- Platform-independent at the core.
- Resistant to unnecessary coupling.
- Suitable for AI-assisted development.
- Capable of evolving without premature architectural complexity.



---

# 📜 Revision History

| Version | Date | Description |
|----------|------------|----------------------------------------------|
| 1.0 | 2026-08-09 | Initial approved release |

---

# ✅ Approval Checklist

- Executive summary completed
- Architecture philosophy documented
- Architectural foundation defined
- Core module structure documented
- Module ownership rules established
- Dependency direction rules established
- External dependency boundaries defined
- Platform boundaries defined
- AI development rules established
- Responsibility matrix completed
- Traceability to Engineering Decisions verified
- Document relationships established

---

# 📌 Conclusion

The Repository & Module Architecture translates the system-level boundaries into a concrete repository structure for Mock:ctl.

It establishes the internal-first module architecture, precise ownership rules, dependency directions, and specific package responsibilities that all implementations must follow.

This document serves as the authoritative structural reference for the repository. Any proposed change to module boundaries, directory structures, or dependency rules should first be evaluated against the principles documented here before code is written.

With the approval of this document, the project has established its codebase foundation and is ready for detailed component design and implementation.

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** Completed

**Architecture Status:** ✅ Established

**Next Document:** **PKS-023 — Data Flow Architecture**