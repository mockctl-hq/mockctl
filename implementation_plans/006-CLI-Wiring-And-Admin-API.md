# 🛠️ IMP-006: Daemon Wiring & Multi-Project Admin API

> **Plan ID:** IMP-006  
> **Epic:** Core Application Wiring & Administration  
> **Status:** 🟢 Approved  
> **Complexity:** Very High  
> **Target Branch:** `feat/daemon-and-admin-api`  

---

## 🎯 1. Objective
Mock:ctl is transitioning into an Enterprise **Master Daemon** architecture. This plan establishes a central Daemon server on a single port (`8080`), utilizing a Thread-Safe API Gateway for dynamic URL routing (`/{projectName}/*`). Crucially, this implementation strictly complies with all PKS engineering boundaries, ensuring flawless dependency injection, context-first design, and proper module isolation as defined in PKS-020 to PKS-030.

---

## 🏗️ 2. Execution Phases

### 🚧 Phase 0: PKS-Compliant Interface Updates
**Target Files:** `internal/shared/`, `internal/storage/`, `internal/project/`, `internal/data/`

- **Task 0.1 (ValueProvider Implementation):** Create `internal/data/provider.go` to implement `ValueProvider` using `gofakeit` for fake data generation.
- **Task 0.2 (Chaos API Context-First):** Update the chaos evaluator interface to include `UpdateConfig(ctx context.Context, errorRate int, latencyMs int)` per PKS-028 standards.
- **Task 0.3 (Separation of State vs. Overrides):** As per PKS-023, do NOT use `MemoryStateStore` for static overrides. Instead, define an `Overrides` map within the `WorkspaceContext` managed by the `internal/project/` package. Update the `MockGenerationEngine` to prioritize these static overrides during runtime execution.
- **Task 0.4 (Hybrid Project Data Model):** Define the `Project` struct in `internal/project/` containing both `OpenAPISpec` (string) and `CustomEndpoints` (array). Update `internal/storage/` to serialize this via BBolt.

### 📦 Phase 1: Clean Daemon Initialization (EDL-014)
**Target File:** `cmd/mockctl/daemon.go`, `internal/app/app.go`

- **Task 1.1 (Dumb CLI):** Create `mockctl daemon` Cobra command. It performs NO business logic (per EDL-014). It only parses the `--port` flag and delegates execution to `app.StartDaemon(ctx, port)`.
- **Task 1.2 (App Bootstrapping & Cross-Platform Paths):** Inside `app.StartDaemon()`, use `filepath.Join(userHome, ".mockctl")` (PKS-028 compliant) to initialize the directory, instantiate `internal/storage/bbolt`, and acquire the `daemon.pid` lock file.
- **Task 1.3 (Non-Blocking Re-hydration):** `App` must read all saved `Project` structs from BBolt in a fast read-transaction. After releasing the DB lock, it iterates over the projects, parses specs, and compiles `RuntimeEngine`s.
- **Task 1.4 (TUI-Safe & Docker-Safe Logs):** Refactor `setupLogger()`. If in Docker/headless mode (PKS-030), write Structured JSON to `os.Stdout`. Otherwise, redirect to `filepath.Join(home, ".mockctl", "daemon.log")` to prevent UI corruption.

### 🧩 Phase 2: Application Orchestration & Dynamic Routing (PKS-020)
**Target File:** `internal/app/app.go`, `internal/runtime/server.go`

- **Task 2.1 (Downward Dependency Flow):** The `HTTPServer` (Infrastructure Layer) must NOT parse specs. Instead, inject the `App` orchestrator (Application Core) into the `HTTPServer`. When a user uploads a project, the HTTP handler simply delegates down: `App.CreateProject(ctx, payload)`.
- **Task 2.2 (ProjectGateway Router):** Inside `internal/runtime/server.go`, implement a custom `ProjectGateway` Handler with a `sync.RWMutex` protected `map[string]*RuntimeEngine` to dynamically mount `/{projectName}/*` routes.
- **Task 2.3 (Thread-Safe Hot-Swapping):** When `App` re-compiles a project (e.g., via Visual Builder), it instructs the `ProjectGateway` to swap the engine pointer under a Write Lock (`mu.Lock()`), ensuring zero dropped in-flight requests.
- **Task 2.4 (Graceful Storage Shutdown):** The `SIGTERM` trap inside `app.go` must delegate state persistence down to the `internal/storage/` layer abstraction, avoiding direct file I/O in the runtime handler (PKS-023).

### 🌐 Phase 3: Project Management Admin API (PKS-027)
**Target File:** `internal/runtime/admin_routes.go`

- **Task 3.1 (Project Lifecycle APIs):**
  - `GET /__mockctl/projects`
  - `POST /__mockctl/projects`: Accepts `multipart/form-data` for 10MB+ YAML streaming. Slugs the project name (`[a-z0-9-]+`), and calls `App.CreateProject(ctx, file)`.
  - `DELETE /__mockctl/projects/{name}`
- **Task 3.2 (Endpoint & Configuration APIs):**
  - `POST /__mockctl/projects/{name}/endpoints`: Visually adds a single custom endpoint. Calls `App.AddEndpoint(ctx, name, endpoint)` which handles BBolt saving and engine re-compilation.
  - `POST /__mockctl/projects/{name}/overrides`: Sets static overrides via the `internal/project/` package, not the state store.
  - `POST /__mockctl/projects/{name}/state/reset`: Flushes live CRUD data inside `internal/storage/memory`.
  - `PATCH /__mockctl/projects/{name}/chaos`: Calls `UpdateConfig(ctx)` on the specific project's chaos evaluator.

*(Note: Real-Time SSE Events are reserved for **IMP-007**. The Interactive Terminal Dashboard (TUI) is reserved for **IMP-008**).*

---

## ⚖️ 3. Constraints & Rules (PKS Compliance)
- **Response Format (PKS-027):** All Admin APIs MUST return the standard JSON Envelope (`{ "success": true/false, "data": ... }`).
- **Security Constraint (PKS-027):** Localhost binding, Rate Limiting (per-IP), and Auth Token (`filepath.Join` to `admin.token`) MUST apply to all `/__mockctl/*` endpoints.
- **Context-First (PKS-028):** All business logic methods across architectural boundaries MUST accept `ctx context.Context`.
- **SECURITY (CORS Hijacking):** The `/__mockctl/` Admin namespace MUST strictly reject wildcard CORS to prevent CSRF attacks. Wildcard CORS is only permitted for user mocked routes.
- **SECURITY (SSRF Protection):** The OpenAPI parser must disable or strictly whitelist external HTTP `$ref` fetching to prevent Server-Side Request Forgery.
- **SECURITY (Path Traversal/LFI):** `ProjectName` must be strictly sanitized using a `[a-z0-9-]+` slugification regex before interacting with the file system or BBolt.
- **SECURITY (Billion Laughs OOM):** The YAML parser must enforce strict memory expansion limits and recursion limits to prevent Out-Of-Memory (OOM) crashes via malicious payloads.
- **SECURITY (Token Leakage):** The TUI and external tools MUST pass the local token exclusively via the `Authorization: Bearer <token>` header, never via CLI arguments or Query Parameters.
---

## ⚙️ 4. Operational, Concurrency & Architecture Safeguards
- **Routing Clashes:** Reserved names (e.g., `__mockctl`, `admin`, `system`) MUST be blacklisted in the Project Validator to prevent namespace collision with the Admin API.
- **Go Cyclic Dependency Prevention:** `internal/runtime` MUST NOT import `internal/app`. It must define a `ProjectManager` interface that `app.App` implements to invert the dependency.
- **Lock Contention / Deadlock Avoidance:** The `ProjectGateway` MUST acquire `mu.RLock()`, copy the engine pointer, and immediately `mu.RUnlock()` BEFORE calling `engine.ServeHTTP()`. It must never hold the lock during request execution.
- **Memory Leaks during Hot-Swap:** When swapping the `RuntimeEngine` pointer, the `context.Cancel()` of the old engine MUST be called to terminate any hanging zombie goroutines (e.g., SSE, Chaos timers).
- **Graceful Shutdown Hangs:** The `SIGTERM` trap MUST enforce a strict timeout (e.g., 3-5 seconds). If requests (like chaos-delayed responses) do not finish, the server must forcefully terminate to avoid freezing the CLI.
- **Compilation Bottlenecks & Rollbacks:** Engine compilation/validation MUST happen BEFORE opening a BBolt write-transaction. If compilation fails, the database remains untouched. Bulk endpoint additions MUST be debounced.
- **Termux `/tmp` Stream Parsing:** The Admin API MUST parse large `multipart/form-data` uploads using streaming (`io.Reader`) instead of `ParseMultipartForm` to bypass memory/disk issues in restricted environments like Android/Termux.
- **Port Conflicts:** If `8080` is in use, the Daemon MUST output a clean, user-friendly error specifying the blocked port, and optionally suggest or fallback to an alternate port, instead of a raw panic.

---

## 🧪 5. Verification & Testing Strategy
- [ ] **Daemon Lock Protection:** Running `mockctl daemon` twice in two terminals immediately panics the second one due to `.pid` lock.
- [ ] **Non-Blocking Startup:** Restarting the Daemon automatically restores all projects from BBolt without locking the database during compilation.
- [ ] **Dynamic Routing Gateway:** `POST /__mockctl/projects` successfully routes traffic to `/auth/*` instantly.
- [ ] **Log Isolation:** Server logs conditionally output to `daemon.log` (local) or `os.Stdout` (Docker).
- [ ] **API Security:** Calling `http://localhost:8080/__mockctl/projects` without a valid token returns `401 Unauthorized`.
