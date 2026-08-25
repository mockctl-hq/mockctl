# Implementation Plan 004: HTTP Mock Server Engine & Runtime

**Status:** ✅ Approved  
**Focus:** Implementing the strictly isolated `RuntimeEngine`, the `HTTPServer` presentation layer, and the `internal/app` Composition Root.

---

## 1. Objective

To build the complete Mock Server Execution pipeline. This design strictly adheres to **EDL-025** (`net/http` + `chi`), **EDL-026** (HTTP Layer Separation), and the architectural boundary rules defined in **PKS-024** and **PKS-025**.

---

## 2. Step-by-Step Implementation

### Step 1: The Composition Root & Boot Sequence
**File:** `internal/app/app.go`

- **Purpose:** The `App.StartServer()` function acts as the Composition Root. It wires all dependencies together before starting the server.
- **Boot Sequence:**
  1. **Port Resolution:** Resolve the target port (Default `8080` unless overridden by project config).
  2. **Admin Token Generation:** Generate a cryptographically secure random 32-byte hex string (strictly using `crypto/rand`, avoiding predictable `math/rand` UUIDs) on boot, save it to the `SystemStore` (`SaveAuthToken`), and print it to the console so the user can access the Admin API.
  3. **Wiring:** Instantiate the `RuntimeEngine` (injecting `StateStore`, `RuntimeDefinition`, `ValueProvider`, `ChaosEvaluator`, and `Clock`).
  4. **Server Initialization:** Instantiate the `HTTPServer`, injecting the `RuntimeEngine` and `SystemStore`.
  5. **Graceful Shutdown Hook:** Listen for `syscall.SIGINT` and `syscall.SIGTERM` using a Go channel. On signal, call `HTTPServer.Shutdown(ctx)` with a **`5*time.Second` timeout context** to gracefully drain connections (preventing indefinite hangs) and instruct the `StateStore` to dump its current memory map to `temp-state.json`.

### Step 2: Presentation Boundary (`HTTPServer`)
**File:** `internal/adapter/http/server.go`

- **Purpose:** Wraps `chi` routing and exposes the server to the OS.
- **Struct Definition Blueprint:**
  ```go
  type HTTPServer struct {
      router      *chi.Mux
      systemStore ports.SystemStore
      engine      http.Handler // The RuntimeEngine
      httpServer  *http.Server
  }
  func (s *HTTPServer) Start(port int) error
  func (s *HTTPServer) Shutdown(ctx context.Context) error
  ```
- **DomainError Blueprint:**
  ```go
  type DomainError struct {
      Code       string
      Message    string
      HTTPStatus int
      Err        error
  }
  ```
- **Middleware Chain (Strict Order in Chi):**
  1. `Recovery`: Catch panics and return standard `500 Internal Server Error`.
  2. `RequestID`: Propagate unique IDs via context.
  3. `Logger`: Integration with `log/slog` for structured request logging. **Security Rule: MUST scrub/redact the `Authorization` header to prevent token leakage in plain-text logs.**
  4. `Metrics`: Track latency and status codes.
  5. `DomainErrorHandler`: A specialized middleware that translates `DomainError` structs into the standardized JSON envelope.

### Step 3: Admin API & Security (`HTTPServer` Level)
**File:** `internal/adapter/http/admin_routes.go`

- **Owner:** The `HTTPServer` owns the Admin API because it requires `SystemStore` access.
- **Reserved Namespace:** All routes under `/__mockctl/*` are mapped here.
- **Security Rules:**
  - **Localhost Binding:** Reject any request not originating from `127.0.0.1` (`403 Forbidden`).
  - **Authentication:** Must include an `Authorization: Bearer <token>` header matching the local admin token. **Security Rule (PKS-028): The comparison MUST use `crypto/subtle.ConstantTimeCompare` to prevent timing attacks.**
  - **Rate Limiting:** Enforce a strict limit of **100 requests/second** per client.
  - **Headers & CORS:** Require `Accept-Version: v1` header. Inject `Access-Control-Allow-Origin: *`.
  - **Content-Type Validation:** Any `POST`/`PUT`/`PATCH` request without `Content-Type: application/json` MUST be rejected with `415 Unsupported Media Type`.

### Step 4: Core Executor (`RuntimeEngine`) & OpenAPI Validation
**File:** `internal/runtime/engine.go`

- **Strict Isolation (EDL-026):** The `RuntimeEngine` resides in the core and implements the standard `http.Handler`. It does **not** import `chi`.
- **Struct Definition Blueprint:**
  ```go
  type RuntimeEngine struct {
      logger        shared.Logger
      definition    *generator.RuntimeDefinition
      state         ports.StateStore
      chaos         ports.ChaosEvaluator
      valueProvider ports.ValueProvider
      clock         shared.Clock
  }
  func (e *RuntimeEngine) ServeHTTP(w http.ResponseWriter, r *http.Request)
  ```
- **Catch-All Delegation:** The `HTTPServer` mounts the `RuntimeEngine` at the root wildcard `/*`.
- **Validation Boundary & DoS Protection:** Inside `ServeHTTP`, before processing any data mutation:
  1. The engine MUST wrap the request body with `http.MaxBytesReader` to prevent **JSON Bomb / RAM DoS attacks**. The limit MUST NOT be hardcoded; it should default to `5MB` but remain configurable via the `AppConfig` to support edge-case large payloads.
  2. The engine MUST use `kin-openapi` (`kin-openapi/openapi3filter` or similar) to validate the incoming JSON payload against the expected schema defined in the `RuntimeDefinition`. Invalid data returns `400 Bad Request`.
- **Media Type Validation:** Similar to the Admin API, incoming mock data requests without `application/json` must be rejected with `415 Unsupported Media Type`.

### Step 5: Route Matching & The Mutation Lifecycle
**File:** `internal/runtime/handlers.go`

- **Context Propagation:** EVERY call to `StateStore` or `ChaosEvaluator` MUST receive `r.Context()` to support client cancellation.
- **Flat Path Strategy:** Dynamic path parameters (e.g., `/{id}`) are extracted and used to build a Flat Path string for the `StateStore` (e.g., `users/123`).
- **GET:** Read from `StateStore`. If a query parameter exists (e.g., `?status=active`), perform a full collection `List()` and filter in-memory. If missing entirely, fallback to `ValueProvider` to generate template data.
- **POST / PUT:**
  - **ID Generation:** Check payload for `id`. If missing, request a new UUID or Auto-Increment integer from the injected **`ValueProvider`**.
  - **Auto-Timestamps:** Inject `createdAt` and `updatedAt` in RFC3339 format using the `Clock` interface, **ONLY IF** these fields are explicitly defined in the user's OpenAPI schema.
  - Insert/Update in `StateStore` using `r.Context()`.
- **PATCH:** Use Deep-Merge logic in `StateStore`.
- **OOM Safety:** Catch `ErrLimitReached` and respond with `HTTP 429` or `HTTP 507`.
- **Response Formatting:** Before calling `w.Write()` with the generated JSON (using standard `encoding/json`), iterate over `ResponseTemplate.Headers` and inject each custom header into `w.Header()`.

### Step 6: Testing Strategy
**Files:** `internal/runtime/*_test.go` & `internal/adapter/http/*_test.go`

- **Time & Data Independence:** Time logic must use a `FakeClock`. Data generation must use a `FakeValueProvider`.
- **No Real Ports in Unit Tests:** Strictly use `httptest.NewRecorder()` and `httptest.NewRequest()`.
- **E2E Integration:** Use Go build tags (`//go:build e2e`) for E2E tests that start the actual `HTTPServer` on a random port.
- **Goroutine Leak Prevention:** Use `go.uber.org/goleak` in E2E tests to verify graceful shutdown.
- **Concurrency:** All execution flows must pass under `go test -race`.

---

## 3. Directory & Package Blueprint
The following files will be created or modified during this implementation:

```text
internal/
├── app/
│   └── app.go                 // App orchestrator & Composition Root (Graceful shutdown, Wiring)
├── adapter/
│   └── http/
│       ├── server.go          // HTTPServer struct, Chi initialization, Middleware pipeline
│       ├── admin_routes.go    // /__mockctl/* routes, CORS, Rate Limiter
│       └── server_test.go     // E2E and HTTP tests
├── core/
│   └── ports/
│       └── data.go            // Interfaces: ValueProvider, ChaosEvaluator
└── runtime/
    ├── engine.go              // RuntimeEngine struct & ServeHTTP implementation (Validation)
    ├── handlers.go            // StateStore mutation logic, r.Context() propagation
    └── engine_test.go         // Logic tests using httptest, FakeClock, FakeValueProvider
```

---

## 4. Pre-requisites for Implementation
Before coding, the following external dependencies will need to be downloaded via `go mod tidy`:
- `github.com/go-chi/chi/v5`
- `github.com/getkin/kin-openapi` (For Schema Validation)
- `golang.org/x/time/rate` (For strict Admin API Rate Limiting)
- `go.uber.org/goleak` (For Testing)

---
*Please review this draft and provide any feedback, refinements, or corrections before marking as "Approved".*
