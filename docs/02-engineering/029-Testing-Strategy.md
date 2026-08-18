# 🧪 PKS-029 — Testing Strategy

> **Project:** Mock:ctl
>
> **Document ID:** PKS-029
>
> **Version:** 1.0
>
> **Status:** Approved
>
> **Owner:** Upen Tudu
>
> **Authors:** Upen Tudu & Antigravity
>
> **Created:** 2026-08-18
>
> **Last Updated:** 2026-08-18
>
> **Category:** Engineering
>
> **Priority:** High

---

# 📖 Executive Summary

The Testing Strategy document defines how the engineering team will validate the correctness, stability, performance, and security of the Mock:ctl Go backend. Since Mock:ctl is an API simulation engine designed to help other developers test their applications, Mock:ctl itself must be impeccably reliable. 

This document operationalizes the testing decisions made in the Engineering Decision Log (EDL-028, EDL-029, EDL-030) by establishing strict rules for test layering (Unit, Integration, E2E), mocking paradigms (manual fakes, time freezing), test design (Table-Driven, Golden Tests, Parallel Execution), advanced paradigms (Fuzzing, Benchmarks, Leak Prevention), and code coverage expectations.

---

# 🎯 Purpose

The objectives of this document are to:
- Establish the official Go testing frameworks and libraries.
- Define the boundaries and responsibilities of Unit, Integration, and End-to-End (E2E) testing layers.
- Manage slow test execution via Go Build Tags.
- Standardize the approach to faking and mocking external dependencies.
- Outline best practices for writing deterministic, table-driven tests.
- Mandate Concurrency, Performance (Benchmarks), and Security (Fuzzing) test standards.
- Set realistic and meaningful code coverage policies.

---

# 📌 Scope

This document applies exclusively to the Go backend (`Mock:ctl`).
It covers:
- Core Testing Tools & Time Mocking
- The Testing Pyramid & Build Tags
- Mocking Strategy (Fakes vs. Mocks)
- Test Design Patterns (Parallel Execution & Data Management)
- Performance, Concurrency, Leak Prevention, and Fuzz Testing
- Coverage Metrics
- Regression Testing Policies

It **does not** cover UI testing for the future Flutter dashboard or load testing for the Cloud synchronization API.

---

# 🛠️ 1. Testing Foundation & Tools (EDL-028)

Mock:ctl relies heavily on the Go standard library to minimize dependency bloat and ensure forward compatibility.

## 1.1 The Standard Library
All tests MUST be written using the standard Go `testing` package. Test files must live adjacent to the code they test (e.g., `parser_test.go` next to `parser.go`).

## 1.2 Assertions
While `testing` is the foundation, standard `if err != nil` assertions can become highly verbose. Therefore, Mock:ctl officially uses `github.com/stretchr/testify/assert` and `testify/require` for all test assertions.
- **Rule:** Use `assert` for checks that allow the test to continue if they fail. Use `require` for fatal checks (like checking if an error is nil before dereferencing a pointer) where a failure should immediately abort the test.

---

# 🏗️ 2. The Testing Layers (EDL-029)

Mock:ctl follows a strict testing pyramid.

## 2.1 Unit Tests (The Primary Layer)
Unit tests verify isolated business logic (e.g., OpenAPI specification parsing, random data generation).
- **Execution Speed:** Must execute in milliseconds.
- **Dependencies:** Strictly forbidden from accessing the physical File System, Network, or actual Database (`bbolt`). All external boundaries must be passed in as interfaces.

## 2.2 Integration Tests
Integration tests verify that our code works correctly with external adapters and libraries.
- **Scope:** Used specifically for testing the `SystemStore` against a real, temporary `bbolt` database file on disk, or testing the FileSystem wrapper against the real OS.
- **Isolation:** Each integration test must generate a unique temporary file/database and clean it up using `t.Cleanup()`.

## 2.3 End-to-End (E2E) Tests
E2E tests verify complete user workflows (e.g., starting the CLI, loading an OpenAPI spec, and making an HTTP request to the simulated endpoint).
- **Execution:** These tests start the actual HTTP server on a random open port, send real HTTP requests using an `http.Client`, and validate the HTTP responses.

## 2.4 Layer Isolation via Build Tags
Because Integration and E2E tests involve heavy I/O operations (Database, Networking), they slow down developer feedback loops.
- **Rule:** Slow tests MUST be separated using Go build tags. Place `//go:build integration` or `//go:build e2e` at the very top of the test file. 
- A standard `go test ./...` will only run fast Unit Tests. CI pipelines will explicitly run `go test -tags=integration,e2e ./...` to execute the full suite.

---

# 🎭 3. Mocking & Faking Strategy

To keep tests readable and prevent fragile test suites, Mock:ctl strictly regulates how mocking is performed.

## 3.1 Manual Fakes over Mocks
As per EDL-029, the use of heavy reflection-based mocking frameworks (like `gomock` or `testify/mock`) is discouraged for simple boundaries.
- **Rule:** Engineers should prefer writing "Manual Fakes" (e.g., `type FakeStateStore struct { ... }`) that implement the required interfaces. Fakes contain real logic (like a simple `map`) rather than programmed expectations (`EXPECT().Save().Return(nil)`).
- **Benefit:** Fakes act like the real system, making tests less brittle to refactoring.

## 3.2 In-Memory File System
When unit testing the CLI or Parser components that require reading OpenAPI YAML files, tests MUST use the Go standard library's `testing/fstest` (MapFS) instead of writing physical files to `/tmp`.

## 3.3 Time & Timezone Mocking
Business logic (like JWT generation or log timestamps) MUST NEVER call `time.Now()` directly, as this causes flaky tests across different timezones (e.g., IST vs UTC in CI).
- **Rule:** Always inject a `Clock` interface into structs. In tests, inject a `FakeClock` that "freezes" time to a hardcoded `time.Time`, ensuring token and date validations remain perfectly deterministic.

---

# 📐 4. Test Design Patterns

## 4.1 Table-Driven Tests
All complex business logic (especially the `SpecificationEngine` and HTTP routing logic) MUST be tested using the Table-Driven Testing pattern.
- This allows a single test function to iterate over dozens of edge-case scenarios (e.g., valid inputs, missing fields, malformed data) without duplicating test setup code.

## 4.2 Golden Tests
When testing systems that generate large, complex outputs (like the entire JSON memory state export, or the generated OpenAPI tree), use Golden Tests.
- **How it works:** The test compares the generated output against a "golden" `.json` file saved in a `testdata/` directory. If the logic changes, the engineer updates the golden file via an environment flag (e.g., `UPDATE_GOLDEN=1 go test`).

## 4.3 Determinism
Tests must be deterministic. They must yield the exact same result every time they are run.
- When testing Fake Data Generation, the pseudo-random number generator (PRNG) MUST be seeded with a static, hardcoded value during the test to ensure the "random" data generated is perfectly predictable.

## 4.4 Test Data Management
Whenever a test requires reading a sample file (e.g., a dummy OpenAPI `api.yaml`), the file MUST be stored inside a `testdata/` directory adjacent to the test file.
- **Why:** The Go compiler explicitly ignores directories named `testdata`, ensuring that mock data files never accidentally bloat the final compiled binary.

## 4.5 Parallel Test Execution
To keep CI pipelines incredibly fast, pure unit tests must execute in parallel.
- **Rule:** Add `t.Parallel()` at the top of every isolated unit test and inside the execution func of Table-Driven Tests. Be extremely cautious to declare loop variables correctly to avoid the pre-Go-1.22 loop variable trap.

---

# 🚀 5. Performance & Security Testing

To guarantee that Mock:ctl remains highly performant and secure against malformed inputs, advanced Go testing tools must be utilized.

## 5.1 Concurrency Safety (Race Detection)
Because Mock:ctl handles concurrent HTTP traffic, silent data races are catastrophic.
- **Rule:** The Continuous Integration (CI) pipeline MUST execute the entire test suite with the race detector enabled (`go test -race ./...`). Any detected race condition will fail the build immediately.

## 5.2 Benchmark Testing
Hot paths—such as the routing logic that matches incoming requests to the OpenAPI specification tree, or the JSON data generator—must be highly optimized.
- **Rule:** Write Go benchmarks (`func BenchmarkXxx(b *testing.B)`) for all critical performance paths. Engineers must use these benchmarks to prove that refactoring does not introduce performance regressions.

## 5.3 Fuzz Testing
The `SpecificationEngine` parses external, arbitrary YAML/JSON files provided by the user. Malformed input could potentially panic the parser.
- **Rule:** Native Go Fuzzing (`func FuzzXxx(f *testing.F)`) MUST be implemented for the parsing layer. This will automatically bombard the parser with thousands of mutated byte arrays to discover hidden edge-case crashes before users do.

## 5.4 Goroutine Leak Prevention
Even if tests pass, background workers or HTTP handlers might remain hanging in memory (goroutine leaks), which would crash the production server over time.
- **Rule:** E2E and Integration test suites MUST use `go.uber.org/goleak`. Calling `goleak.VerifyNone(t)` at the end of the test explicitly verifies that the server shut down cleanly without leaving orphaned goroutines.

---

# 📊 6. Coverage Policy (EDL-030)

Coverage is treated as a confidence indicator, not a hard mathematical target that developers must game.

- **Critical Paths:** The core routing engine, OpenAPI parser, and Authentication (JWT) validation logic must maintain near 100% test coverage.
- **Diminishing Returns:** Engineers are explicitly instructed NOT to write low-value tests (like testing simple struct getters/setters) just to inflate the coverage percentage.
- **Exclusions:** Generated code (if any) and basic CLI command wrappers (which contain no business logic as per EDL-014) are excluded from coverage metrics.

---

# 🐛 7. Regression Testing

Every time a user reports a bug in the Mock:ctl backend:
1. The engineer MUST write a failing test (Unit or E2E) that explicitly reproduces the reported bug.
2. The engineer fixes the code to make the test pass.
3. The test remains in the suite forever as a Regression Test to ensure the bug never returns.

---

# 📌 Conclusion

The Testing Strategy outlined in this document ensures that Mock:ctl will remain stable and reliable as it scales. By leaning on the standard library, mandating table-driven and golden tests, enforcing parallel execution for speed, and utilizing advanced tools like Fuzzing, Benchmarks, and `goleak` to prevent memory corruption, the engineering team can move fast without breaking the core simulation engine. This strategy prioritizes meaningful developer confidence and system resilience over artificial coverage metrics.

---

# 🔗 Related Documents

**Foundation**

- PKS-000 — Repository Blueprint
- PKS-002 — Documentation Style Guide

**Engineering**

- PKS-020 — System Architecture
- PKS-024 — Component Architecture
- PKS-028 — Coding Standards
- Engineering-Decision-Log.md (EDL-028, EDL-029, EDL-030)

**Next Document**

- PKS-030 — Deployment Architecture

---

# 📜 Revision History

| Version | Date | Description |
|----------|------------|----------------------------------------------|
| 1.0 | 2026-08-18 | Initial Draft |

---

# ✅ Approval Checklist

- Executive summary completed
- Testing foundation and assertion libraries (`testing`, `testify`) established
- Boundaries for Unit, Integration, and E2E tests defined
- Fakes vs Mocks strategy defined (Manual Fakes preferred)
- Test design patterns (Table-driven, Golden, Deterministic) defined
- Test Data Management (`testdata/`) standardized
- Parallel Execution (`t.Parallel()`) mandated for speed
- Advanced testing (Race Detection, Benchmarks, Fuzzing) enforced
- Time mocking (`FakeClock`) and Goroutine leak prevention (`goleak`) mandated
- Code coverage philosophy and regression testing rules established
- Formatting follows PKS style guide
- Conclusion section included

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** ✅ Reviewed & Approved

**Architecture Status:** ✅ Established

**Next Document:** **PKS-030 — Deployment Architecture**
