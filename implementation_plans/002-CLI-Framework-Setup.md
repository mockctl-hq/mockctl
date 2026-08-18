# Implementation Plan: 002-CLI-Framework-Setup

## 🎯 Objective
Initialize the Cobra CLI framework (`spf13/cobra`) to act as the primary entry point for the Mock:ctl internal testing harness. This will establish the `mockctl` root command and a basic `version` subcommand.

## 📌 Prerequisites
- Project successfully initialized (001-Project-Initialization completed).
- GitHub CI cache currently disabled (awaiting the first `go.sum` generation).

## 🛠️ Execution Steps

### Step 1: Install CLI Dependencies
Download and install the official `spf13/cobra` framework (EDL-013) and `spf13/viper` for configuration management.
- Command: `go get -u github.com/spf13/cobra@latest github.com/spf13/viper@latest`
- Command: `go mod tidy`
- *Edge-Case Addressed:* This will finally generate the `go.sum` file.

### Step 2: Implement the Root Command & Configuration
Create `root.go` inside the `cmd/mockctl/` directory. This file will define the base `mockctl` command and its `ExecuteContext()` function.
- File: `cmd/mockctl/root.go`
- *Viper Setup:* Initialize `viper` to read from an optional `.mockctl.yaml` config file (searching in `$HOME` and `.`). Use `viper.SetEnvPrefix("MOCKCTL")` to strictly isolate environment variables.
- *Logger Setup:* Initialize Go's built-in `log/slog` for structured JSON logging via a `PersistentPreRun` hook.
- *Rule:* Following EDL-014, this file MUST contain NO business logic, only CLI parsing and configuration.

### Step 3: Implement the Version Subcommand & Build Metadata
Create `version.go` to provide a way to verify the CLI is running correctly.
- File: `cmd/mockctl/version.go`
- *Metadata Injection:* Declare `Version`, `Commit`, and `Date` package-level variables that the Go compiler will populate via `-ldflags`.
- *Behavior:* Will print a formatted string including the version, git commit hash, and build timestamp when `mockctl version` is executed.

### Step 4: Implement Initial Unit Tests
To maintain high code coverage from day one, we will write basic unit tests for the CLI framework.
- Files: `cmd/mockctl/root_test.go` and `cmd/mockctl/version_test.go`
- *Behavior:* Tests will verify that commands initialize properly without panicking and that flags are correctly bound.

### Step 5: Update the Entry Point (Resiliency Features)
Modify the existing `main.go` to invoke the Cobra Root command securely.
- File: `cmd/mockctl/main.go`
- *Panic Handler:* Implement a global `defer recover()` block that catches fatal crashes and writes a `mockctl_crash_dump.json` file (as per PKS-030 Section 4.4).
- *Graceful Shutdown:* Use `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)` to pass a cancellable context to the CLI. This ensures the app gracefully handles `Ctrl+C` and system terminations.

### Step 6: Re-enable CI Caching
Now that `go.sum` is created, we will re-enable the GitHub Actions caching to significantly speed up future CI runs.
- File: `.github/workflows/ci.yml` (Change `cache: false` back to `cache: true`).

### Step 7: Verify CLI
Compile and test the new Cobra CLI commands.
- Command: `make test`
- Command: `make build`
- Command: `./bin/mockctl version`
- Command: `./bin/mockctl --help`

---

## ⚠️ Known Edge-Cases & Warnings
1. **GitHub Actions Cache Dependency:** Re-enabling the cache in Step 5 is safe *only because* we are committing `go.sum` in the same branch/commit. If `go.sum` fails to push, the CI will crash again.
2. **Cobra Package Deprecations:** We will strictly use the standard Cobra API (like `RunE` instead of `Run` for proper error bubbling) to prevent future technical debt.

## ✅ Expected Outcome
The `mockctl` binary will behave like a true enterprise CLI application with a structured `--help` menu, configuration binding (Viper), structured logging (slog), a global crash dumper, graceful shutdown signal handling, and a dedicated `version` command, fully powered by the Cobra framework.

**Status:** ✅ Approved
