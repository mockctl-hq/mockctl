# 🌐 PKS-027 — API Design Guidelines

> **Project:** Mock:ctl
>
> **Document ID:** PKS-027
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

Mock:ctl is an API simulation engine. While it hosts the user's mocked endpoints, it also needs to expose its own **Admin API** to allow developers, test scripts, and the future Flutter UI to control the server dynamically (e.g., resetting state, tweaking chaos levels).

This document establishes the strict RESTful design guidelines for Mock:ctl's internal and administrative APIs. By standardizing namespaces, response envelopes, and error formats, we ensure that the Mock:ctl Admin API is predictable and never collides with the user's mocked endpoints.

---

# 🎯 Purpose

The objectives of this document are to:
- Establish a safe, collision-free namespace for Admin endpoints.
- Define a consistent JSON response envelope (success and error).
- Standardize HTTP status code usage within the Mock:ctl system.
- Map out the initial core Admin API endpoints, including WebSockets/SSE and Frictionless Monetization APIs.
- Define security (Localhost binding, Local Auth Tokens), CORS policies, and pagination standards.
- Ensure the Admin API is robust enough to serve as the backend for the future SaaS and Desktop UI.

---

# 📌 Scope

This document specifies the design guidelines for APIs *built by* Mock:ctl (Admin & Internal APIs). 
It **does not** enforce rules on the user's OpenAPI schemas (which Mock:ctl blindly simulates regardless of their design quality).

This document covers:
- URL Namespace isolation
- Response standardizations
- Query Parameters & Pagination Standards
- Admin API Security (Local Auth Token) & CORS Policy
- Core Admin Endpoints (State, Chaos, Events, System/Monetization)
- API Versioning & Deprecation Policy

---

# 🛡️ 1. Namespace Isolation

The most critical challenge of a mock server is ensuring that its Admin API does not accidentally overwrite or block a user's mock API.

**Rule:** All Mock:ctl administrative endpoints MUST be isolated under a reserved, double-underscore prefix.

- **Reserved Prefix:** `/__mockctl/`
- **Example Path:** `/__mockctl/chaos`

If a user attempts to load an OpenAPI spec that contains a route starting with `/__mockctl/`, the `RuntimeEngine` MUST reject it with a startup error, as this namespace is strictly protected.

---

# 📦 2. Standard Response Envelope

To make the Admin API easy to consume for automated test scripts and the Flutter UI, all responses must follow a strict, predictable JSON envelope.

**Rule (Content-Type):** The Admin API strictly communicates using `application/json`. Any incoming request containing a body must send `Content-Type: application/json`. If not, the server MUST reject it with a `415 Unsupported Media Type` error. 

The only exceptions are:
1. **Real-Time Events:** uses `text/event-stream`.
2. **File Uploads (Projects):** uses `multipart/form-data` for OpenAPI spec streaming.

## 2.1 Success Response

All successful requests (2xx) must return a JSON object containing a `data` key.

```json
{
  "success": true,
  "data": {
    "server": "online",
    "port": 8080,
    "chaos_level": 0
  }
}
```

- **`success` (boolean):** Always `true` for 2xx responses.
- **`data` (object/array):** The actual payload. It is never `null`. If there is no data, return an empty object `{}`.

## 2.2 Error Response

All failed requests (4xx, 5xx) must return a standardized error object mapped directly from the Go `DomainError` (defined in PKS-025).

```json
{
  "success": false,
  "error": {
    "code": "ERR_STATE_RESET_FAILED",
    "message": "Failed to reset memory state because a write lock is active.",
    "status": 500
  }
}
```

- **`success` (boolean):** Always `false` for 4xx/5xx responses.
- **`error.code` (string):** A programmatic, uppercase string for the UI to handle logic.
- **`error.message` (string):** A human-readable message.

---

# 🚥 3. HTTP Status Code Standards

The Admin API strictly adheres to standard REST semantics:

- **200 OK:** Request succeeded.
- **201 Created:** A resource (like a temporary override) was successfully created.
- **204 No Content:** Request succeeded, but no data needs to be returned.
- **400 Bad Request:** The client sent invalid parameters (e.g., setting Chaos Level to 150%).
- **401 Unauthorized:** Missing or invalid Local Auth Token.
- **403 Forbidden:** The action is not allowed (e.g., Localhost binding violation).
- **404 Not Found:** Admin route does not exist.
- **409 Conflict:** The action cannot be performed due to current state (e.g., trying to start a plugin that is already running).
- **415 Unsupported Media Type:** Client did not send `application/json`.
- **429 Too Many Requests:** The server is rejecting the request because the rate limit or concurrent connection limit has been exceeded.
- **500 Internal Server Error:** Mock:ctl encountered an unrecoverable Go panic or internal failure.

---

# 🔍 4. Query Parameters & Pagination Standards

To ensure consistency when the Desktop UI or CLI fetches large datasets (like telemetry logs), all list-based Admin APIs must support a standard pagination format:

- **`?page=1`**: The current page number (1-indexed).
- **`?limit=50`**: The number of items per page (Max: 100).
- **`?sort=createdAt_desc`**: Standardized sorting string (Field + `_asc` or `_desc`).

---

# 🔒 5. Admin API Security & CORS Policy

Because the Admin API can wipe the database and manage Premium Licenses, it must be protected from external network tampering and browser-based attacks.

## 5.1 Localhost Binding Restriction
By default, the `/__mockctl/` endpoints are strictly bound to `127.0.0.1`. Even if the developer binds the mock server to `0.0.0.0` to share it on their local office network, incoming requests from external IPs to the `/__mockctl/` namespace MUST be rejected with a `403 Forbidden`. Only the developer's local machine (or the Flutter UI) can access the Admin API.

## 5.2 Local Authorization Token
To prevent other malicious software running on the user's localhost from hijacking the Admin API (e.g., wiping the database or stealing SaaS tokens), the Mock:ctl binary generates a secure, random `admin.token` file in the `~/.mockctl/` directory on startup. 
All CLI and Flutter UI requests to `/__mockctl/` MUST include this token as a header: `Authorization: Bearer <token>`. Requests without a valid local token are rejected with `401 Unauthorized`.

## 5.3 CORS (Cross-Origin Resource Sharing)
To support future Web-based SaaS Dashboards that communicate with the local daemon, the Admin API strictly implements the following CORS headers for all `/__mockctl/` routes:
- `Access-Control-Allow-Origin: *` (or a specific SaaS Dashboard URL).
- `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS`.
- `Access-Control-Allow-Headers: Authorization, Content-Type`.

## 5.4 Internal Rate Limiting
To prevent the Desktop UI from accidentally crashing the Go backend with aggressive polling loops, all Admin APIs enforce a hard rate limit of **100 requests per second** per local client.

---

# 🗺️ 6. Core Admin Endpoints (v1)

These are the foundational endpoints that will be built into the `RuntimeEngine` alongside the user's mocked routes.

## 6.1 System Health (Docker / CI Readiness)
- `GET /__mockctl/health`
  - A completely unauthenticated, lightweight endpoint that returns a simple `{"status":"ok"}`. Used exclusively for Docker/Kubernetes health checks and CI/CD readiness probes.

## 6.2 State Management
- `POST /__mockctl/projects/{name}/state/reset`
  - Instantly wipes the project's memory state and re-initializes it.
- `GET /__mockctl/projects/{name}/state/export`
  - Returns the specific project's current memory state as a downloadable JSON object.

## 6.3 Chaos Engineering
- `PATCH /__mockctl/projects/{name}/chaos`
  - Updates the active chaos configuration for the specific project (e.g., payload: `{"error_rate": 20, "latency_ms": 500}`).

## 6.4 Real-Time Events (WebSockets/SSE)
- `GET /__mockctl/events`
  - Opens a Server-Sent Events (SSE) or WebSocket stream. The server pushes live metrics and incoming API request logs to the connected Flutter UI to populate real-time dashboards (similar to Charles Proxy or Postman).

## 6.5 Workspace & Projects
- `GET /__mockctl/status`
  - Returns server health, active daemon status, and total global request counts.
- `GET /__mockctl/projects`
  - Lists all active projects hosted by the daemon.
- `POST /__mockctl/projects`
  - Creates a new project (accepts `multipart/form-data` for OpenAPI file streaming) and mounts it to `/{projectName}`.
- `DELETE /__mockctl/projects/{name}`
  - Deletes a project and unmounts its routes.
- `POST /__mockctl/projects/{name}/endpoints`
  - Visually adds a single endpoint to an existing Hybrid Project without a full spec file.
- `POST /__mockctl/projects/{name}/overrides`
  - Injects a temporary static JSON response into a specific project route for immediate testing.

## 6.6 System & Monetization (SystemStore)
These endpoints interact directly with the embedded `bbolt` database to manage the application's configuration and premium features.
- `GET /__mockctl/system/settings`
  - Returns the developer's global configurations (e.g., UI theme, default port).
- `POST /__mockctl/system/auth`
  - Accepts a Frictionless Magic Link token (from the OAuth login flow) instead of a manual license key. Triggers a background validation with the Cloud Server before saving the returned JWT to the `auth_bucket`.
- `GET /__mockctl/system/telemetry`
  - Returns the paginated list of offline usage metrics waiting to be synced to the Cloud.

---

# 📅 7. API Versioning & Deprecation Policy

To ensure clean URLs and long-term stability for automated testing pipelines and the UI, Mock:ctl uses **Header-Based Versioning**.

- **Version Header:** Every request to the Admin API MUST include the header `Accept-Version: v1`.
- **Non-Breaking Changes:** Adding new endpoints, or adding new fields to a JSON response, will not bump the version.
- **Breaking Changes:** Removing endpoints or modifying existing JSON structures requires a new version (e.g., `Accept-Version: v2`).
- **Deprecation Window:** When `v2` is introduced, `v1` will remain active and fully supported for a strict **6-month sunset period**, allowing users to migrate their test scripts without changing the endpoint URLs.

---

# 📌 Conclusion

The API Design Guidelines establish a clean, predictable, and isolated structure for Mock:ctl's internal administrative communication.
By protecting the `/__mockctl/` namespace and strictly enforcing JSON envelopes, we guarantee that developers can dynamically control the mock server without risking collisions with their own application logic. This foundation is essential for building a robust automated testing ecosystem and the future Flutter-based Desktop/SaaS UI.

---

# 🔗 Related Documents

**Foundation**

- PKS-000 — Repository Blueprint
- PKS-002 — Documentation Style Guide

**Engineering**

- PKS-024 — Component Architecture
- PKS-025 — Software Design Document (Master SDD)
- PKS-026 — Database Design (Memory State Architecture)

**Next Document**

- PKS-028 — Coding Standards

---

# 📜 Revision History

| Version | Date | Description |
|----------|------------|----------------------------------------------|
| 1.0 | 2026-08-15 | Added Monetization APIs, WebSockets, Header Versioning, and Final Approval |

---

# ✅ Approval Checklist

- Executive summary completed
- Namespace isolation rule (`/__mockctl/`) established
- Success response envelope standardized
- Error response envelope standardized
- Content-Type constraints (`application/json`) established
- HTTP status code usage defined
- Query Parameters & Pagination standards established
- Admin API Security (Localhost Binding, Local Auth Token, Rate Limiting) defined
- CORS Policy defined
- Core Admin endpoints mapped (Health, Reset, Chaos, Events, Auth)
- API Versioning & Deprecation Policy defined
- Formatting follows PKS style guide
- Conclusion section included

---

**Document Status:** ✅ Approved

**Version:** 1.0

**Review Status:** ✅ Reviewed & Approved

**Architecture Status:** ✅ Established

**Next Document:** **PKS-028 — Coding Standards**
