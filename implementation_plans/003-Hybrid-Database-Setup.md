# Implementation Plan: 003-Hybrid-Database-Setup

## 🎯 Objective
Initialize the **Hybrid Storage Architecture** as strictly defined in `PKS-026`. This entails building the dual-engine database system: 
1. The **StateStore**: An ultra-fast, ephemeral, thread-safe in-memory map for simulating REST API payloads.
2. The **SystemStore**: A persistent, pure-Go `bbolt` key-value embedded database for storing monetization tokens, telemetry, and CLI settings.

## 📌 Prerequisites
- Project initialized and Cobra CLI framework configured (002 complete).
- Full comprehension of `PKS-026` (Memory State Architecture).

## 🛠️ Execution Steps

### Step 1: Install Database Dependencies
Download the pure-Go key-value engine required for the `SystemStore`.
- Command: `go get go.etcd.io/bbolt`
- Command: `go mod tidy`

### Step 2: Define Core Domain Models
Following `PKS-024` and `PKS-025`, define the core domain entities that represent Mock:ctl's routing logic.
- File: `internal/core/domain/endpoint.go`
- *Entities:* Define `EndpointHandler` struct (Method, Path, PathParams, QueryParams) and `ResponseTemplate` struct (Headers, Body).

### Step 3: Define Core Domain Interfaces (Hexagonal Ports)
Establish the core database contracts that the rest of the application will rely on via strict Dependency Injection (No Global State allowed for stores).
- File: `internal/core/ports/database.go`
- *Sentinel Errors:* Define domain-specific errors (e.g., `ErrNotFound`, `ErrLimitReached`) so that adapters can map database-specific errors (like bbolt errors) to pure domain errors.
- *StateStore Interface:* Define methods for `Insert`, `Get`, `List`, `Update`, `Delete`, and `Reset` (passing `context.Context`).
- *SystemStore Interface:* Define methods for `GetSetting`, `SetSetting`, `SaveAuthToken`, `GetAuthToken`, `LogTelemetry`, and `Close` (all passing `context.Context` as the first parameter as per PKS-028 boundary rules).

### Step 4: Implement the In-Memory StateStore
Construct the core engine responsible for handling mock API data.
- File: `internal/adapter/db/memory_state.go`
- *Implementation:* Create a `map[string]map[string]any` representing `[CollectionName][DocumentID]Payload`.
- *Concurrency:* Wrap the entire structure in a global `sync.RWMutex`. Use `RLock()` for reading and `Lock()` for mutating.
- *Flat Path Routing:* Nested resources must be flattened into composite collections (e.g., `/users/123/posts/456` becomes Collection `users/123/posts`).
- *PATCH Deep-Merge:* Implement a recursive deep-merge algorithm for PATCH requests.
- *ID Generation Hierarchy:* 1) Client-provided ID, 2) Schema-aware Auto-Increment, 3) UUIDv4 fallback (via `gofakeit`).
- *Auto-Timestamps:* Automatically inject `createdAt` and `updatedAt`.
- *OOM Protection:* Strictly limit each collection to a maximum of 10,000 documents (return 429/507).
- *Graceful Auto-Save:* Ensure the map state is dumped to `temp-state.json` on `SIGINT`/`SIGTERM`.

### Step 5: Implement the bbolt SystemStore
Construct the permanent embedded database for system configurations.
- File: `internal/adapter/db/bbolt_system.go`
- *Implementation:* Initialize the `bbolt` database file at `~/.mockctl/config.db`.
- *Buckets:* Ensure initialization of logical buckets (`auth_bucket`, `settings_bucket`, `telemetry_bucket`, `metadata_bucket`).
- *Multi-Process File Locking:* Acquire the lock with a strict **1-second timeout**. If denied, fallback to **Read-Only Mode** (skipping telemetry writes).
- *Migrations:* Compare the binary version against `db_version` in `metadata_bucket` and sequentially apply schema migrations.
- *Limits & JWTs:* Enforce a 50MB hard size limit on `config.db` and ensure License keys are handled as RS256 JWTs.

### Step 6: Implement Tests (Fakes & Integration)
Ensure the stores function flawlessly under extreme conditions while adhering to PKS-029 rules.
- Files: `internal/adapter/db/memory_state_test.go` and `internal/adapter/db/bbolt_system_test.go`
- *Manual Fakes:* As per PKS-029, testing domain logic must utilize "Manual Fakes" (`FakeStateStore`) rather than dynamic mock generation libraries.
- *Integration Tests:* `bbolt_system_test.go` MUST be treated as an Integration Test. It must generate a unique temporary database file for each test using `t.TempDir()` and ensure cleanup via `t.Cleanup()`.
- *Concurrency Stress Testing:* Spawn hundreds of parallel goroutines mutating the `StateStore` to explicitly verify that `fatal error: concurrent map writes` does not occur.

---

## ⚠️ Known Edge-Cases & Warnings
1. **Dumb Relationships:** The StateStore does not enforce foreign keys or cascading deletes. Deleting a parent resource does not delete nested flat paths.
2. **Type Coercion on Scans:** The StateStore lacks a SQL engine. Queries will require O(N) full collection scans with aggressive type coercion for query parameters.
3. **Telemetry Compaction:** Future steps will require a background goroutine to aggressively TTL (Time-To-Live) delete telemetry logs older than 30 days to prevent disk bloat.

## ✅ Expected Outcome
The `internal/adapter/db/` package will expose production-ready stores. The state store will handle fast, concurrent schema-agnostic API simulation with OOM protection, while the system store safely manages physical `.db` files with migration handling and lock fallbacks.

**Status:** ✅ Approved
