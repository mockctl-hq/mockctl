# 📏 PKS-028 — Coding Standards

> **Project:** Mock:ctl
>
> **Document ID:** PKS-028
>
> **Version:** 1.0
>
> **Status:** Approved
>
> **Owner:** Upen Tudu
>
> **Authors:** Upen Tudu & Antigravity
>
> **Created:** 2026-08-16
>
> **Last Updated:** 2026-08-16
>
> **Category:** Engineering
>
> **Priority:** High

---

# 📖 Executive Summary

The Coding Standards document establishes the definitive implementation conventions for the Mock:ctl Go codebase. While architectural documents (PKS-020 to PKS-027) dictate *what* we build and *how* the systems interact, this document dictates exactly *how the code is written*.

By strictly enforcing Go idioms, domain-driven boundaries, error handling mechanisms, and advanced concurrency safety, we ensure that the Mock:ctl codebase remains highly readable and maintainable. Furthermore, by standardizing network timeouts, cryptography, database transactions, and structured logging, we guarantee that the system operates at an enterprise-grade level, free of typical software bloat and vulnerabilities.

---

# 🎯 Purpose

The objectives of this document are to:
- Establish a uniform, `gofmt`-governed coding style across the entire Mock:ctl backend.
- Define strict rules for memory safety, advanced concurrency (goroutines/mutexes), and global state.
- Standardize Error Handling to cleanly map to the Admin API JSON responses.
- Define interface boundaries, database transaction limits, and testing structures.
- Enforce network safety (HTTP timeouts) and cryptographic security standards.
- Ensure automated tools (`golangci-lint`) act as the final authority on code syntax and quality.

---

# 📌 Scope

This document applies exclusively to the Go backend (`Mock:ctl`).
It covers:
- Core Go Idioms (Interfaces, Structs, Pointers)
- State & Concurrency Rules (Mutexes & Goroutines)
- Error Handling & Propagation
- Testing Standards
- Package & Variable Naming
- Logging & Diagnostics
- Configuration Injection
- JSON Serialization
- Cross-Platform File System usage
- Security & Code Documentation
- Dependency Management & HTTP Timeouts

It **does not** cover coding standards for the future Flutter UI or the Cloud API, which will be addressed in separate platform-specific documents.

---

# 🛠️ 1. Core Go Idioms & Formatting

Mock:ctl embraces the philosophy of "Idiomatic Go" (Clear is better than clever). The rules in this section are directly derived from the official Engineering Decision Log.

## 1.1 Code Style Governance (EDL-033)
Human debates over formatting are banned. `gofmt` is the single source of truth for code style.
- No personal formatting preferences.
- No style debates in reviews.
- AI-generated code must follow identical standards.

## 1.2 Code Formatting Standard (EDL-031)
- All code MUST be formatted using `gofmt` before committing.
- Import organization MUST be automated using `goimports`.
- Formatting changes are not manually reviewed.

## 1.3 Static Analysis Standard (EDL-032)
The CI pipeline will strictly enforce `golangci-lint` to catch anti-patterns.
To maintain high code quality without excessive developer friction, only the following curated linters are approved:
- `govet`
- `staticcheck`
- `errcheck`
- `ineffassign`
- `unused`
- `revive`

*Rules:*
- New linters require engineering approval.
- Lint suppressions require documented justification using the specific syntax: `//nolint:errcheck // reason: intentional drop`. Naked `//nolint` is strictly banned.

## 1.4 "Accept Interfaces, Return Structs"
To keep the application highly testable, functions and methods should accept narrow interfaces but return concrete structs.
- **Good:** `func ProcessData(reader io.Reader) *Result`
- **Bad:** `func ProcessData(file *os.File) ResultInterface`

## 1.5 Context-First Design
Every function that crosses an architectural boundary (e.g., Network I/O, File System, Database) MUST accept a `context.Context` as its very first parameter to support cancellation and timeouts.
- **Example:** `func (s *BoltSystemStore) SaveAuthToken(ctx context.Context, token string) error`

## 1.6 Enums & Custom Types
Go lacks native enums. Never use naked `string` or `int` to represent finite states (e.g., status codes, roles). Always define a custom type and use `iota` or typed constants.
- **Example:** `type ServerStatus string; const StatusRunning ServerStatus = "running"`

## 1.7 Pointers vs Values (Memory Management)
- **By Value:** Pass small structs (like configurations) by value to reduce Garbage Collection overhead.
- **By Pointer:** Pass large structs, or structs that contain mutable state (`sync.RWMutex`), by pointer (`*Struct`). Never copy a Mutex by passing it by value.

## 1.8 Memory Pre-allocation (Slices & Maps)
When the final size of a slice or map is known in advance, it MUST be pre-allocated using `make([]T, 0, capacity)` or `make(map[K]V, capacity)`. This prevents expensive runtime memory reallocations and reduces Garbage Collector pauses under high traffic.

---

# 🚫 2. State & Concurrency Rules

Because Mock:ctl serves high-throughput HTTP traffic simulating user APIs, data races are catastrophic.

## 2.1 Strictly No Global State
Global variables (`var Data = make(map[string]any)`) are strictly forbidden outside of the `main.go` bootstrap process.
- All dependencies must be injected via constructors (e.g., `NewRuntimeEngine(store StateStore)`).
- `init()` functions should be avoided unless absolutely necessary for registering decoders/drivers.

## 2.2 Mutex Protection
The `StateStore` (In-Memory Database) holds mutable data that can be read and written simultaneously by hundreds of incoming HTTP requests.
- All maps and slices inside the `StateStore` MUST be protected by a `sync.RWMutex`.
- **Rule:** Keep lock holding times as short as possible. Never perform heavy I/O operations while holding a lock.

## 2.3 Database Transactions (bbolt)
The `SystemStore` uses `bbolt`, which strictly locks the database during transactions.
- **Micro-Transactions:** Transactions must execute in microseconds. NEVER perform network I/O, heavy computation, or HTTP calls while holding a `bolt.Tx`.
- **Read vs Write:** Use `View()` for read-only operations to allow concurrent readers. Use `Update()` exclusively when modifications are required.

---

# 🚨 3. Error Handling & Propagation

Error handling must be explicit and informative, avoiding naked panics.

## 3.1 Domain Errors
Do not return simple strings (`fmt.Errorf("bad request")`) across boundary layers. Instead, use the structured `DomainError` which maps cleanly to the Admin API JSON response (defined in PKS-027).

```go
type DomainError struct {
    Code    string // e.g., "ERR_VALIDATION_FAILED"
    Message string // Human-readable message
    Status  int    // HTTP Status (e.g., 400)
    Err     error  // The underlying Go error (for internal logging)
}
```

## 3.2 Error Wrapping
When an error bubbles up from a lower layer (e.g., the FileSystem layer), it must be wrapped using `%w` so the caller can inspect the root cause using `errors.Is` or `errors.As`.
- **Example:** `fmt.Errorf("failed to read spec: %w", err)`

## 3.3 No Naked Panics
The `panic()` function is strictly forbidden in application logic. The only acceptable use of `panic` is during application startup (e.g., if the `config.db` cannot be opened or a vital dependency fails to boot).

## 3.4 Defer Error Handling
Errors returned by deferred functions (e.g., `defer file.Close()`) must not be silently ignored. They must be explicitly handled, logged, or named-returned to ensure data integrity during resource cleanup.

## 3.5 Panic Recovery Middleware
While application code must never panic, third-party libraries might. The HTTP router MUST include a top-level `Recovery Middleware` that catches any unhandled panics, logs the stack trace as an `ERROR`, and returns a safe `500 Internal Server Error` to the client, preventing the entire Go process from crashing.

---

# 🧪 4. Testing & Assertions

Mock:ctl enforces a strict testing philosophy derived directly from the Engineering Decision Log to ensure long-term code quality.

## 4.1 Testing Foundation & Assertions (EDL-028)
- Mock:ctl shall use the Go Standard Library `testing` package as its foundation.
- The `testify` library shall be used only for assertions (`assert.Equal`) and test utilities.

## 4.2 Testing Philosophy (EDL-029)
- Unit tests are the primary testing layer. Integration tests verify adapters. End-to-End tests verify complete workflows.
- Every bug fix requires a regression test.
- **Manual Fakes:** Manual fakes are *preferred* over heavy mocking frameworks (`gomock`). Write simple struct implementations of interfaces for tests.
- **Table-Driven Tests:** All complex business logic MUST be tested using the Table-Driven Testing pattern.
- **Golden Tests:** Validate generated outputs (e.g., JSON responses) using Golden tests.

## 4.3 Coverage Policy (EDL-030)
- Coverage is a confidence indicator, not a hard mathematical target.
- Meaningful tests are preferred over artificial coverage. Low-value tests created solely to increase percentages are banned.
- Critical business logic must maintain very high coverage, while generated code is excluded from coverage metrics.

---

# 🏷️ 5. Naming Conventions & Packages

Clear naming is critical for codebase navigation.

## 5.1 Package Naming & Layout
- **Internal by Default:** All proprietary business logic MUST reside inside the `internal/` directory (as per EDL-005). Only packages intended for public consumption may live outside `internal/`.
- Packages must be named with short, single-word, lowercase names (e.g., `storage`, `runtime`, `spec`).
- **Forbidden:** Never create `util`, `common`, `helper`, or `misc` packages. Code must be grouped by its domain purpose.

## 5.2 Variable & Interface Naming
- **Acronyms:** Acronyms must be fully uppercase or fully lowercase (e.g., `UserID`, not `UserId`; `ServeHTTP`, not `ServeHttp`).
- **Interfaces:** Interfaces with a single method should have an `-er` suffix (e.g., `FileReader`, `StateWriter`).
- **Variable Scope:** Keep variable names short in limited scopes (e.g., `r` for `http.Request`, `i` for index) and descriptive in global/struct scopes.

---

# 📝 6. Logging Standards (EDL-018, EDL-019)

Mock:ctl relies heavily on structured logging for observability.

## 6.1 Structured Format
- All logs must be output as structured JSON objects containing context fields. Naked `fmt.Println` or `log.Printf` are strictly forbidden outside of CLI output commands.

## 6.2 Log Levels
- **INFO:** Lifecycle events (e.g., `server_started`, `db_migrated`).
- **DEBUG:** Verbose execution paths (e.g., `route_matched`, `spec_parsed`).
- **WARN:** Handled edge cases that did not crash the request (e.g., `file_not_found_using_fallback`).
- **ERROR:** Operations that failed and triggered a 5xx response.

---

# ⚙️ 7. Configuration & Environment Injection

- **No Deep `os.Getenv`:** Business logic and internal packages (`internal/runtime`, `internal/storage`) MUST NEVER call `os.Getenv()` or parse flags.
- **Constructor Injection:** All configurations must be resolved at the edge (`main.go` or `cmd/`) and passed into components via explicit structs (e.g., `config.RuntimeConfig`).

---

# 🗃️ 8. JSON Serialization & Struct Tags

Mock:ctl handles APIs, meaning JSON serialization must be highly predictable.

- **Naming Convention:** All JSON struct tags must use `snake_case` (e.g., `json:"chaos_level"`).
- **Optional Fields:** For fields that can be empty/null, use pointers and the `omitempty` tag (e.g., `Description *string json:"description,omitempty"`). This prevents Go's zero-values from accidentally polluting API responses.
- **Timestamps:** All `time.Time` fields MUST be serialized in ISO 8601 / RFC3339 format (e.g., `2026-08-16T15:04:05Z`). Never use Unix epoch integers for JSON API boundaries.

---

# 🔄 9. Advanced Concurrency: Goroutines & Context

Beyond Mutexes, background operations (like Cloud Sync) must be managed carefully.

- **No Orphaned Goroutines:** Never start a goroutine without a clear plan for how and when it will stop.
- **Context Cancellation:** Every long-running goroutine MUST listen to a `context.Context` channel (`<-ctx.Done()`) to ensure clean shutdown when the Mock:ctl server is stopped.
- **Errgroup over WaitGroup:** When managing multiple concurrent tasks (e.g., fan-out HTTP requests), use `golang.org/x/sync/errgroup` instead of a naked `sync.WaitGroup` to properly capture and propagate errors from background workers.

---

# 📁 10. File System & Cross-Platform Rules

Mock:ctl is designed for Linux, macOS, and Windows.
- **Path Construction:** NEVER use string concatenation with slashes (e.g., `dir + "/" + file`). Always use `filepath.Join(dir, file)` to ensure Windows compatibility.
- **Permissions:** When creating files or directories, use least-privilege modes (e.g., `0600` for tokens, `0755` for directories). Never use `0777`.

---

# 🔒 11. Security & Cryptography Practices

- **Randomness:** For any security-sensitive generation (Tokens, IDs), strictly use `crypto/rand`. `math/rand` is banned for security payloads.
- **Timing Attacks:** When comparing sensitive data (e.g., Admin Tokens, JWT signatures), NEVER use the standard `==` operator. You MUST use `subtle.ConstantTimeCompare` to prevent timing attacks.
- **Memory Safety:** Sensitive strings (like Cloud API Keys or plain License keys) must not be logged.

---

# 📚 12. Code Documentation (GoDoc)

- Every exported identifier (struct, interface, function, constant) MUST have a GoDoc comment directly above it.
- The comment must be a complete sentence that begins with the name of the identifier.
  - *Good:* `// RuntimeEngine processes incoming HTTP traffic.`
  - *Bad:* `// this processes traffic`

---

# 📦 13. Dependency Management (go.mod)

- **Clean Modules:** `go mod tidy` MUST be run before every commit to ensure `go.sum` and `go.mod` remain clean.
- **No Replacements:** The `replace` directive in `go.mod` is strictly forbidden in production code. It may only be used temporarily during local debugging.

---

# 🌐 14. Server Lifecycle, Timeouts & Safety

Network operations and process lifetimes must be managed safely.

- **No Default Clients:** The default Go `http.Client` has no timeout and can hang forever. NEVER use `http.Get()`. Always instantiate a custom client with an explicit `Timeout` (e.g., `client := &http.Client{Timeout: 10 * time.Second}`).
- **Server Timeouts:** When bootstrapping the Mock:ctl `http.Server`, always define `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` to prevent malicious clients from holding connections open indefinitely (Slowloris attacks).
- **Graceful Shutdown:** The `main.go` bootstrap process MUST listen for OS Signals (`SIGINT`, `SIGTERM`). Upon receiving a signal (e.g., `CTRL+C`), it must stop accepting new HTTP requests, wait for active requests to finish, and safely call `Close()` on the `bbolt` database to prevent corruption.

---

# 📌 Conclusion

By adhering strictly to these coding standards, the engineering team guarantees that the Mock:ctl Go backend will remain performant, predictable, and highly stable. The absolute ban on global state, the enforcement of Context cancellation, strict HTTP timeouts, micro-transactions for `bbolt`, and the standardized `DomainError` structure form the foundation necessary to confidently scale the application toward its SaaS future without accumulating technical debt or security vulnerabilities.

---

# 🔗 Related Documents

**Foundation**

- PKS-000 — Repository Blueprint
- PKS-002 — Documentation Style Guide

**Engineering**

- PKS-020 — System Architecture
- PKS-024 — Component Architecture
- PKS-025 — Software Design Document (Master SDD)
- PKS-027 — API Design Guidelines

**Next Document**

- PKS-029 — Testing Strategy

---

# 📜 Revision History

| Version | Date | Description |
|----------|------------|----------------------------------------------|
| 1.0 | 2026-08-16 | Approved |

---

# ✅ Approval Checklist

- Executive summary completed
- Go formatting tools (`gofmt`, `goimports`) mandated (EDL-031, EDL-033)
- Static Analysis (`golangci-lint` with approved linters) enforced (EDL-032)
- Idiomatic Go rules ("Accept Interfaces, Return Structs") established
- Concurrency rules (Mutexes and Goroutine Contexts) and Global State ban enforced
- Structured Error Handling (`DomainError`) defined
- Testing philosophy, fakes over mocks, and coverage policy aligned with EDL-028, 029, 030
- Naming, Logging, Configuration, JSON, File System, Security, and GoDoc rules established
- Advanced rules (bbolt transactions, HTTP timeouts, Pointers, go.mod, Enums) enforced
- Formatting follows PKS style guide
- Conclusion section included

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** ✅ Reviewed & Approved

**Architecture Status:** ✅ Established

**Next Document:** **PKS-029 — Testing Strategy**
