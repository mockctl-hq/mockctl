# ⚙️ PKS-021 — Technology Stack

> **Project:** Mock:ctl
>
> **Document ID:** PKS-021
>
> **Version:** 1.0
>
> **Status:** Approved
>
> **Owner:** Upen Tudu
>
> **Authors:** Upen Tudu & ChatGPT
>
> **Created:** 2026-08-06
>
> **Last Updated:** 2026-08-06
>
> **Category:** Engineering
>
> **Priority:** Critical

---

# 📖 Executive Summary

This document defines the official technology stack of the Mock:ctl project.

Its purpose is to establish a single source of truth for every technology approved for use within the project. Rather than documenting the complete engineering decision process, this specification identifies the official technologies, their intended responsibilities, and the Engineering Decision Logs (EDLs) that formally approved them.

Detailed technical evaluations, alternative analyses, replacement risks, and engineering discussions are maintained separately within the Engineering Decision Log (EDL) documentation.

This separation ensures that:

- The Technology Stack remains concise and easy to reference.
- Engineering decisions remain fully traceable.
- Technology changes are centrally managed through the EDL process.
- Documentation duplication is minimized.

This document shall be referenced whenever implementing, reviewing, or extending the Mock ecosystem.

---

# 🎯 Purpose

Technology selection within Mock:ctl follows a long-term engineering philosophy rather than short-term development trends.

The primary objective of this document is to establish the specific tools, languages, and frameworks that all contributors must use. Every adopted technology should support the project's core engineering principles:

- Long-term maintainability
- Simplicity
- Reliability
- Cross-platform compatibility
- Performance
- Excellent documentation
- Mature ecosystem
- Strong community support
- AI-assisted development
- Minimal operational complexity

---

# 📌 Scope

This document specifies the exact technological choices for:
- Core language and runtime
- Desktop and Android framework (FFI)
- HTTP Server and routing
- OpenAPI parsing and data generation
- Build and CI/CD tools
- CLI Framework
- WASM Plugin Runtime

This document does not explain the code-level usage of these technologies, which are covered in the respective Architectural and Software Design documents.

---

The project intentionally prefers stable and well-established technologies over experimental alternatives.

Whenever multiple technologies satisfy the same requirement, preference is given to the solution that:

- Integrates naturally with the existing architecture.
- Minimizes dependencies.
- Reduces maintenance cost.
- Provides predictable long-term support.
- Preserves architectural consistency.

All technology selections are governed through the Engineering Decision Log (EDL) process.

---

# 🧩 Core Technology Stack

| Category | Official Technology |
|----------|---------------------|
| Programming Language | Go |
| Architecture | Modular Monolith |
| Runtime | Go Runtime |
| Package Manager | Go Modules |
| Version Control | Git |
| Repository Hosting | GitHub |

---

# 💻 Programming Language

## Official Technology

**Go**

### Primary Purpose

- Core business logic
- Application services
- CLI implementation
- Runtime engine
- HTTP services
- Plugin management
- Future cloud services

### Why Go

Go was selected because it provides:

- Excellent performance
- Simple language design
- Cross-platform compilation
- Outstanding tooling
- Strong standard library
- Long-term maintainability

### Reference

**Engineering Decision:** EDL-001

---

# 🏗 Project Architecture

## Official Technology

**Modular Monolith**

### Primary Purpose

The Modular Monolith architecture serves as the foundational software architecture for the Mock ecosystem.

It provides:

- Clear module boundaries
- High maintainability
- Low operational complexity
- Shared business logic
- Simplified deployment

### Why Modular Monolith

The architecture was selected because it provides:

- Excellent scalability
- Easier development
- Lower infrastructure complexity
- Better code organization
- Smooth evolution toward future services if required

### Reference

**Engineering Decision:** EDL-002

---

# ⚙ Runtime

## Official Technology

**Go Runtime**

### Primary Purpose

The Go Runtime executes:

- CLI commands
- HTTP services
- Business logic
- Plugin execution
- Application services
- Background operations

### Why Go Runtime

The official Go runtime provides:

- Native execution speed
- Efficient concurrency
- Small deployment footprint
- Excellent cross-platform support
- Stable runtime environment

### Reference

**Engineering Decision:** EDL-003

---

# 📦 Package Manager

## Official Technology

**Go Modules**

### Primary Purpose

Go Modules manages:

- Project dependencies
- Module versions
- Dependency integrity
- Reproducible builds

### Why Go Modules

Go Modules was selected because it offers:

- Native Go integration
- Reliable dependency management
- Secure module verification
- Long-term ecosystem support

### Reference

**Engineering Decision:** EDL-005

---

# 🌿 Version Control

## Official Technology

**Git**

## Repository Hosting

**GitHub**

### Primary Purpose

The project's source code shall be managed using Git and hosted on GitHub.

Version control is responsible for:

- Source code management
- Branch management
- Code reviews
- Pull Requests
- Release tagging
- Collaboration

### Why Git & GitHub

The selected platform provides:

- Industry-standard workflows
- Excellent collaboration tools
- Tight GitHub Actions integration
- Strong open-source ecosystem
- Reliable long-term hosting

### Reference

**Engineering Decision:** EDL-034

---

# 🖥 CLI Framework

## Official Technology

**Cobra**

### Primary Purpose

Cobra serves as the official Command-Line Interface (CLI) framework for Mock:ctl.

It is responsible for:

- Command registration
- Command hierarchy
- Flag parsing
- Argument validation
- Help generation
- Shell completion

### Why Cobra

Cobra was selected because it provides:

- Mature ecosystem
- Excellent Go integration
- Strong community support
- High maintainability
- Excellent documentation

### Reference

**Engineering Decision:** EDL-004

---

# 📄 OpenAPI Parser

## Official Technology

**kin-openapi**

### Primary Purpose

The OpenAPI parser is responsible for:

- Reading OpenAPI specifications
- Schema validation
- Request analysis
- Response analysis
- Component resolution
- Reference resolution

### Why kin-openapi

kin-openapi was selected because it provides:

- Complete OpenAPI support
- Mature implementation
- Reliable validation
- Excellent Go compatibility
- Long-term stability

### Reference

**Engineering Decision:** EDL-009

---

# 🌐 HTTP Server

## Official Technology

**Go net/http + Chi**

### Primary Purpose

The HTTP runtime provides:

- HTTP serving
- Request routing
- Middleware execution
- API endpoints
- Mock response delivery

### Why net/http + Chi

The selected stack provides:

- High performance
- Minimal overhead
- Native Go integration
- Flexible routing
- Excellent maintainability

### Reference

**Engineering Decision:** EDL-011

---

# 🎲 Fake Data Library

## Official Technology

**gofakeit**

### Primary Purpose

The fake data library is responsible for generating:

- Names
- Email addresses
- Phone numbers
- Addresses
- UUIDs
- Dates
- Numbers
- Custom fake values

### Why gofakeit

gofakeit was selected because it provides:

- Rich functionality
- Excellent documentation
- Active maintenance
- Native Go integration
- High flexibility

### Reference

**Engineering Decision:** EDL-010

---

# ⚙ Configuration Format

## Official Technology

**YAML**

### Primary Purpose

YAML is used for:

- Project configuration
- Runtime configuration
- Build configuration
- Plugin configuration
- Environment settings

### Why YAML

YAML was selected because it provides:

- Human readability
- Industry adoption
- Flexible structure
- Easy maintenance
- Excellent tooling

### Reference

**Engineering Decision:** EDL-007

---

# 📋 Logging

## Official Technology

**log/slog**

### Primary Purpose

The logging framework provides:

- Structured logging
- Error logging
- Warning logging
- Debug logging
- Runtime diagnostics

### Why log/slog

log/slog was selected because it provides:

- Official Go support
- Structured logging
- Zero external dependency
- Long-term stability

### Reference

**Engineering Decision:** EDL-008

---

# 🗄 Database

## Official Technology

**Hybrid Storage (bbolt + In-Memory Map)**

### Primary Purpose

Mock:ctl utilizes a two-tier database architecture to separate high-speed API simulation from permanent configuration storage:

- **StateStore (In-Memory Map):** Used exclusively for managing the ephemeral JSON state of the mock API endpoints. It operates at nanosecond speed and wipes on exit.
- **SystemStore (bbolt):** A pure Go Key-Value embedded database used for permanent storage of License Keys, Telemetry, and User Settings (Monetization Engine).

### Why Hybrid Storage (bbolt)

This architecture was selected because it provides:

- **Zero-CGO cross-compilation:** `bbolt` is pure Go, keeping Flutter Desktop/Android integrations simple.
- **Insane Speed:** In-memory maps serve dynamic REST endpoints instantly.
- **SaaS-Readiness:** The SystemStore provides a permanent anchor for licensing, analytics, and update histories.
- **Data Isolation:** Separates volatile mock data from critical system data.

### Reference

**Engineering Decision:** EDL-050

---

# 🧪 Testing Stack

## Official Technology

**Go testing + testify**

### Primary Purpose

The testing stack supports:

- Unit testing
- Integration testing
- CLI testing
- Runtime testing
- Regression testing

### Why testing + testify

The selected testing stack provides:

- Native Go compatibility
- Reliable assertions
- Fast execution
- Excellent maintainability

### Reference

**Engineering Decision:** EDL-013

---

# ✨ Linting & Formatting

## Official Technologies

### Formatting

- gofmt
- goimports

### Static Analysis

- go vet
- golangci-lint

### Primary Purpose

The formatting and linting toolchain ensures:

- Consistent code style
- Import organization
- Static analysis
- Code quality
- Early defect detection

### Why These Tools

The selected toolchain provides:

- Official Go tooling
- Excellent automation
- Reliable diagnostics
- Strong ecosystem support

### Reference

**Engineering Decision:** EDL-014

---

# 🔨 Build System

## Official Technology

**Go Toolchain**

### Primary Purpose

The official build system is responsible for:

- Project compilation
- Dependency resolution
- Test execution
- Binary generation
- Cross-platform builds

### Why Go Toolchain

The Go toolchain was selected because it provides:

- Native compilation
- Fast builds
- Cross-compilation support
- Excellent reliability

### Reference

**Engineering Decision:** EDL-006

---

# 🔄 CI/CD

## Official Technology

**GitHub Actions**

### Primary Purpose

The CI/CD platform automates:

- Pull Request validation
- Build verification
- Testing
- Static analysis
- Cross-platform builds
- Release workflows

### Why GitHub Actions

GitHub Actions was selected because it provides:

- Native GitHub integration
- Excellent Go ecosystem support
- Matrix builds
- Release automation
- Low maintenance

### Reference

**Engineering Decision:** EDL-034, EDL-035

---

# 🚀 Release System

## Official Technology

**GoReleaser**

### Primary Purpose

The release system automates:

- Binary generation
- Cross-platform packaging
- Checksum generation
- Release notes
- GitHub Releases

### Why GoReleaser

GoReleaser was selected because it provides:

- Mature automation
- Excellent Go integration
- Reproducible releases
- Minimal manual effort

### Reference

**Engineering Decision:** EDL-036

---

# 🖥 Desktop Framework

## Official Technology

| Component | Official Technology |
|-----------|---------------------|
| Desktop Framework | Embedded Go Core (Dart FFI) |
| Desktop Frontend | Flutter |

### Primary Purpose

The desktop application provides a native graphical interface for the Mock ecosystem while reusing the same backend and business logic used by the CLI.

### Responsibilities

- Project management
- API import and export
- Mock server management
- Configuration management
- Plugin management
- Future visual tooling

### Why Flutter

Flutter was selected because it provides:

• Single UI framework for Desktop and Android
• Maximum code sharing
• Native performance
• Excellent developer productivity
• Long-term maintainability
• Consistent user experience across platforms

### Reference

**Engineering Decision:** EDL-040, EDL-041, EDL-042

---

# 📱 Android Strategy

## Official Technology

| Component | Official Technology |
|-----------|---------------------|
| Mobile Framework | Flutter |
| Backend Integration | Embedded Go Core (Dart FFI) |

### Primary Purpose

The Android application extends the Mock ecosystem to mobile devices while sharing the same backend architecture and business logic.

### Responsibilities

- Project management
- Mock server control
- API inspection
- Runtime management
- Local development workflow

### Why Flutter

Flutter was selected because it provides:

- Cross-platform UI framework
- Excellent developer productivity
- Native performance
- Rich widget ecosystem
- Long-term support

### Architectural Principle

The Android application shall reuse the shared Go backend through an embedded runtime, ensuring that business logic exists only once across all supported platforms.

### Reference

**Engineering Decision:** EDL-043, EDL-044, EDL-045

---

# 🧩 Plugin System (Future)

## Official Technology

| Component | Official Technology |
|-----------|---------------------|
| Plugin Format | WebAssembly (WASM) |
| Runtime | wazero |

### Primary Purpose

The plugin system allows Mock:ctl to be extended without modifying the core application.

### Responsibilities

- Plugin loading
- Plugin execution
- Runtime isolation
- Extension APIs
- Future ecosystem support

### Why WebAssembly + wazero

The selected plugin platform provides:

- Secure execution
- Cross-platform compatibility
- Runtime isolation
- Small runtime footprint
- Native Go integration

### Reference

**Engineering Decision:** EDL-046, EDL-047, EDL-048

---

# ☁ Cloud Support (Future)

## Official Strategy

**Local-first, Cloud-ready**

### Primary Purpose

The project is designed to function entirely offline while remaining capable of supporting optional cloud services in the future.

### Future Possibilities

- Project synchronization
- Team collaboration
- Remote mock servers
- Shared workspaces
- Plugin registry
- Cloud backup

### Design Principles

- Offline-first experience
- Cloud features remain optional
- No mandatory user accounts
- Existing local workflows remain unchanged

### Reference

**Engineering Decision:** EDL-049

---

# 📊 Technology Summary Matrix

| Category | Official Technology | EDL |
|----------|---------------------|-----|
| Programming Language | Go | EDL-001 |
| Project Architecture | Modular Monolith | EDL-002 |
| Runtime | Go Runtime | EDL-003 |
| CLI Framework | Cobra | EDL-004 |
| Package Manager | Go Modules | EDL-005 |
| Build System | Go Toolchain | EDL-006 |
| Configuration Format | YAML | EDL-007 |
| Logging | log/slog | EDL-008 |
| OpenAPI Parser | kin-openapi | EDL-009 |
| Fake Data Library | gofakeit | EDL-010 |
| HTTP Server | net/http + Chi | EDL-011 |
| Database | Hybrid Storage (bbolt) | EDL-050 |
| Testing Stack | testing + testify | EDL-013 |
| Linting & Formatting | gofmt, goimports, go vet, golangci-lint | EDL-014 |
| Repository Hosting | GitHub | EDL-034 |
| CI/CD | GitHub Actions | EDL-035 |
| Release System | GoReleaser | EDL-036 |
| Desktop Framework | Embedded Go Core (Dart FFI) | EDL-040 |
| Desktop Frontend | Flutter | EDL-041 |
| Shared Core Architecture | Shared Application Layer | EDL-042 |
| Android Framework | Flutter | EDL-043 |
| Android Architecture | Embedded Go Core + Dart FFI | EDL-044 |
| Shared Backend Strategy | Single Shared Backend | EDL-045 |
| Plugin Format | WebAssembly (WASM) | EDL-046 |
| Plugin Runtime | wazero | EDL-047 |
| Plugin Distribution | GitHub Releases (Future Registry) | EDL-048 |
| Future Cloud Support | Local-first, Cloud-ready | EDL-049 |

---

# 📚 Related Documents

## Foundation

- PKS-000 — Repository Blueprint
- PKS-001 — Documentation Philosophy
- PKS-002 — Documentation Style Guide
- PKS-003 — Project Knowledge System
- PKS-004 — Documentation Index

---

## Product

- PKS-010 — Vision
- PKS-011 — User Personas
- PKS-012 — User Journey
- PKS-013 — Functional Requirements
- PKS-014 — Non-Functional Requirements
- PKS-015 — MVP Definition
- PKS-016 — Product Requirements Document (PRD)
- PKS-017 — Product Roadmap

---

## Engineering

- PKS-020 — System Architecture
- PKS-021 — Technology Stack
- PKS-022 — Repository & Module Architecture
- PKS-023 — Data Flow Architecture
- PKS-024 — Component Architecture
- PKS-025 — Software Design Document (Master SDD)
- PKS-026 — Database Design
- PKS-027 — API Design
- PKS-028 — Coding Standards
- PKS-029 — Testing Strategy
- PKS-030 — Deployment Architecture

---

# 📝 Revision History

| Version | Date | Description |
|----------|------|-------------|
| 1.0.0 | Initial Release | First official Technology Stack specification based on approved Engineering Decision Logs (EDL-001 → EDL-049). |

---

# ✅ Approval Checklist

- Programming Language defined
- Project Architecture defined
- Runtime defined
- CLI Framework defined
- Package Manager defined
- Build System defined
- Configuration Format defined
- Logging defined
- OpenAPI Parser defined
- Fake Data Library defined
- HTTP Server defined
- Database strategy defined
- Testing Stack defined
- Linting & Formatting defined
- CI/CD defined
- Release System defined
- Desktop Framework defined
- Android Strategy defined
- Plugin System defined
- Future Cloud Support defined
- Technology references verified
- Related documents updated

---

# 📌 Conclusion

The technologies defined in this document constitute the official implementation stack for the Mock:ctl ecosystem.

This specification serves as the authoritative reference for all implementation work and should be consulted before introducing, replacing, or upgrading any technology. Detailed engineering evaluations, alternatives, and approval rationale are maintained separately in the Engineering Decision Log (EDL), ensuring a clear separation between technology specification and engineering governance.

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** Completed

**Architecture Status:** ✅ Foundation Established

**Next Document:** **PKS-022 — Repository & Module Architecture**