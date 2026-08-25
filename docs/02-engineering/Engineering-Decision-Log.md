# Engineering Decision Log (EDL)

> **Project:** Mock:ctl
> 
> Document ID: EDL
>
> **Version:** 1.0
> 
> Status: Active
> 
> **Owner:** Upen Tudu
>
> **Authors:** Upen Tudu & ChatGPT
>  
> Authority: Engineering Decision Log
> 
> Decision Range: EDL-001 → EDL-054
> 
> Document Type: Engineering Governance / Decision Record
> 
> Primary Related Document: PKS-021 — Technology Stack
>
> **Created:** 2026-08-06
>
> **Last Updated:** 2026-08-06
>
> **Category:** Engineering

---

## Summary

The Engineering Decision Log (EDL) records the major engineering decisions that define the technical foundation, architectural direction, development standards, tooling, runtime strategy, testing approach, release process, and platform strategy of Mock:ctl.

Each Engineering Decision represents an intentional technical choice made for the project and serves as the authoritative record of **what was decided, why it was decided, and what alternatives were considered** where applicable.

The EDL is complementary to the Project Knowledge System (PKS).

The PKS documents describe the project's requirements, architecture, technology stack, design, and implementation guidance, while the EDL preserves the engineering rationale behind significant technical decisions.

---

## Purpose

The purpose of this Engineering Decision Log is to:

- Record significant engineering decisions made during the development of Mock:ctl.
- Preserve the reasoning behind technology and architecture choices.
- Document important alternatives considered before making a decision.
- Establish an authoritative engineering baseline.
- Prevent previously resolved engineering decisions from being repeatedly reconsidered without new evidence.
- Provide traceability between the technology stack and the engineering rationale behind it.
- Support consistent implementation across CLI, Desktop, Android, and future interfaces.
- Provide reliable context for AI-assisted development.
- Make future technology replacement and architectural evolution easier to evaluate.

The EDL is intended to preserve **decision history**, not merely a list of technologies.

---

## Scope

This document covers significant engineering decisions concerning:

- Programming language and runtime
- Core architecture
- Repository and module structure
- Build and dependency management
- CLI architecture
- Configuration
- Logging
- Specification parsing
- Fake data generation
- HTTP architecture
- Testing
- Code quality
- CI/CD
- Release management
- Desktop architecture
- Android architecture
- Plugin architecture
- Cross-platform architecture
- Future cloud strategy
- Technology-stack governance

The EDL does not replace:

- Product requirements
- Functional requirements
- Non-functional requirements
- System architecture
- Software design documentation
- Database design
- API design
- Coding standards
- Testing strategy
- Deployment architecture

Those concerns are documented in their respective PKS documents.

---

## Engineering Decision Philosophy

Engineering decisions in Mock:ctl should favor solutions that are:

1. **Simple**
2. **Maintainable**
3. **Portable**
4. **Reliable**
5. **Testable**
6. **AI-assisted-development friendly**
7. **Consistent across platforms**
8. **Appropriate for the project's current scale**
9. **Capable of evolving without unnecessary complexity**

The project should avoid introducing complexity merely because a more sophisticated solution exists.

Technology and architecture should serve the product rather than becoming the product.

---

## Decision Principles

### Official First

When the Go Standard Library or official Go Toolchain adequately satisfies a requirement, it should generally be preferred over third-party dependencies.

Third-party dependencies should be introduced when they provide a clear engineering advantage that cannot reasonably be achieved through the standard library.

### Single Source of Truth

Important engineering behavior should have a clearly defined authoritative source.

Examples include:

- The Go Toolchain for official builds.
- The internal specification model for application-level specification handling.
- `gofmt` for Go source formatting.
- The Engineering Decision Log for recorded engineering decisions.

### Separation of Concerns

Presentation, transport, application, domain, and infrastructure responsibilities should remain appropriately separated.

Business logic should not become coupled to a specific user interface or transport mechanism.

### Replaceability

External technologies that may reasonably need replacement should be isolated behind project-owned abstractions where appropriate.

This protects the core application from unnecessary vendor or library coupling.

### Determinism

Builds, tests, generated output, and other engineering workflows should favor deterministic and reproducible behavior wherever practical.

### Human Maintainability

Engineering decisions should remain understandable to human developers.

AI-assisted development should improve development efficiency without becoming a substitute for clear architecture or engineering judgment.

---

## Decision Status

Each decision in this document has an explicit status.

### Approved

The decision is currently accepted as an authoritative engineering decision.

### Proposed

The decision has been documented for consideration but has not yet become authoritative.

### Deprecated

The decision was previously accepted but has been superseded by another decision.

### Rejected

The decision was explicitly considered and rejected.

The current EDL decision set consists of approved decisions unless otherwise stated.

---

## Replacement Risk

Replacement Risk describes the estimated engineering impact of replacing the selected technology or decision.

### None

Replacement should have little or no meaningful architectural impact.

### Very Low

Replacement should be relatively straightforward and have limited impact.

### Low

Replacement requires some engineering work but should remain manageable.

### Medium

Replacement would require meaningful architectural or implementation changes.

### High

Replacement would have substantial impact on the system.

Replacement Risk is not a prediction of whether a technology will actually be replaced.

---

# Engineering Decisions

## EDL-001 — Programming Language

**Decision:**  
Programming Language = Go

**Status:**  
✅ Approved

**Reason:**  
Best balance of portability, simplicity, performance, cross-platform support, AI-assisted development, and maintainability.

**Alternatives Considered:**

- Rust
- TypeScript
- Python

****Replacement Risk:****  Low

---

## EDL-002 — UI-Independent Core Business Logic

**Decision:**  
Core Business Logic is UI Independent

**Status:**  
✅ Approved

**Reason:**  
The same Go core will power:

- CLI
- Desktop App
- Android App
- Future Interfaces

without rewriting business logic.

****Replacement Risk:****  Very Low

---

## EDL-003 — Project Architecture

**Decision:**  
Project Architecture = Modular Monolith

**Status:**  
✅ Approved

**Reason:**  
Provides the best balance of simplicity, modularity, maintainability, AI-assisted development, and future scalability while avoiding unnecessary distributed-system complexity.

**Alternatives Considered:**

- Traditional Monolith
- Microservices

****Replacement Risk:****  Very Low

---

## EDL-004 — Consistent Internal Module Structure

**Decision:**  
Every major module shall follow a consistent internal structure.

**Status:**  
✅ Approved

**Reason:**  
Improves code discoverability, consistency, testing, maintainability, and AI-assisted development.

**Recommended Internal Structure:**

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

**Replacement Risk:** Very Low


---

## EDL-005 — Project Layout

Decision:
Project Layout = Internal-First Architecture

Status:
✅ Approved

Reason:
Provides strong encapsulation, clear module ownership, better maintainability, and excellent compatibility with AI-assisted development.

Selected Structure:

```text
mockctl/
├── cmd/
├── internal/
├── docs/
├── scripts/
├── test/
├── assets/
└── go.mod
```

Alternatives Considered:

Standard Go Layout

Flat Layout


**Replacement Risk:** Very Low


---

## EDL-006 — Go Module Strategy

Decision:
Repository shall use a Single Go Module.

Status:
✅ Approved

Reason:
Simplifies dependency management, builds, testing, refactoring, releases, and AI-assisted development.

Alternatives Considered:

Multi-module Repository


**Replacement Risk:** Very Low


---

## EDL-007 — Runtime

Decision:
Runtime = Go Native Runtime

Status:
✅ Approved

Reason:
Provides excellent performance, portability, concurrency, low resource usage, and cross-platform compatibility while keeping distribution simple through a single executable.

Alternatives Considered:

JavaScript Runtime (Node.js/Bun/Deno)

JVM

.NET Runtime


**Replacement Risk:** Very Low


---

## EDL-008 — Application Runtime Model

Decision:
Application Runtime = Hybrid (CLI + Server)

Status:
✅ Approved

Reason:
The same binary supports both command execution and long-running server mode, enabling future desktop and Android integrations without changing the business logic.

Execution Modes:

CLI Mode

Server Mode


**Replacement Risk:** Very Low


---

## EDL-009 — Engineering Principle: Official First

**Decision:**  
Engineering Principle = Official First

**Status:**  
✅ Approved

**Statement:**  
When the Go Standard Library or Official Go Toolchain adequately satisfies a requirement, it shall be preferred over third-party dependencies.

Third-party libraries shall only be adopted when they provide a clear, measurable engineering advantage that cannot reasonably be achieved using the standard library.

**Benefits:**

- Reduced dependency count
- Improved long-term maintainability
- Better security
- Easier upgrades
- More predictable AI-generated code
- Reduced supply-chain risk

**Replacement Risk:**  
None

---

## EDL-010 — Package Manager

**Decision:**  
Package Manager = Go Modules

**Status:**  
✅ Approved

**Supporting Rule:**  
Vendoring shall remain optional and shall only be used when a release, enterprise deployment, or fully offline build explicitly requires it.

**Reason:**  
Go Modules provide the official, simplest, most maintainable, and industry-standard dependency management solution for Mock:ctl.

**Alternatives Considered:**

- Go Workspaces
- Permanent Vendoring
- Hybrid Workspace

**Replacement Risk:**  Very Low

---

## EDL-011 — Build System

**Decision:**  
Build System = Go Toolchain

**Status:**  
✅ Approved

**Reason:**  
The official Go Toolchain shall be the primary build system for Mock:ctl.

It provides:

- Official support
- Cross-platform compatibility
- Excellent Termux support
- Zero additional build dependencies
- Fast compilation
- Excellent AI compatibility

**Alternatives Considered:**

- Make
- Mage
- Task

**Replacement Risk:**  Very Low

---

## EDL-012 — Go Toolchain as the Single Source of Truth for Builds

**Decision:**  
Go Toolchain is the Single Source of Truth for Builds

**Status:**  
✅ Approved

**Statement:**  
All official builds shall be performed using the Go Toolchain.

Automation scripts such as Make, Task, Mage, Shell, PowerShell, and similar tools may exist only as convenience wrappers.

They must never contain build logic that cannot be executed directly through the Go Toolchain.

**Reason:**

- Prevents build fragmentation
- Eliminates tool lock-in
- Improves portability
- Simplifies onboarding
- Ensures deterministic builds
- Supports AI-assisted development

**Replacement Risk:**  None

---

## EDL-013 — CLI Framework

**Decision:**  
CLI Framework = Cobra

**Status:**  
✅ Approved

**Reason:**  
Cobra provides the best balance of:

- Mature ecosystem
- Long-term maintainability
- Nested command support
- Auto-generated help
- Shell completion
- AI-friendly APIs
- Cross-platform compatibility
- Excellent documentation

**Alternatives Considered:**

- Go Standard Library (`flag`)
- `urfave/cli`
- `kong`

**Replacement Risk:**  Low

---

## EDL-014 — CLI Commands Shall Contain No Business Logic

**Decision:**  
CLI Commands Shall Contain No Business Logic

**Status:**  
✅ Approved

**Statement:**  
CLI commands are responsible only for:

- Input validation
- Argument parsing
- Calling Application Services
- Rendering output

Business logic shall never reside inside Cobra commands.

**Reason:**  
This enables the same Application Core to be reused by:

- CLI
- Desktop Application
- Android Application
- Future REST API
- Future Plugin System

**Replacement Risk:**  None

---

## EDL-015 — Configuration Format

**Decision:**  
Configuration Format = YAML

**Status:**  
✅ Approved

**Statement:**  
YAML shall be the primary configuration format for Mock:ctl.

Only a simple, human-friendly subset of YAML shall be used. Complex YAML features shall be avoided.

**Reason:**

- Human-readable
- Easy to edit
- Supports comments
- Excellent documentation compatibility
- AI-friendly
- Industry standard for configuration

**Alternatives Considered:**

- JSON
- TOML
- HCL

**Replacement Risk:**  Low

---

## EDL-016 — Layered Configuration Precedence

**Decision:**  
Layered Configuration Precedence

**Status:**  
✅ Approved

**Configuration Order (Highest → Lowest):**

1. CLI Flags
2. Environment Variables
3. Project Configuration
4. Global Configuration
5. Built-in Defaults

**Reason:**  
Provides predictable behavior while allowing temporary overrides without modifying configuration files.

**Replacement Risk:**  Very Low


---

## EDL-017 — Human-First Configuration Philosophy

**Decision:**  
Human-First Configuration Philosophy

**Status:**  
✅ Approved

**Principles:**

- Human editable
- Fully validated
- Version controlled
- Backward compatible
- Self-documenting
- Safe defaults
- Unknown keys generate warnings
- Strict mode treats unknown keys as errors

**Reason:**  
Configuration should optimize for clarity, maintainability, and long-term usability rather than compactness or cleverness.

**Replacement Risk:**  None

---

## EDL-018 — Logging Library

**Decision:**  
Logging Library = Go Standard Library (`log/slog`)

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use the Go Standard Library `log/slog` package as its primary logging framework.

**Reason:**

- Official Go package
- Structured logging built-in
- Zero third-party dependency
- Excellent long-term maintainability
- Human-readable and JSON output support
- AI-friendly API
- Cross-platform compatibility

**Alternatives Considered:**

- Zap
- Zerolog
- Logrus

**Replacement Risk:**  Low

---

## EDL-019 — Structured Logging by Default

**Decision:**  
Structured Logging by Default

**Status:**  
✅ Approved

**Rules:**

- Structured key-value logging
- Standard log levels only: `DEBUG`, `INFO`, `WARN`, `ERROR`
- Human-readable logs during development
- JSON logs for production
- Sensitive information must always be masked
- CLI output and diagnostic logs remain separate
- Every error log should include sufficient context

**Reason:**  
Provides consistent debugging, observability, maintainability, and production readiness while keeping user-facing output clean.

**Replacement Risk:**  None

---

## EDL-052 — Goroutine Leak Detection Library

**Decision:**  
Goroutine Leak Detection = `go.uber.org/goleak`

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use Uber's `goleak` library to detect and prevent goroutine leaks during Integration and End-to-End (E2E) testing.

**Reason:**  
- Detects orphaned background workers (e.g., Cloud Sync goroutines).
- Ensures the HTTP server shuts down cleanly upon context cancellation.
- Prevents catastrophic memory leaks in the production simulation engine.
- Industry-standard tool for Go concurrency safety.

**Alternatives Considered:**
- Manual `runtime.NumGoroutine()` checks (brittle and inaccurate).

**Replacement Risk:**  Very Low

---

## EDL-053 — CLI as an Internal Developer Harness

**Decision:**  
CLI Deployment Target = Internal Use Only

**Status:**  
✅ Approved

**Statement:**  
The Mock:ctl Command Line Interface (CLI) is strictly an internal developer and testing tool. It will NOT be distributed to end-users as a standalone public product. The public release distributed to "Real Users" will exclusively be the Flutter Application (which internally wraps the Go Backend).

**Reason:**  
- Ensures the Go engine is fully testable and debuggable locally by engineers without launching the UI.
- Prevents end-user confusion by offering only one polished, premium GUI product.
- Maintains a clean architecture where the CLI and Desktop share the same application core.

**Replacement Risk:**  None

---

**End of Document**

**Decision Range: EDL-001 → EDL-053**

---

## EDL-020 — OpenAPI Parser

**Decision:**  
OpenAPI Parser = `kin-openapi`

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use the `kin-openapi` library as its primary OpenAPI parser.

**Reason:**

- Excellent OpenAPI 3.x support
- Mature ecosystem
- Active maintenance
- Schema validation
- `$ref` resolution
- YAML and JSON support
- Strong documentation
- High community adoption

**Alternatives Considered:**

- Custom Parser
- `libopenapi`
- Swagger Parser Libraries

**Replacement Risk:**  Low

---

## EDL-021 — Parser Abstraction Layer

**Decision:**  
Parser Abstraction Layer

**Status:**  
✅ Approved

**Statement:**  
All third-party parser libraries shall be isolated behind an internal Parser Adapter.

The remainder of the application shall interact only with Mock:ctl's internal specification model.

**Rules:**

- No business logic depends directly on parser library types.
- Internal specification model is the canonical model.
- Parser implementations are replaceable.
- Generator and Runtime remain parser-agnostic.

**Reason:**  
Reduces coupling, simplifies testing, enables future parser replacement, and protects the core architecture from external dependency changes.

**Replacement Risk:**  None

---

## EDL-022 — Fake Data Library

**Decision:**  
Fake Data Library = `gofakeit`

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use the `gofakeit` library as its primary fake data provider.

**Reason:**

- Mature ecosystem
- Excellent documentation
- Deterministic seed support
- Realistic fake data
- Locale support
- High performance
- AI-friendly API

**Alternatives Considered:**

- Custom Faker
- `faker`
- Random-only Generation

**Replacement Risk:**  Low

---

## EDL-023 — Deterministic Fake Data Generation

**Decision:**  
Deterministic Fake Data Generation

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall generate deterministic fake data by default.

**Rules:**

- Same input + same seed = same output
- Seed is configurable
- Random mode is optional
- Reproducible testing is the default behavior

**Reason:**  
Ensures reliable testing, debugging, regression testing, and consistent development environments.

**Replacement Risk:**  Very Low

---

## EDL-024 — Fake Data Provider Abstraction

**Decision:**  
Fake Data Provider Abstraction

**Status:**  
✅ Approved

**Statement:**  
The fake data library shall be isolated behind an internal provider interface.

Business logic shall never directly depend on the external faker library.

**Rules:**

- Provider interface owned by Mock:ctl
- Library-specific code isolated in adapters
- Custom providers supported
- Future library replacement without generator changes

**Reason:**  
Maintains architectural independence, simplifies testing, and protects the application core from external library changes.

**Replacement Risk:**  None


---

## EDL-025 — HTTP Foundation and Router

**Decision:**

- HTTP Foundation = Go Standard Library (`net/http`)
- HTTP Router = Chi

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use the Go Standard Library `net/http` as its HTTP foundation.

Routing shall be implemented using the `Chi` router.

**Reason:**

- Preserves compatibility with the Go standard ecosystem
- Lightweight and mature router
- Excellent middleware support
- Minimal vendor lock-in
- High maintainability
- Strong AI compatibility
- Easy future migration if required

**Alternatives Considered:**

- `net/http` only
- Gin
- Echo
- Fiber

**Replacement Risk:**  Very Low

---

## EDL-026 — HTTP Layer Separation

**Decision:**  
HTTP Layer Separation

**Status:**  
✅ Approved

**Statement:**  
The HTTP Adapter layer shall own all HTTP-specific logic.

Application Services and Domain Models shall never depend on HTTP request/response types or router-specific types.

**Rules:**

- Chi types remain inside the HTTP Adapter.
- Domain remains transport-agnostic.
- Business logic is reusable by CLI, Desktop, Android, and future APIs.

**Reason:**  
Protects the core architecture from transport-layer changes and enables multiple interfaces over the same business logic.

**Replacement Risk:**  None

---

## EDL-027 — Standard Middleware Pipeline

**Decision:**  
Standard Middleware Pipeline

**Status:**  
✅ Approved

**Pipeline:**

```text
Recovery
    ↓
Request ID
    ↓
Structured Logging
    ↓
Metrics
    ↓
Authentication (Future)
    ↓
Request Handler
```

**Rules:**

Middleware order shall remain deterministic.

Recovery middleware must always be first.

Request IDs shall be propagated through the request lifecycle.

Logging shall use structured logs (EDL-019).


**Reason:**
Provides predictable request handling, simplifies debugging, and establishes a consistent processing model.

**Replacement Risk:** Very Low


---

## EDL-028 — Testing Foundation and Assertions

**Decision:**

Testing Foundation = Go Standard Library (testing)

Assertions = testify


**Status:**
✅ Approved

Statement:
Mock:ctl shall use the Go Standard Library testing package as its testing foundation.

The testify library shall be used only for assertions and test utilities.

**Reason:**

Official testing framework

Zero-friction integration

Readable assertions

Excellent documentation

High AI compatibility

Minimal dependency footprint


**Alternatives Considered:**

testing only

gomock

Other testing frameworks


**Replacement Risk:** Low


---

## EDL-029 — Testing Philosophy

**Decision:**
Testing Philosophy

**Status:**
✅ Approved

**Rules:**

Unit tests are the primary testing layer.

Integration tests verify adapters.

End-to-End tests verify complete workflows.

Every bug fix requires a regression test.

Golden tests validate generated outputs.

Tests shall be deterministic by default.

Manual fakes are preferred over mocking frameworks.

Table-driven tests should be preferred where appropriate.


**Reason:**
Ensures high confidence, maintainable test suites, reliable regression detection, and long-term code quality.

**Replacement Risk:** None


---

## EDL-030 — Coverage Policy

**Decision:**
Coverage Policy

**Status:**
✅ Approved

**Rules:**

Coverage is a confidence indicator, not a target.

Critical business logic should maintain very high coverage.

Generated code is excluded from coverage metrics.

Meaningful tests are preferred over artificial coverage.

Benchmarks are maintained for performance-critical paths.


**Reason:**
Encourages effective testing without incentivizing low-value tests created solely to increase percentages.

**Replacement Risk:** None


---

## EDL-031 — Code Formatting Standard

**Decision:**
Code Formatting Standard

**Status:**
✅ Approved

**Tools:**

gofmt

goimports


**Statement:**
Mock:ctl shall use gofmt as the official code formatter and goimports for automatic import management.

**Rules:**

Formatting is automatic.

Import organization is automatic.

Formatting changes are not manually reviewed.

All source code must conform before merge.


**Reason:**
Provides a consistent codebase, eliminates formatting debates, and aligns with the Go ecosystem.

**Replacement Risk:** None


---

## EDL-032 — Static Analysis Standard

**Decision:**
Static Analysis Standard

**Status:**
✅ Approved

**Tools:**

go vet

golangci-lint


**Approved Linters:**

govet

staticcheck

errcheck

ineffassign

unused

revive


**Rules:**

Small curated linter set

CI enforces lint compliance

New linters require engineering approval

Lint suppressions require documented justification


**Reason:**
Maintains high code quality while avoiding excessive developer friction.

**Replacement Risk:** Low


---

## EDL-033 — Code Style Governance

**Decision:**  
Code Style Governance

**Status:**  
✅ Approved

**Statement:**  
`gofmt` is the single source of truth for code style.

**Rules:**

- No personal formatting preferences
- No style debates in reviews
- Consistency over individual taste
- AI-generated code follows identical standards

**Reason:**  
Reduces review noise, improves readability, and ensures a uniform codebase across all contributors.

**Replacement Risk:**  
None

---

## EDL-034 — Repository Hosting & CI Platform

**Decision:**  
Repository Hosting & CI Platform

**Status:**  
✅ Approved

**Repository:**  
GitHub

**CI Platform:**  
GitHub Actions

**Statement:**  
Mock:ctl shall use GitHub as the official source code repository and GitHub Actions as the primary Continuous Integration platform.

**Reason:**

- Native GitHub integration
- Excellent Go ecosystem support
- Matrix builds
- Release automation
- Large community
- Strong AI compatibility
- Low maintenance overhead

**Alternatives Considered:**

- GitLab CI
- Jenkins
- CircleCI
- Azure Pipelines

**Replacement Risk:**  
Low

---

## EDL-035 — Continuous Integration Pipeline

**Decision:**  
Continuous Integration Pipeline

**Status:**  
✅ Approved

**Pipeline:**

1. Repository Checkout
2. Go Environment Setup
3. Dependency Restore
4. Format Verification
5. Import Verification
6. Static Analysis
7. Linting
8. Unit Tests
9. Integration Tests
10. Coverage Report
11. Build Verification

**Rules:**

- Every Pull Request must pass the complete CI pipeline.
- Failed CI blocks merging.
- CI configuration is version controlled.
- Pipeline execution must remain deterministic.

**Reason:**  
Provides consistent quality gates and prevents unstable changes from entering the main branch.

**Replacement Risk:**  
Very Low

---

## EDL-036 — Release Automation

**Decision:**  
Release Automation

**Status:**  
✅ Approved

**Rules:**

- Releases are created from Git tags.
- Cross-platform binaries are generated automatically.
- Checksums are generated automatically.
- Release notes are generated automatically.
- Manual binary creation is prohibited.
- Release artifacts are reproducible.

**Reason:**  
Ensures reliable, repeatable, and auditable releases while eliminating manual release errors.

**Replacement Risk:**  
Very Low

---

## EDL-037 — Release Automation Tool

**Decision:**  
Release Automation Tool = GoReleaser

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use GoReleaser as the official release automation tool.

**Reason:**

- Industry-standard Go release automation
- Cross-platform builds
- Automatic packaging
- Checksum generation
- GitHub Release integration
- Strong community adoption
- Excellent long-term maintainability

**Alternatives Considered:**

- Manual Releases
- Custom Release Scripts

**Replacement Risk:**  
Low

---

## EDL-038 — Versioning Strategy

**Decision:**  
Versioning Strategy

**Status:**  
✅ Approved

**Standard:**  
Semantic Versioning (SemVer)

**Format:**

```text
MAJOR.MINOR.PATCH

Rules:

MAJOR = Breaking changes

MINOR = New features

PATCH = Bug fixes
```
Git tags are the official release source

Pre-releases use SemVer pre-release identifiers


**Reason:**
Provides predictable versioning, industry-standard compatibility, and clear upgrade expectations.

**Replacement Risk:** None


---

## EDL-039 — Distribution Strategy

**Decision:**
Distribution Strategy

**Status:**
✅ Approved

**Phase 1:**

- GitHub Releases
- Cross-platform binaries


**Phase 2:**

- Homebrew
- Scoop
- Winget


**Phase 3:**

- Docker
- APT
- Chocolatey


**Rules:**

- Every distribution channel must be fully automated.
- Manual distribution is prohibited.
- Release artifacts must be reproducible.


**Reason:**
Supports gradual ecosystem expansion while maintaining high release quality and low maintenance overhead.

**Replacement Risk:** Very Low


---

## EDL-040 — Desktop Framework

**Decision:**
Desktop Framework

Status:
✅ Approved

**Framework:**
Flutter

**Backend:**
Embedded Go Core (Dart FFI)

**Statement:**
Mock:ctl shall use Flutter as the official desktop application framework. The desktop application shall embed the shared Go backend using Dart FFI.

**Reason:**

Shared UI framework with Android

Maximum code reuse

Native performance

Cross-platform support

Single frontend technology

Long-term maintainability


**Alternatives Considered:**

- Wails + React
- Electron
- Tauri


**Replacement Risk:** Low


---

## EDL-041 — Desktop Application Architecture

**Decision:**  
Desktop Application Architecture

**Status:**  
✅ Approved

**Statement:**  
The desktop application shall embed the shared Go backend.

Business logic shall remain inside Go while Flutter is responsible only for the presentation layer.

**Rules:**

- Flutter contains no business logic.
- Go owns all application services.
- UI communicates through Dart FFI.
- Desktop remains a thin client.

**Reason:**  
Ensures consistent architecture across Desktop and Android while eliminating duplicated business logic.

**Replacement Risk:** Very Low

---

## EDL-042 — Shared Core Architecture

**Decision:**  
Shared Core Architecture

**Status:**  
✅ Approved

**Statement:**  
All user interfaces shall share the same Application Services and Core business logic.

**Rules:**

- CLI and Desktop reuse the same Application layer.
- Android reuses the same backend.
- Business logic exists only once.
- UI layers remain thin.
- Domain models remain platform independent.

**Reason:**  
Maximizes code reuse, minimizes maintenance effort, ensures consistent behavior across all platforms, and preserves architectural integrity.

**Replacement Risk:** None

---

## EDL-043 — Android Framework

**Decision:**  
Android Framework

**Status:**  
✅ Approved

**Framework:**  
Flutter

**Backend:**  
Embedded Go Core (Dart FFI)

**Statement:**  
Mock:ctl shall use Flutter as the official Android application framework while embedding the shared Go backend.

**Reason:**

- Shared UI framework
- Shared backend
- Native performance
- Excellent developer productivity
- Cross-platform architecture
- Long-term maintainability

**Alternatives Considered:**

- Native Android
- React Native
- Kotlin Multiplatform

**Replacement Risk:** Low

---

## EDL-044 — Plugin System

**Decision:**  
Plugin System

**Status:**  
✅ Approved

**Plugin Format:**  
WebAssembly (WASM)

**Statement:**  
Mock:ctl shall support extensibility through WebAssembly plugins.

**Reason:**

- Secure sandboxing
- Cross-platform compatibility
- Language independence
- Excellent portability
- Long-term ecosystem support

**Alternatives Considered:**

- Go plugins
- Lua
- JavaScript scripting

**Replacement Risk:** Low

---

## EDL-045 — Plugin Runtime

**Decision:**  
Plugin Runtime

**Status:**  
✅ Approved

**Runtime:**  
wazero

**Statement:**  
Mock:ctl shall execute WebAssembly plugins using wazero.

**Reason:**

- Pure Go implementation
- No external runtime
- Excellent portability
- Small dependency footprint
- Strong maintainability

**Alternatives Considered:**

- Wasmtime
- Wasmer

**Replacement Risk:** Low

---

## EDL-046 — Plugin Distribution

**Decision:**  
Plugin Distribution

**Status:**  
✅ Approved

**Statement:**  
Plugins shall initially be distributed through GitHub Releases.

A dedicated plugin registry may be introduced in the future without changing the plugin architecture.

**Reason:**

- Simple distribution
- Minimal infrastructure
- Easy version management
- Future migration path

**Replacement Risk:** Very Low

---

## EDL-047 — Future Cloud Strategy

**Decision:**  
Future Cloud Strategy

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall follow a Local-first, Cloud-ready architecture.

**Rules:**

- Local functionality is the primary experience.
- Cloud features remain optional.
- Existing workflows must function offline.
- Cloud services must not become mandatory.

**Reason:**  
Preserves developer independence while allowing future collaboration features.

**Replacement Risk:** Very Low

---

## EDL-048 — Shared Cross-Platform Strategy

**Decision:**  
Shared Cross-Platform Strategy

**Status:**  
✅ Approved

**Statement:**  
Desktop, Android and future platforms shall reuse the same Go backend while Flutter provides the user interface where applicable.

**Reason:**

- Single codebase for business logic
- Consistent behavior
- Reduced maintenance
- Easier testing
- Faster feature delivery

**Replacement Risk:** None

---

## EDL-049 — Technology Stack Completion

**Decision:**  
Technology Stack Completion

**Status:**  
✅ Approved

**Statement:**  
The approved technology stack defined by EDL-001 through EDL-049 constitutes the official implementation foundation of the Mock:ctl project.

**Rules:**

- Technology changes require a new Engineering Decision.
- Existing approved technologies remain authoritative until superseded.
- PKS-021 shall reference these Engineering Decisions.

**Reason:**  
Provides a stable engineering baseline for implementation, documentation, and future architectural evolution.

**Replacement Risk:** None

---

## EDL-050 — Hybrid Storage Architecture (bbolt + In-Memory)

**Decision:**  
Hybrid Storage Architecture

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use a two-tier database architecture. Ephemeral mock API state will reside strictly in RAM (`map[string]map[string]any`) guarded by a `sync.RWMutex`. Permanent system configuration, user licensing, and telemetry data will be persisted using `bbolt`, a pure-Go embedded Key-Value store.

**Reason:**  
- **Zero-CGO:** Using `bbolt` avoids C-bindings (unlike SQLite), maintaining seamless cross-compilation for Flutter Desktop and Android platforms.
- **Monetization Readiness:** Provides a secure, persistent local anchor (`SystemStore`) for managing premium SaaS features without polluting the fast, ephemeral `StateStore` used for Mock routing.
- **Performance:** In-memory maps execute CRUD operations for mock endpoints in nanoseconds.

**Replacement Risk:** High (Storage abstraction makes it possible, but data migration scripts would be required).

---

## EDL-051 — Frictionless Authentication & JWT Licensing

**Decision:**  
Frictionless Authentication & JWT Licensing

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall not use manual, copy-paste License Keys for unlocking premium SaaS features. Instead, it will use Frictionless Authentication (OAuth Magic Links) to authenticate users and fetch a JSON Web Token (JWT) secured with Asymmetric Cryptography (RS256).

**Reason:**  
- **User Experience:** Magic links and Deep Links (`mockctl://auth`) provide a seamless, modern upgrade path from the Desktop UI.
- **Security:** RS256 JWTs cannot be tampered with locally. If a user modifies the payload (e.g., extending the Expiry Date), the signature becomes invalid and the token is rejected by the Go backend.
- **Piracy Prevention:** Enables a strict "Offline Lease" (e.g., 30 Days) that automatically locks Pro features if the app cannot phone-home to renew the lease.

**Replacement Risk:** High (Tied deeply into the `SystemStore` and Cloud API architecture).

---

## EDL-052 — Goroutine Leak Detection Library

**Decision:**  
Goroutine Leak Detection = `go.uber.org/goleak`

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use Uber's `goleak` library to detect and prevent goroutine leaks during Integration and End-to-End (E2E) testing.

**Reason:**  
- Detects orphaned background workers (e.g., Cloud Sync goroutines).
- Ensures the HTTP server shuts down cleanly upon context cancellation.
- Prevents catastrophic memory leaks in the production simulation engine.
- Industry-standard tool for Go concurrency safety.

**Alternatives Considered:**
- Manual `runtime.NumGoroutine()` checks (brittle and inaccurate).

**Replacement Risk:**  Very Low

---

## EDL-053 — CLI as an Internal Developer Harness

**Decision:**  
CLI Deployment Target = Internal Use Only

**Status:**  
✅ Approved

**Statement:**  
The Mock:ctl Command Line Interface (CLI) is strictly an internal developer and testing tool. It will NOT be distributed to end-users as a standalone public product. The public release distributed to "Real Users" will exclusively be the Flutter Application (which internally wraps the Go Backend).

**Reason:**  
- Ensures the Go engine is fully testable and debuggable locally by engineers without launching the UI.
- Prevents end-user confusion by offering only one polished, premium GUI product.
- Maintains a clean architecture where the CLI and Desktop share the same application core.

**Replacement Risk:**  None

---

## EDL-054 — API Rate Limiting Foundation

**Decision:**  
Rate Limiter = `golang.org/x/time/rate`

**Status:**  
✅ Approved

**Statement:**  
Mock:ctl shall use the official `golang.org/x/time/rate` package (Token Bucket algorithm) for all HTTP API rate-limiting requirements (specifically the Admin API).

**Reason:**  
- Ensures robust, race-condition-free rate limiting (e.g., 100 req/s).
- Maintained by the core Go team, ensuring long-term stability and security.
- Prevents malicious or bug-induced local DoS attacks on the Mock server.

**Alternatives Considered:**
- Implementing a custom rate limiter (complex, high risk of concurrency bugs).
- Third-party packages (violates Official First principle when `x/time` exists).

**Replacement Risk:**  Low

---

## Document Governance

This Engineering Decision Log is the authoritative record of approved engineering decisions for Mock:ctl.

Any change to an approved engineering decision must be documented as a new Engineering Decision or as an explicit superseding decision.

Existing decisions remain authoritative until they are formally superseded.

The Engineering Decision Log should be reviewed whenever a significant change is proposed to:

- Technology Stack
- System Architecture
- Repository Architecture
- Component Architecture
- Software Design
- Database Architecture
- API Architecture
- Coding Standards
- Testing Strategy
- Deployment Architecture

The Engineering Decision Log is maintained as part of the Project Knowledge System and remains synchronized with the engineering documentation that depends upon it.

---

## Document Status

**Status:** Active

**Decision Range:** EDL-001 → EDL-054

**Authority:** Engineering Decision Log

**Primary Related Document:** PKS-021 — Technology Stack

**Next Engineering Documentation:** PKS-022 — Repository & Module Architecture