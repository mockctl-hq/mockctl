# 🚀 PKS-030 — Deployment Architecture

> **Project:** Mock:ctl
>
> **Document ID:** PKS-030
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

The Deployment Architecture document defines how the Mock:ctl application transitions from source code to a distributable product. It establishes the Continuous Integration (CI) and Continuous Deployment (CD) pipelines, release automation, versioning rules, and distribution channels.

This document operationalizes the deployment decisions established in the Engineering Decision Log (EDL-034 through EDL-039), ensuring that the software lifecycle remains deterministic, automated, and error-free.

---

# 🎯 Purpose

The objectives of this document are to:
- Establish GitHub Actions as the single source of truth for CI/CD.
- Define the exact steps required for a Pull Request to be merged (CI Pipeline).
- Standardize the automated release process using GoReleaser.
- Enforce Semantic Versioning (SemVer) across all releases.
- Map out the phased Distribution Strategy (GitHub, Package Managers, Docker).

---

# 📌 Scope

This document applies to the Mock:ctl Go Backend.
It covers:
- Continuous Integration (CI) Pipeline
- Release Automation (CD)
- Versioning Rules
- Automated Artifact Generation (Binaries, Checksums)
- Phased Distribution Strategies

It **does not** cover the deployment of the future Flutter UI or Android APKs, which will be detailed in separate platform-specific deployment documents.

---

# ⚙️ 1. Continuous Integration (CI) Pipeline (EDL-034, EDL-035)

To maintain absolute code quality and prevent regressions, Mock:ctl utilizes **GitHub Actions** as its official CI platform.

## 1.1 Pull Request Quality Gates
Every Pull Request (PR) against the `main` branch MUST pass the following automated pipeline before it can be merged. Manual bypasses are strictly forbidden.

1. **Environment Setup:** Provision Linux, macOS, and Windows runners. Install Go.
2. **Dependency Check:** Run `go mod tidy` and verify no uncommitted changes exist.
3. **Format Verification:** Verify code complies with `gofmt` and `goimports`.
4. **Static Analysis & Linting:** Run `golangci-lint` (as defined in PKS-028) to catch anti-patterns.
5. **Fast Unit Tests:** Execute isolated unit tests.
6. **Race Detection:** Execute tests with `go test -race` enabled to catch concurrency bugs.
7. **Integration Tests:** Execute slow tests (Database/OS) triggered via `//go:build integration` tags.

## 1.2 Deterministic Execution
The CI pipeline MUST execute in a reproducible environment. Docker containers or explicitly versioned GitHub Action runners must be used to ensure the build environment does not silently change.

## 1.3 CI Pipeline Optimization (Caching)
To ensure fast developer feedback loops, the CI pipeline MUST implement aggressive caching. 
- Go modules (`go mod download`) and build caches MUST be cached using `actions/cache` to reduce CI execution time from minutes to seconds.

## 1.4 Secrets & Credential Management
Zero-trust security applies to the deployment pipeline.
- Hardcoding passwords, GPG keys, or Personal Access Tokens (PATs) in the repository is strictly banned.
- All deployment credentials must be injected securely at runtime exclusively via **GitHub Actions Secrets**.

## 1.5 Strict Branch Protection & Governance
The `main` branch is considered sacred. No human may push directly to `main`.
- **Mandatory Reviews:** Every PR must receive at least one approval from a code owner before merging.
- **Status Checks:** The GitHub Actions CI pipeline must pass completely. A failed CI run permanently blocks the "Merge" button.

## 1.6 Vulnerability Scanning (CVE)
To ensure the binary remains secure over time, automated vulnerability scanning is mandatory.
- **Dependabot:** GitHub Dependabot must be enabled to scan `go.mod` for known CVEs.
- **Trivy:** The CI pipeline MUST run an automated Trivy scan on the generated Docker images before publishing them to the registry.

## 1.7 Artifact Retention Policies
To prevent cloud storage bloat and reduce costs, the CI/CD pipeline MUST enforce strict retention lifecycle rules:
- **Nightly Builds & PR Artifacts:** Automatically deleted after 14 days.
- **Official Releases (Tags):** Retained indefinitely.

---

# 📦 2. Release Automation (EDL-036, EDL-037)

Mock:ctl explicitly bans manual release creation. Developers must never compile binaries on their local machines for public distribution.

## 2.1 GoReleaser Integration
**GoReleaser** is the official release automation tool.
- It is triggered automatically by GitHub Actions whenever a new Git Tag is pushed (e.g., `git tag v1.0.0 && git push origin v1.0.0`).
- It automatically cross-compiles the Go binary for Linux, macOS (Intel & Apple Silicon), and Windows.
- It generates `.tar.gz` and `.zip` archives.
- It automatically calculates and signs `sha256` checksums for security verification.

## 2.2 Release Notes & Conventional Commits
Release notes will be automatically generated by GoReleaser based on the Git commit history. 
- **Rule:** To prevent messy or useless release notes, the repository MUST strictly enforce the **Conventional Commits** standard (e.g., `feat:`, `fix:`, `docs:`). GoReleaser will use these prefixes to automatically categorize the changelog into "Features", "Bug Fixes", and "Chores".

## 2.3 Supply Chain Security (SLSA & Code Signing)
To ensure Mock:ctl is trusted by enterprise users and modern operating systems:
- **SBOM:** GoReleaser MUST automatically generate a Software Bill of Materials (SBOM) for every release to comply with SLSA standards.
- **Code Signing:** macOS binaries MUST be codesigned and notarized (to bypass Gatekeeper). Windows binaries MUST be signed with an Authenticode certificate.

## 2.4 Explicit Compilation Matrix
To guarantee native performance across all modern hardware architectures, GoReleaser MUST explicitly compile binaries against a defined matrix.
- **macOS:** `darwin/amd64` (Intel) and `darwin/arm64` (Apple Silicon M1/M2+ native).
- **Windows:** `windows/amd64` and `windows/arm64`.
- **Linux:** `linux/amd64` and `linux/arm64`.

---

# 🏷️ 3. Versioning Strategy (EDL-038)

Mock:ctl strictly adheres to **Semantic Versioning 2.0.0 (SemVer)**.

- **MAJOR (vX.0.0):** Breaking changes to the CLI flags, API endpoints, or configuration file formats.
- **MINOR (v1.X.0):** New features that are backward-compatible.
- **PATCH (v1.0.X):** Backward-compatible bug fixes and security patches.

*Pre-releases:* Alpha and Beta versions must use SemVer pre-release identifiers (e.g., `v1.0.0-beta.1`).

## 3.1 Hotfix & Rollback Strategy
If a critical production bug is discovered in a released version (e.g., `v1.2.0`):
- Engineers MUST NOT wait for the `main` branch to mature.
- A hotfix branch MUST be created from the `v1.2.0` tag. The fix is applied, and a new tag `v1.2.1` is pushed to instantly trigger the GoReleaser pipeline for a patch release.

## 3.2 Nightly & Canary Builds
To allow the community and QA teams to test upcoming features without waiting for an official release, the CI pipeline MUST automatically generate and publish a "Nightly Build" artifact on every push to the `main` branch. These builds are for testing only and are explicitly marked as unstable.

---

# 🌍 4. Distribution Strategy (EDL-039)

To manage complexity, Mock:ctl will distribute artifacts in three progressive phases.

## 4.1 Phase 1 (Internal Developer Artifacts)
- **CLI Binaries:** The Mock:ctl CLI binary is strictly an **internal developer tool** (EDL-053). GoReleaser will compile and upload these binaries to GitHub Releases as hidden/internal artifacts for QA and developer testing. They will NOT be advertised to end users.
- **Verification:** Internal engineers can verify the download against the GoReleaser-generated `checksums.txt`.

## 4.2 Phase 2 (Public GUI Applications)
- **Public Product:** The only product released to "Real Users" is the **Mock:ctl Flutter Application**.
- **Distribution:** The CI pipeline will bundle the Go Backend into the Flutter App and distribute it via:
  - **macOS:** Apple App Store & `.dmg` files.
  - **Windows:** Microsoft Store & `.exe` / MSIX installers.
  - **Android:** Google Play Store & `.apk` files.

## 4.3 Phase 3 (Containers & System Packages)
- **Docker:** Official Docker images (`docker pull mockctl:latest`) published to GitHub Container Registry (GHCR) and Docker Hub.
  - *Security Rule:* Docker images MUST be built using **multi-stage builds** starting from a `scratch` or `alpine` base to keep the image size under 20MB.
  - *Runtime Rule:* The container MUST be configured to run as a **non-root** user to minimize the security attack surface.
  - *Reliability Rule:* The Dockerfile MUST include a `HEALTHCHECK` instruction that pings the Mock:ctl `/health` endpoint to enable automatic container restarts on failure.
  - *Data Safety Rule:* Docker deployment instructions MUST mandate mounting the `bbolt` database to a **Persistent Volume** to prevent catastrophic data loss if the container is destroyed.
  - *Resource Quota Rule:* Container manifests MUST explicitly define CPU and Memory Requests/Limits (e.g., `memory: 256Mi`, `cpu: 100m`) to prevent Out of Memory (OOM) node crashes.
  - *Observability Rule:* When running inside a container, Mock:ctl MUST automatically switch its logging output to **Structured JSON** for seamless ingestion by Datadog, ELK, or Loki.
- **Linux:** APT (Debian/Ubuntu) and RPM packages for system-level installation.

## 4.4 Post-Deployment Crash Reporting
When Mock:ctl is deployed on an end-user's machine, runtime panics are difficult to debug remotely.
- **Crash Dumps:** The CLI MUST contain a global panic handler that intercepts fatal errors. Instead of printing a raw stack trace to the console, it MUST generate a localized `mockctl_crash_dump.json` file. The user can then voluntarily attach this file to a GitHub Issue for triage.

## 4.5 Zero-Downtime Deployments (Cloud API Mocks)
For enterprise teams deploying Mock:ctl permanently on cloud servers (via Docker Swarm or Kubernetes):
- The deployment architecture MUST support **Rolling Updates** (Blue/Green). A new container instance must report as healthy (via the `/health` endpoint) before the orchestrator shuts down the old version, ensuring zero downtime for running tests.
- **Graceful Shutdown & Draining:** When Kubernetes sends a `SIGTERM` signal, the Mock:ctl server MUST NOT die immediately. It must gracefully finish processing all in-flight HTTP requests and flush logs to disk before exiting.

## 4.6 Helm Charts & Orchestration
To prevent teams from writing raw, error-prone Kubernetes YAMLs:
- The repository MUST provide an official **Helm Chart** (`helm install mockctl`). This chart will automatically provision the Deployment, Service, Persistent Volume Claims (PVC), and ConfigMaps in a single, standardized command.

## 4.7 Edge Security & Rate Limiting
Mock:ctl must NEVER be directly exposed to the public internet without a shield.
- **Reverse Proxy:** Any public-facing cloud deployment MUST be placed behind an API Gateway or Reverse Proxy (such as NGINX or Traefik).
- **Protection:** The proxy MUST handle TLS termination (HTTPS), IP-based Rate Limiting, and basic DDoS protection to prevent the Mock:ctl engine from being overwhelmed by malicious traffic.

---

# 📌 Conclusion

The Deployment Architecture ensures that Mock:ctl scales from source code to a global user base effortlessly. By centralizing CI/CD in GitHub Actions and automating the entire release lifecycle with GoReleaser, the engineering team eliminates manual toil, guarantees deterministic builds, and secures the software supply chain.

---

# 🔗 Related Documents

**Foundation**

- PKS-000 — Repository Blueprint
- PKS-004 — Documentation Index

**Engineering**

- PKS-020 — System Architecture
- PKS-028 — Coding Standards
- PKS-029 — Testing Strategy
- Engineering-Decision-Log.md (EDL-034 to EDL-039)

---

# 📜 Revision History

| Version | Date | Description |
|----------|------------|----------------------------------------------|
| 1.0 | 2026-08-18 | Initial Draft |

---

# ✅ Approval Checklist

- CI/CD platform (GitHub Actions) established
- Pull Request quality gates defined
- Release automation (GoReleaser) mandated
- Semantic Versioning (SemVer) enforced
- Phased distribution strategy outlined
- Supply Chain Security (Code Signing, SBOM) mandated
- Secure Multi-Stage Docker builds enforced
- CI Optimization (Caching) and Secrets Management defined
- Hotfix and Rollback Strategy established
- Vulnerability Scanning, Branch Protection, and Nightly Builds enforced
- Explicit Compilation Matrix (Apple Silicon) and Crash Dumps standardized
- SRE Standards (Conventional Commits, Retention Policies, HEALTHCHECK, Persistent Volumes, Zero-Downtime) enforced
- Cloud-Native Standards (Helm Charts, Graceful Shutdown, JSON Logging, Resource Quotas, Edge Security) enforced
- Formatting follows PKS style guide
- Conclusion section included

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** ✅ Reviewed & Approved

**Architecture Status:** ✅ Established
