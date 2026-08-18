# 🌊 PKS-023 — Data Flow Architecture

> **Project:** Mock:ctl
>
> **Document ID:** PKS-023
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

The Data Flow Architecture defines how information moves through the Mock:ctl system.

It translates the static module boundaries defined in PKS-022 into dynamic runtime sequences.

This document establishes the exact paths that requests, configurations, specifications, mock data, and internal state must follow.

By defining these flows, this architecture ensures that no subsystem bypasses established boundaries, mutates data unexpectedly, or violates the core system architecture.

---

# 🎯 Purpose

The objectives of this document are to:

- Map the core operational workflows of the application.
- Define how external input is transformed into internal models.
- Formalize the lifecycle of stateful API interactions.
- Establish data immutability and mutation boundaries.
- Ensure all module communication adheres to PKS-020 rules.
- Guide developers in implementing state transitions correctly.

---

# 📌 Scope

This document covers:

- Core Data Models.
- The Specification Import data flow.
- The Mock Generation data flow.
- Stateless and Stateful Runtime Execution flows.
- Dynamic Fake Data generation flows.
- The Configuration Resolution flow.
- Rules for data immutability and error propagation.

This document does not define the exact Go interfaces or struct definitions (which belong in PKS-025).

---

# 📦 Core Data Models

Data flowing across boundaries must be encapsulated in clearly defined, domain-owned models.

The primary data models driving Mock:ctl are:

- **Configuration Model:** Represents the merged, resolved configuration from all sources. Owned by `internal/config/`.
- **Specification Model:** Represents the normalized API schema extracted from OpenAPI. Owned by `internal/spec/`.
- **Runtime Definition Model:** Represents the blueprint for executable routes, headers, and response templates. Owned by `internal/generator/`.
- **Memory State Tree (Context):** Represents the live, mutable database of created/updated entities during an active simulation. Owned by `internal/runtime/`.

---

# 🔄 Core Data Flows

Mock:ctl operates through several distinct, sequential data flows.

Each flow crosses specific architectural boundaries defined in PKS-022.

---

## 1️⃣ System & Project Initialization Flow

This flow boots up the Mock:ctl binary, initializing permanent system data and project workspaces.

```text
User Command
      ↓
Presentation (cmd/)
      ↓
Application Core (internal/app/)
      ↓
Storage Layer (SystemStore - bbolt) ➔ Loads License, Telemetry, Global Configs
      ↓
Project Manager (internal/project/)
      ↓
Storage Layer (FileSystem) ➔ Reads Workspace `.mockctl/`
```

The Application Core receives the validated CLI instruction.

First, it accesses the `SystemStore` (the physical database) to read premium license keys, apply global user configurations, and trigger the background Cloud Sync (Phone-Home).

Then, it instructs the Project Manager to initialize or read the local project state, requesting file I/O operations strictly through the Storage Layer abstraction.

---

## 2️⃣ Specification Import Flow

This flow converts an external OpenAPI document into a predictable internal format.

```text
File Path / URL
      ↓
Application Core (internal/app/)
      ↓
Specification Engine (internal/spec/)
      ↓
External Parser (kin-openapi)
      ↓
Internal Specification Model
```

The Application Core supplies the raw API specification path.

The Specification Engine delegates parsing to `kin-openapi`.

The engine then extracts and normalizes this data into a strictly read-only Mock:ctl-specific Specification Model.

No other module is permitted to interact directly with the `kin-openapi` types.

---

## 3️⃣ Mock Generation Flow

This flow translates the internal specification into executable runtime definitions.

```text
Internal Specification Model
      ↓
Application Core (internal/app/)
      ↓
Mock Generation Engine (internal/generator/)
      ↓
Data Generation Engine (internal/data/)
      ↓
Runtime Definition Model
```

The Application Core passes the parsed specification to the Mock Generation Engine.

The generator iterates through endpoints and schemas.

It consults the Data Generation Engine (`gofakeit`) to produce realistic default template payloads.

The final output is a structured, read-only Runtime Definition Model containing routes, methods, and templates.

---

## 4️⃣ Stateless Runtime Data Flow (GET)

This flow handles simple data retrieval during frontend development.

```text
Frontend Client Request (GET)
      ↓
HTTP Server (Chi / net/http)
      ↓
Runtime Engine (internal/runtime/)
      ↓
Route Matching & Memory State Lookup
      ↓
Response Formatting
      ↓
HTTP Server Response
```

The HTTP Server intercepts the incoming request.

The Runtime Engine matches the request against its active Runtime Definitions.

It checks the Memory State Tree to see if the requested entity exists.

The formatted response is returned to the client without modifying any internal state.

---

## 5️⃣ Stateful Mutation Data Flow (POST / PUT / DELETE)

This flow allows Mock:ctl to simulate a real, stateful backend database.

```text
Frontend Client Request (POST/PUT/DELETE)
      ↓
HTTP Server (Chi / net/http)
      ↓
Runtime Engine (internal/runtime/)
      ↓
Memory State Tree Mutation (Insert/Update/Delete)
      ↓
Response Formatting (e.g., 201 Created)
      ↓
HTTP Server Response
```

When a request arrives that implies data mutation, the Runtime Engine processes the payload.

Instead of just returning a static template, it updates its internal Memory State Tree (e.g., saving a new user record).

Future `GET` requests for this data will retrieve the updated state.

The mutation is isolated strictly within the Runtime Engine's memory; it never alters the original Specification Model.

---

## 6️⃣ Dynamic Fake Data Flow

This flow occurs when the runtime needs fresh, randomized data on-the-fly (e.g., a route that returns a random quote on every request).

```text
Runtime Engine (internal/runtime/)
      ↓
Application Boundary Abstraction
      ↓
Data Generation Engine (internal/data/)
      ↓
Fresh Payload
      ↓
Runtime Engine
```

The Runtime Engine must not depend directly on `gofakeit`.

It requests data through an internal abstraction layer (typically defined in `internal/shared/` or injected by `internal/app/`).

The Data Generation Engine returns a fresh, deterministic payload based on the requested schema.

---

## 7️⃣ User-Defined Overrides Flow

This flow allows users to inject custom static payloads, overriding the automatic mock generation.

```text
User Custom File (e.g., overrides.yaml)
      ↓
Project Manager (internal/project/)
      ↓
Application Core (internal/app/)
      ↓
Mock Generation Engine (internal/generator/)
      ↓
Runtime Definition Model (Customized)
```

The Project Manager reads any custom override files supplied by the user.

These overrides are passed through the Application Core into the Mock Generation Engine.

The engine merges the custom overrides with the generated OpenAPI definitions, giving priority to the user's custom payloads before producing the final Runtime Definition Model.

---

## 8️⃣ Chaos & Error Simulation Flow

This flow enables intentional error simulation during frontend testing.

```text
Frontend Client Request
      ↓
HTTP Server
      ↓
Runtime Engine (internal/runtime/)
      ↓
Chaos Rule Evaluation
      ↓
Simulated Error Formatting (e.g., 500, Delay)
      ↓
HTTP Server Response
```

The Runtime Engine checks active chaos configuration rules (e.g., "fail 10% of requests" or "add 500ms delay").

If a rule triggers, the normal data generation or retrieval flow is interrupted.

The Runtime Engine formats the specified error and returns it immediately to the client.

---

## 9️⃣ State Persistence Flow (Snapshot & Restore)

This flow saves the in-memory state tree to disk so that it can survive server restarts.

```text
Runtime Engine (internal/runtime/)
      ↓
Application Core (internal/app/)
      ↓
Storage Abstraction
      ↓
State File (state.json)
```

When a snapshot is triggered (automatically or via CLI), the Runtime Engine exports its active Memory State Tree.

The Application Core routes this data to the Storage Layer to be serialized.

On the next startup, this flow operates in reverse: the Application Core reads the state file and injects it back into the Runtime Engine, restoring the mock backend exactly as it was.

---

## 🔟 Configuration Resolution Flow

This flow resolves the operational parameters of the application.

```text
Built-in Defaults
        ↓
Global Configuration (~/.mockctl)
        ↓
Project Configuration (.mockctl.yaml)
        ↓
Environment Variables
        ↓
Command-Line Flags
```

Configuration is resolved sequentially by the Configuration Manager (`internal/config/`).

The finalized configuration object is injected into the Application Core.

Individual subsystems do not independently parse configuration files or environment variables.

---

# 🔌 Future Extension Points (Plugin Boundary)

Mock:ctl is designed to support WASM plugins in the future (EDL-046, EDL-047). 

While plugins are not part of the MVP, the data flow architecture establishes a clear boundary for future interception.

```text
Runtime Engine
      ↓
[ Future Plugin Interception Point ]
      ↓
WASM Plugin (wazero)
      ↓
Modified Payload
      ↓
Runtime Engine
```

When implemented, plugins will act as middleware around the Runtime Engine. 

They will receive the parsed request data, apply custom logic, and return a modified payload back to the Runtime Engine. 

Because the Runtime Engine is strictly isolated from the Application Core and HTTP layer, this extension point can be added later without requiring a massive architectural rewrite.

---

# 🧱 Data Immutability Rules

To prevent side-effects, data flowing between major subsystems must follow strict immutability rules.

The Specification Model produced by `internal/spec/` is strictly read-only.

The Mock Generation Engine must not alter the specification while generating mock definitions.

The Runtime Definition Model produced by `internal/generator/` is read-only after creation.

The Memory State Tree maintained by `internal/runtime/` is strictly isolated and does not persist across server restarts unless explicitly configured.

---

# 🛡️ Error Propagation & Translation

All incoming data must be validated at the earliest possible architectural boundary.

Errors must be translated appropriately based on the boundary they cross.

**CLI Errors:**
Errors encountered during project setup or CLI execution bubble up to the Presentation layer (`cmd/`), where they are formatted into human-readable terminal output.

**HTTP Errors:**
Errors encountered during an active simulation (e.g., validation failures, missing resources, internal faults) are intercepted by the Runtime Engine.
The Runtime Engine translates these into appropriate HTTP responses (e.g., `400 Bad Request`, `404 Not Found`, `500 Internal Server Error`).

Internal stack traces should never leak into HTTP responses unless the application is running in explicit debug mode.

---

# 🔍 Traceability to Engineering Decisions

The data flow architecture aligns with the following approved Engineering Decisions.

| Decision | Architectural Impact |
|---|---|
| EDL-005 | Validates the separation of concerns between core logic and external requests. |
| EDL-009 | Ensures OpenAPI parsing (`kin-openapi`) remains isolated in the Specification Flow. |
| EDL-010 | Ensures fake data generation (`gofakeit`) remains isolated in the Data Flow. |
| EDL-011 | Establishes `net/http` + `Chi` as the boundary for the Runtime Execution Flow. |

---

# 🔗 Related Documents

**Foundation**

- PKS-000 — Repository Blueprint
- PKS-002 — Documentation Style Guide

**Product**

- PKS-016 — Product Requirements Document

**Engineering**

- PKS-020 — System Architecture
- PKS-021 — Technology Stack
- PKS-022 — Repository & Module Architecture

**Next Document**

- PKS-024 — Component Architecture

---

# 📜 Revision History

| Version | Date | Description |
|----------|------------|----------------------------------------------|
| 1.0 | 2026-08-15 | Initial approved release |

---

# ✅ Approval Checklist

- Executive summary completed
- Core data models defined
- Data flows mapped to PKS-022 boundaries
- Project initialization flow defined
- Specification import flow defined
- Mock generation flow defined
- Stateless and Stateful execution flows separated
- Dynamic data generation flow defined
- User-defined overrides flow defined
- Chaos and error simulation flow defined
- State persistence flow defined
- Configuration resolution flow mapped
- Future plugin extension boundary established
- Immutability and mutation rules established
- Error propagation and HTTP translation defined
- Traceability to EDLs verified
- Formatting follows PKS style guide

---

# 📌 Conclusion

The Data Flow Architecture defines the precise movement of information across Mock:ctl's subsystems.

By explicitly modeling stateful mutations and dynamic data generation alongside static initialization flows, this architecture ensures that the system remains predictable and decoupled.

These flows form the behavioral contract for the application and will directly inform the structural design of the Go components.

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** Completed

**Architecture Status:** ✅ Established

**Next Document:** **PKS-024 — Component Architecture**
