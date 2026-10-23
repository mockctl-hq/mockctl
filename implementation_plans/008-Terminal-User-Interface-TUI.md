# IMP-008: Terminal User Interface (TUI) Dashboard

> **Plan ID:** IMP-008
> **Version:** 2.0 (Compliance Refactor)
> **Epic:** Presentation Layer & User Experience
> **Status:** Approved
> **Complexity:** Extremely High (Elm Architecture, VS Code Layout, Collapsible Panels, Security, Mobile UX)
> **Target Branch:** `feature/imp-008-tui-dashboard`

## 🎯 Objective

Build a highly responsive, 3-pane Terminal User Interface (TUI) for Mock:ctl using the **Charmbracelet ecosystem** (`bubbletea`, `lipgloss`, `bubbles`) as officially adopted in **EDL-057**. **Crucially, as per EDL-053, this TUI is NOT a public end-user product.** It serves exclusively as an **Internal Developer Harness** designed to rigorously battle-test the Go Backend, Telemetry streams, and SystemStore locally before porting those exact features to the public Flutter Desktop/Mobile apps.

---
Final UX gaps resolved.


## 🏗 Architectural Blueprint (The Elm Architecture)

The TUI strictly follows Bubble Tea's Model-Update-View paradigm. Business logic is forbidden (EDL-014); it acts purely as a Presentation Layer Client.

### 🚨 Critical Architecture Rules (PKS-006 & PKS-020 Compliance)

1. **Embedded Engine Architecture (Strict Layering):** To perfectly mirror the future Flutter FFI architecture, the TUI MUST directly instantiate the Application Core (`app.MockctlApp`), which internally orchestrates the `ProjectManager`, `EventBroker`, and other domain services. The TUI (Presentation layer) MUST NOT bypass the Core Orchestrator to instantiate domain services itself. It embeds the Backend Engine into memory using the public APIs defined in PKS-025 and MUST NOT check for `admin.token`.
2. **Zero-Allocation Telemetry:** The TUI MUST NOT use an SSE client. Instead, it must directly call `broker.SubscribeRaw()` to receive real-time structs via zero-allocation Go channels, matching the IMP-007 architecture. Components MUST call `broker.UnsubscribeRaw()` when unmounted (e.g., closing overlays) to prevent goroutine/memory leaks.
3. **Daemon Subprocess Lifecycle:** `mockctl studio` acts as the master process holding the BBolt DB lock. It MUST boot the HTTP server via `app.StartServer(ctx)` in a background goroutine so external devices can connect. When the user exits the TUI (`q` or `Ctrl+C`), it MUST trigger the associated `context.CancelFunc()`. Crucially, the TUI MUST NOT exit immediately; it MUST wait on a `sync.WaitGroup` or shutdown channel until the Application Core signals that the `temp-state.json` auto-save and DB lock releases are complete, preventing data loss.
4. **Abstracted Telemetry Access:** To support the `[History]` overlay, the Application Core MUST expose an `app.GetTelemetryHistory(ctx, limit)` method. The TUI MUST call this abstraction, which internally delegates the read request to the `SystemStore`, strictly preserving the layered architecture and avoiding direct database access from the Presentation layer.
5. **Headless Debugging & Graceful Exit:** TUI MUST use the unified `slog` logger (configured to write to a `daemon.log` file, not stdout) to trace crashes, ensuring compliance with EDL-018. The daemon relies on BBolt's inherent file-locking robustness; it MUST ONLY listen for interceptable signals like `SIGINT` and `SIGTERM`, avoiding mathematically impossible `SIGKILL` traps.

### 🧱 Component Modular Architecture & Directory Tree

To prevent a monolithic "God Object", components are strictly separated. Common utilities are placed in a `shared/` directory to enforce DRY principles.

**TUI Directory Blueprint:**

```text
internal/tui/
├── app.go                // Root Bubble Tea Model (Struct, Init, Main Loop)
├── app_update.go         // Root Update() logic (Keyboard & Message routing)
├── app_view.go           // Root View() logic (Main 3-pane layout rendering)
├── commands.go           // tea.Cmd definitions (app.StartServer, app.ReloadEngine)
├── context.go            // TUIContext (Active Port, Global Configs)
├── keys.go               // Keybindings definitions (Instantiated in main.go, injected into TUI; NO Global Vars)
├── messages.go           // Strongly-typed tea.Msg event structs
├── models.go             // State structs shared across UI components (Passed by reference; NO Global Vars)
├── theme.go              // Adaptive Color definitions & NO_COLOR support
├── shared/               // Reusable cross-component utilities
│   ├── clipboard.go      // clipboardUtil (OS Clipboard handling & normalization)
│   ├── search.go         // SearchModel (Fuzzy Search/Filter component)
│   ├── toast.go          // ToastModel (Transient Notification component)
│   └── confirm.go        // ConfirmModel (Reusable Yes/No/Force modal component)
└── components/
    ├── header/
    │   ├── header.go     // HeaderModel (40% Left, 60% Right Layout Container)
    │   ├── logo/
    │   │   └── logo.go   // "Mock:ctl" text & ':' hover tooltip (Daemon Status)
    │   └── actions/
    │       └── icons.go  // Right side overlay trigger icons & click routing
    ├── sidebars/
    │   ├── left/
    │   │   ├── sidebar.go         // LeftSidebarModel (Main container wrapping Icons Tab & Inner)
    │   │   ├── icons/             // Far-Left narrow strip
    │   │   │   ├── icons.go       // Defines available icons & click routing
    │   │   │   └── template.go    // Empty Sidebar Template (Interactive blank state for future icons)
    │   │   └── inner/             // The expandable side panel
    │   │       ├── header.go      // HeaderModel (Left: Title, Right: '+' Icon)
    │   │       ├── scaffold.go    // ScaffoldInputModel (Inline input triggered by '+')
    │   │       └── body/          // Inner panel content
    │   │           ├── explorer.go        // ProjectTreeModel (Struct & Init)
    │   │           ├── explorer_update.go // ProjectTree navigation
    │   │           └── explorer_view.go   // ProjectTree drawing logic
    │   └── right/
    │       ├── sidebar.go         // RightSidebarModel (Main container wrapping Icons Tab & Inner)
    │       ├── icons/             // Far-Right narrow strip
    │       │   ├── icons.go       // Defines available icons (Chaos, Snippets, Overrides)
    │       │   └── template.go    // Empty Sidebar Template (Interactive blank state for future icons)
    │       └── inner/             // The expandable side panel
    │           ├── header.go      // HeaderModel (Contains ONLY Title text)
    │           └── body/          // Inner panel content routing
    │               ├── chaos.go   // ChaosControlModel (Delay, Error Rate, Status Code Override)
    │               ├── snippets.go // CodeSnippetsModel (cURL generator)
    │               └── static.go  // OverridesModel (Hardcoded JSON static responses)
    ├── workspace/
    │   ├── workspace.go  // WorkspaceModel (Middle container & Active Tab manager)
    │   ├── tabs.go       // TabBarModel (Dirty state & orphaned tab tracking)
    │   ├── editor.go     // YamlEditorModel (Struct & Init)
    │   ├── editor_update.go // YamlEditor typing, paging & async highlighting
    │   ├── editor_view.go   // YamlEditor line numbers & viewport rendering
    │   ├── editor_format.go // YamlEditor auto-formatting & Windows EOL cleanup
    │   └── tester/              // API Tester Sub-Component (Multi-Protocol)
    │       ├── tester.go        // ApiTesterModel (Bottom panel container & Protocol Tabs manager)
    │       └── protocols/       // Protocol-specific testing UIs
    │           └── http/        // HTTP/REST API Tester (MVP)
    │               ├── http.go      // HttpTesterModel (Container for HTTP specifically)
    │               ├── client.go    // HTTP Networking, Context Timeouts & Binary protection
    │               ├── input.go     // Method dropdown & Path input
    │               ├── headers.go   // HTTP Headers textarea
    │               ├── body.go      // JSON Body textarea & auto-inference
    │               └── viewer.go    // Response Status Bar & Chroma JSON renderer
    ├── footer/           // Global sticky footer
    │   └── footer.go     // HelpFooterModel (Context-sensitive keybindings)
    └── overlays/
        ├── header_tools/ // Full-screen tools triggered via Header icons
        │   ├── inspector/    // Interactive CRUD JSON tree overlay
        │   │   ├── inspector.go // Struct & Init
        │   │   ├── update.go    // Tree navigation & CRUD input logic
        │   │   └── view.go      // Colored JSON tree rendering
        │   ├── settings/     // Settings overlay with nested tabs
        │   │   ├── settings.go  // Main settings container
        │   │   ├── theme.go     // Theme (Dark/Light) & UI animations
        │   │   ├── network.go   // CORS, TLS, Validation Mode & Timeouts
        │   │   ├── editor.go    // Tab size, Auto-format, Line wrapping
        │   │   └── tester.go    // Follow redirects & Global headers
        │   ├── history/         // Telemetry History Overlay (Multi-Protocol)
        │   │   ├── history.go   // Main history container & Protocol Tabs manager
        │   │   └── protocols/   // Protocol-specific history viewers
        │   │       └── http/    // HTTP/REST traffic history (MVP)
        │   │           ├── http.go   // Struct & Init (BBolt connection)
        │   │           ├── update.go // Scrolling, List selection & Search logic
        │   │           └── view.go   // Event List & JSON Details UI rendering
        │   ├── logs/            // Global System Log Overlay
        │   │   ├── logs.go      // Struct & Init (Ring buffer store)
        │   │   ├── update.go    // Auto-tail toggle & search typing logic
        │   │   └── view.go      // Colored log lines & search bar rendering
        │   └── profiler/
        │       └── profiler.go  // GoProfilerModel (runtime.MemStats chart)
        └── system_prompts/ // Shared blocking modals & alerts
            ├── delete.go    // DeleteProjectModel (Catastrophic deletion typed-prompt)
            ├── crashed.go   // DaemonCrashedModel (Port already in use error screen)
            └── warning.go   // ScreenWarningModel (Terminal size obstruction modal)
```

> **Note on "Method-Split Pattern":** To strictly prevent "God Objects" in complex BubbleTea components, we utilize Go's ability to split struct methods across multiple files within the same package. The Root App, YamlEditor, ProjectTree, and StateInspector MUST split their `Update()` and `View()` functions into dedicated `_update.go` and `_view.go` files to maintain a strict < 300 LOC limit per file.

1. **The Component Interface (Extensibility)**
    ```go
    type Component interface {
        Init(ctx TUIContext) tea.Cmd
        Update(msg tea.Msg) (Component, tea.Cmd)
        View() string
        SetFocus(focus bool)
    }
    ```
2. **Message Passing (Event Bus)**
   Components communicate via strongly-typed `tea.Msg` events (`ProjectSelectedMsg`, `EndpointSelectedMsg`, `TesterToggleMsg`).

### 1. The Global Model (`tui.Model`)

The Root Model delegates events to its children and manages global context:

- `Size`: Stores `tea.WindowSizeMsg` for responsive layout calculations. **Auto-Collapse:** The `Update()` loop MUST dynamically recalculate the API Tester's height on every resize. If the terminal height drops below a safe threshold while expanded (e.g. landscape rotation), it MUST auto-collapse to prevent `lipgloss` rendering panics.
- `FocusRing`: An array of `Component` references tracking strict keyboard focus. **Visual Indicator:** The currently focused panel MUST render with a Bright Blue (or Primary theme color) `lipgloss` border, while inactive panels render with Dim Gray borders, preventing "blind scrolling" navigation errors. **Fallback Routing:** If the currently focused component is unmounted or hidden (e.g., API Tester is collapsed via `Ctrl+\`), the `Update()` loop MUST explicitly intercept this and fallback the focus to the Middle Workspace (YAML Editor) to prevent a dead UI state.
- `ActiveEndpointID`: Tracks the currently selected endpoint (routes data to the Right Sidebar & API Tester).
- `ApiTesterExpanded`: Boolean controlling whether the bottom API Tester panel is currently visible.
- `ActiveOverlay`: An Enum tracking **Full Screen Overlays** (`None`, `StateInspector`, `History`, `GlobalLog`, `Settings`, `Profiler`, `DaemonCrashedModal`). The `DaemonCrashedModal` MUST display the raw `net.Listen` error (e.g., "port already in use") if the background HTTP server fails to bind on boot.
- `Drafts`: A `map[string]RequestState` caching Body/Headers/Path modifications per `EndpointID`. **State Persistence:** It MUST also cache `LastResponsePayload`, `LastStatusCode`, and `LastLatency` so API Tester responses survive Tab switching. _(See Security section for strict memory-only constraints)._

### 🪟 Detailed Layout Grid (The "VS Code" Architecture)

The layout is dynamically computed in the `View()` function, utilizing horizontal and vertical joins:

**1. Header (Top Bar - Height: 3, Width: 100%)**

- Left: Logo `Mock` (White/Adaptive) + `:` (Blinking Orange Indicator) + `ctl` (Orange).
- Status: Displays Daemon Status (e.g., `[ Daemon: Port 8080 ]`).
- Right: Five icons `[State Inspector]`, `[History]`, `[Global Log]`, `[Profiler]`, `[Settings]`. Activating any of these opens a **Full Screen Overlay**, completely covering the grid below until dismissed.
- **Global Toast Notification:** A transient 3-second flash message (e.g., `✅ Project Saved`) rendered dynamically at the top right. This MUST be implemented as a **Shared Component** (`shared/toast.go`) that listens for a global `ShowToastMsg`, allowing any child component in the TUI to easily trigger a visual feedback notification without tightly coupling to the Header.

**2. Left Panel (Navigation)**

- **Far-Left Icons Tab:** Narrow vertical strip. Switches the content of the Inner Sidebar (e.g., HTTP Projects vs other tools).
- **Inner Left Sidebar:**
    - Header: Title (Projects), `+` (Context-Aware Create).
    - **Context-Aware Scaffolding (The `+` Logic):**
        - **No Project Selected:** Clicking `+` renders the Inline Input below the header for a *Project Name*. Upon `Enter`, it calls `ProjectManager.Initialize(ctx, path)` to create the project and inject the default OpenAPI template, strictly delegating file creation to the core.
        - **Project Selected:** Clicking `+` renders the Inline Input below the active project for an *Endpoint Path* (e.g., `/users`). Upon `Enter`, the backend automatically appends the boilerplate YAML for this endpoint directly into the active editor.
    - Search: A reusable **Shared Fuzzy Search** component.
    - Body: A **Custom Tree Component** behaving like a file explorer. **External Sync:** The TUI MUST listen to `EventBroker` for `SystemStateChanged` events (e.g., project deleted via CLI) and automatically re-render the Tree to prevent "Ghost Projects" from causing crashes.
        - **Nodes:** Projects (Collapsible `▼ / ▶`).
        - **Leaves:** Endpoints (e.g., `GET /users`). **Visual Indicators:** If an endpoint has active Chaos constraints, its node MUST render a Yellow Dot `🟡`. If it has an active Static Override, it MUST render a Purple Dot `🟣`, ensuring users are globally aware of degraded or hardcoded routes.

**3. Middle Workspace (The Core Editor & Empty State)**

- **Empty State (Default):** If no project is loaded, the workspace displays a centered hero message: "No Project Selected. Select a project from the sidebar or create a new one", accompanied by a highly visible `[ + Create New Project ]` interactive button.
- **Top Tab Bar:** Renders open YAML files (e.g., `[ openapi.yaml ✖ ]`). **Dirty State Indicator:** If the YAML has unsaved changes, an asterisk MUST be rendered (e.g., `[ openapi.yaml* ✖ ]`). For Default/Scaffolded YAMLs, a `[ 📥 Import ]` icon MUST be rendered on the far right, allowing users to override the default template with their own local YAML file. **Orphaned State:** If an open project is deleted via the backend or sidebar, the tab MUST NOT close. Instead, it enters an `Orphaned` state, rendering a yellow dot (`[ openapi.yaml 🟡 ]`). Saving (`Ctrl+S`) an orphaned tab will recreate the project. **Re-Creation Conflict:** If the user creates a new project with the same name before saving the orphaned tab, the TUI MUST reject the save with a Toast Error to prevent data overwrite.
- **Center (YAML Viewer):** Renders the full YAML file content with Line Numbers. Uses `viewport` + `chroma` for syntax highlighting. MUST explicitly disable text-wrapping and implement **Horizontal Scrolling** (`Shift+Left/Right`) to prevent line-number misalignment on narrow screens. **Truncation Indicator:** The Viewer MUST render a visual indicator (e.g., `▶▶` or a fade effect) on the right edge of any line that exceeds the viewport width to alert the user of hidden text. **Async Highlighting:** For files >50KB, `chroma` parsing MUST run asynchronously in a `tea.Cmd` to prevent the main UI thread from freezing. **Cursor Status Bar:** A dedicated line below the editor MUST render the real-time cursor coordinates (e.g., `Ln 14, Col 5`) by querying `textarea.LineInfo()`, enabling users to blindly locate syntax errors without a mouse. **Graceful Empty State:** If the backend parses a 0-byte file, it MUST NOT throw an error. The Viewer MUST open in a normal, blank Read Mode. A floating hint (e.g., *"Press `Ctrl+I` to inject base template"*) MUST be rendered, allowing the user to seamlessly switch to Edit Mode and type from scratch without being blocked. **Paging Bounds Protection:** If a file is reloaded (e.g., via external sync) and its new length is shorter than the current cursor's Y-offset, the `Update()` loop MUST clamp the cursor and viewport Y-offset to `Max(0, len(lines)-1)` to prevent immediate `viewport` panics.
- **Expandable Error Diagnostics:** If the backend core (`internal/spec`) returns complex multi-line validation errors, the Error Bar MUST be expandable (or open a modal) to prevent critical YAML parsing errors from being truncated to a single unreadable line. The TUI MUST rely entirely on the Application Core to parse and safely reject malicious YAML payloads, acting strictly as an Anti-Corruption Layer client.
- **Dirty State Warning:** If a user attempts to exit or switch projects while the YAML editor has unsaved changes, the workspace MUST intercept the command and render an "Unsaved Changes! Are you sure?" confirmation modal to prevent data loss. **Ghost Tab Protection:** If a user attempts to close (`Ctrl+W`) a newly created, unsaved tab that exists only in RAM, the TUI MUST strictly prompt "Discard Unsaved Draft?" to prevent instant data loss. **External Modification Sync:** The TUI MUST check the underlying file's `ModTime` by calling `app.GetProjectModTime(ctx, projectName)` when the editor gains focus, strictly delegating disk I/O to the Core. If the file was modified externally, it MUST prompt "File changed on disk. Reload?". **Sync Race Condition Fix:** If the file is modified externally AND the TUI has unsaved local changes, a strict Conflict Resolution modal MUST be rendered: *"File modified externally AND you have unsaved changes! [Force Overwrite] or [Discard Local]?"*
- **Bottom (Collapsible API Tester):** Resembles the VS Code Terminal panel. Can be expanded/collapsed. Does *not* cover the YAML code completely. Contains:
  - **HTTP Method Toggle** & **Path Input** (`bubbles/textinput`). **Local History Buffer:** The Path Input and Body Textarea MUST support a local history stack (accessed via `Up/Down` arrows when focused), allowing users to quickly recall previously executed test payloads and paths without re-typing them.
  - **Tabs:** `[Body] | [Headers]` for dedicated input. The `[Headers]` tab MUST be a single raw `textarea` where users type standard HTTP header strings (e.g., `Key: Value`). **Auto-Inference & Sanitization:** If the user types a JSON payload in the Body tab, the Tester MUST automatically inject `Content-Type: application/json`. The underlying client MUST sanitize the user's manual headers and dynamically override any user-provided `Content-Length` header. **CRITICAL:** The TUI must NEVER force sanitization of `\t` or `\r\n` inside the API Tester Body, ensuring the exact raw bytes the user provides are sent for accurate testing. **Double-Encoding Trap:** If the body is `application/json`, the client MUST pass the raw `[]byte` from the textarea directly to `http.NewRequest`; it MUST NOT call `json.Marshal()` on it, which would dangerously double-encode the JSON string.
  - **Response Status Bar:** A dedicated line above the response viewer rendering the HTTP Status Code (e.g., `200 OK`) and Latency (e.g., `12ms`). **Enforced Timeout & TLS:** The underlying `http.Client` MUST explicitly use the "Default HTTP Timeout" from the Settings config. It MUST also configure a custom `http.Transport` with `ForceAttemptHTTP2: false` and `InsecureSkipVerify: true` to prevent TLS handshake errors. **Chunked Response Protection:** To prevent the TUI from hanging indefinitely on stalled `Transfer-Encoding: chunked` responses, the Body read operation MUST be explicitly wrapped in a Context with a strict `ReadTimeout` or `io.LimitReader`.
  - **Binary Response Protection:** The Response Viewer MUST inspect the `Content-Type` and body size. If the response is binary (`image/*`, `application/pdf`) or `> 1MB`, it MUST bypass `chroma` syntax highlighting and render a safe placeholder (e.g., `[ Binary Data - 5MB ]`) to prevent CPU lockups and TUI freezing.

**4. Right Panel (Context & Integration)**

- **Inner Right Sidebar:** Content switches based on the Far-Right Icons Tab.
    - State 1: **Chaos Controls** (Network Delay, Error Rate, and Status Code Override). **Empty State:** If no endpoint is selected, render "Select an endpoint to configure chaos constraints." Otherwise, provide a `[ Toggle Chaos: ON/OFF ]` master switch to explicitly mount/unmount the middleware. The Error Rate text input MUST enforce strict numeric-only validation. Crucially, provide a **Status Code Override** dropdown/input (e.g., `Force 400 Bad Request`) allowing the developer to force the Mock Server to bypass the default 200 OK response and test alternative spec responses.
    - State 2: **Code Snippets** (cURL, fetch snippets). **Empty State:** If no endpoint is selected, render "Select an endpoint to generate code snippets." Otherwise, provide a **Language Switcher** toggle/dropdown. **Path Interpolation:** The generator MUST automatically inject dummy values (e.g., `/users/1`) into the snippet to ensure it is instantly runnable without syntax errors.
    - State 3: **Static Overrides** (Configure hardcoded JSON response overrides). **Empty State:** If no endpoint is selected, render "Select an endpoint to create static overrides."
- **Far-Right Icons Tab:** Narrow vertical strip mirroring the left side. Contains three icons (Chaos, Code, Override) used to switch the Inner Sidebar.

**5. Footer**

- **Help Footer (Fixed Height):** A dynamic `bubbles/help` component rendering **Context-Sensitive** keybindings. It MUST have a fixed `Max-Height` (e.g., 2 lines). If context-sensitive bindings change length on focus swap, a dynamic height will cause the entire 3-pane grid to "bounce" (layout thrashing), which is strictly forbidden.

---

## 📱 Termux & Mobile Environment Constraints (NEW)

Designing for Termux requires addressing OS-level idiosyncrasies not present in desktop terminal emulators:

1. **Android Soft-Keyboard Obstruction & Debounce:** When a user focuses a `textarea`/`textinput`, the Android keyboard consumes ~50% of the screen height. Since Termux `SIGWINCH` events can be unreliable when the soft-keyboard appears, the TUI MUST utilize a strict flex-box layout (`lipgloss`) where text inputs explicitly snap/anchor to the top 50% of the terminal space to prevent "blind typing". **Resize Debouncer:** Keyboard animations emit dozens of rapid `SIGWINCH` events; the `Update()` loop MUST implement a 50ms debouncer before recalculating `lipgloss.JoinHorizontal/Vertical` to prevent CPU layout-thrashing and lockups.
2. **Kinetic Touch-Drag Scrolling:** Swiping on a mobile screen does not reliably emit mouse-wheel events. The TUI viewports and lists MUST explicitly map Touch-Drag events to vertical scroll offsets, enabling native-feeling kinetic scrolling.
3. **Zombie Connections (Doze Mode) & Battery Drain:** When Termux is backgrounded, Android OS suspends the process. The TUI MUST explicitly listen for BubbleTea `tea.BlurMsg` and `tea.FocusMsg` events. On `BlurMsg`, it MUST pause all visual spinners and UI ticks to conserve battery. Upon `FocusMsg` (resume), it MUST instantly refresh its memory channel subscriptions to resume normal real-time operation.
4. **Emoji & Ligature Misalignment:** Emojis like `⚠️` have variable East Asian Widths. Furthermore, developer fonts with Ligatures (e.g., `!==`, `=>`) merge characters visually but not logically. The TUI MUST NOT attempt to manually calculate cursor offsets for ligature characters. It MUST rely entirely on `textarea.LineInfo()` and strict `runewidth` normalizers to ensure `lipgloss` border math remains pixel-perfect and the visual cursor does not detach from the text.
5. **Windows EOL (`\r\n`) Pasting Corruption:** Pasting Windows-formatted text into Termux (Linux) introduces invisible carriage returns (`^M`), which corrupts Bubble Tea rendering. The `shared/clipboard.go` utility MUST explicitly run a regex/replace to normalize all `\r\n` to `\n` ONLY for the YAML Code Editor.
6. **Multi-Line Paste Hijacking & YAML Tab Trap:** To prevent multi-line pasted text from executing terminal commands unexpectedly, the Root Model `Update()` loop MUST explicitly prioritize capturing `tea.PasteMsg`. Furthermore, because YAML 1.2 forbids literal tabs, the `PasteMsg` handler MUST run a regex sanitization loop to convert all `\t` characters into 2 spaces *before* passing the text to the `textarea`, applying strictly to the YAML Code Editor, NOT the API Tester.
7. **X11 Clipboard Dependency:** Standard Go clipboard libraries fail in Termux due to missing X11/Wayland. The `shared/clipboard.go` utility MUST explicitly implement a fallback to `termux-clipboard-set` and `termux-clipboard-get` via `exec.Command` to prevent copy actions from silently failing on mobile.

---

## 🛡️ Security & Threat Mitigation (CRITICAL)

As a local developer tool, the TUI must be hardened against privilege escalation, SSRF, terminal injection, and DoS attacks.

1. **Terminal Injection Attack (ANSI Spoofing):** The TUI MUST strictly sanitize all incoming telemetry strings, stripping raw ANSI control characters before rendering them into the `viewport` to prevent terminal corruption.
2. **Memory Exhaustion (Paste Bomb):** The YAML Editor (`textarea`) and the Shared Clipboard utility MUST enforce a strict **5MB size limit** on pasted text to prevent instant Out-Of-Memory (OOM) crashes.
3. **SSRF (Server-Side Request Forgery) Prevention:** The API Tester's base URL MUST be strictly locked to the Mock Server's daemon host/port. The TUI must reject attempts to test arbitrary external/internal IP addresses.
4. **Sensitive Data Leak (Drafts Protection):** The `Drafts` map (which may contain API keys) MUST remain exclusively in memory (RAM). It must never be flushed to a local disk file (e.g., `studio.state.json`).
5. **Telemetry Flood (DDoS by EventBroker):** The TUI's `broker.SubscribeRaw()` handler MUST implement a **Rate Limiter** (e.g., max 100 events/sec) that drops excess events, protecting the TUI's CPU/RAM during heavy daemon load.
6. **Environment Hijacking (`LD_PRELOAD` / `PATH` spoofing):** When spawning the background daemon process, the TUI MUST use an absolute path (`os.Executable()`) and sanitize environment variables to prevent shell injection.
9. **Path Traversal via Spec Import (Dir Hijacking):** The UI MUST strictly sanitize user inputs for Project Names and endpoint paths, rejecting any Directory Traversal patterns (e.g., `../`).
10. **OS Clipboard Overflow Attack:** The Shared Clipboard utility MUST impose a **1MB truncation limit** on text copied from the API Tester to prevent crashing desktop environments.
11. **File Descriptor (Socket) Exhaustion:** The TUI MUST utilize a global `http.Transport` connection pool. To support long Chaos Delays without risking FD leaks, the TUI must explicitly call `Body.Close()` and trigger a Context Cancellation (`context.Cancel()`) when the user aborts an active request via `Esc`.
12. **Data Remanence (RAM Safety):** As Go strings are immutable, the TUI MUST NOT attempt to manually zero out (`\0`) memory using `unsafe` CGO hacks. It must rely on secure JWT storage within the `SystemStore` and avoid keeping plain-text secrets in long-lived memory structures outside of strictly necessary scopes.
13. **YAML Anchor Bomb (Billion Laughs):** The Application Core (`internal/spec`) strictly disables YAML alias/anchor expansion (`&a` / `*a`). The TUI relies on this backend defense to prevent memory exhaustion.
14. **ReDoS (Regex Denial of Service) in Chroma:** Syntax highlighting utilizes complex regular expressions. To prevent CPU lockups (ReDoS), the TUI MUST enforce a strict **100KB size limit** on responses sent to `chroma`. Any JSON/YAML exceeding this limit will be rendered as plain text.
15. **OSC-52 Clipboard Hijacking:** The terminal sanitizer MUST explicitly strip **OSC (Operating System Command)** escape sequences (like `\033]52;`) from mock responses to prevent attackers from silently injecting malicious scripts into the user's OS clipboard.

---

## ⚙️ Backend Subsystem Prerequisites (TUI Blockers)

The TUI requires the following backend systems to be implemented or exposed before or during TUI development:

1. **Dynamic Route Reloading (Live Reload):** When `Ctrl+S` is pressed to save YAML edits, saving to disk is insufficient. The TUI MUST NOT orchestrate backend pipelines manually; it MUST call the unified `app.ReloadEngine(ctx, projectName)` method on the Application Core to dynamically rebuild the Spec, MockGenerator, and hot-swap the HTTP router.
2. **Chaos Middleware & Overrides API:** The Right Sidebar features "Chaos Controls" and "Static Overrides", but the Core Engine currently lacks these middlewares. A `ChaosMiddleware` (handling Network Delay & Error Rates) and an `OverrideStore` MUST be implemented in the backend so the TUI has actual state to toggle.
3. **Reactive Data-Binding (Event Bus):** The "Code Snippets" (cURL snippets) in the Right Sidebar MUST be fully reactive. When a user modifies headers/body in the Bottom API Tester, the event bus must instantly emit state updates to re-render the cURL snippet to match the user's current draft.
4. **Tree Synchronization (Event Bus):** When a user modifies an endpoint path in the YAML editor, the Left Panel (Tree Explorer) must automatically rename the node. The TUI's internal event bus MUST emit a `ProjectStructureChangedMsg` post-save, passing the updated, read-only `SpecModel` DTO generated by the Application Core. The Sidebar MUST strictly use this DTO to re-render the tree, and MUST NOT attempt to parse the YAML schema itself.
5. **Daemon Port Binding Injection:** The API Tester (Bottom Panel) issues real HTTP requests. If the daemon was started with a custom port (e.g., `mockctl studio --port 9090`), the TUI's `tui.Model` Context MUST be injected with this `ActivePort` at startup so the Tester resolves the correct `localhost` URL.

---

## ⌨️ Keybinding Map & Mouse Controls

### 🚫 The Keyboard Trap Prevention

- **Global Input Trap:** When any input component is focused, single-letter global shortcuts MUST be disabled. Press `Esc` to unfocus.
- **The Tab Key Trap:** In the YAML Editor's Edit Mode, the `Tab` key MUST insert 2 spaces for indentation and **NOT** trigger a FocusRing change.
- **Focus-Scoped Keybindings (Modal Trapping):** Components MUST strictly intercept keybindings based on the `FocusRing`. When a Full Screen Overlay is active, all keyboard events MUST be trapped within its specific focus ring. Uncaught events MUST be explicitly discarded to prevent "focus bleed" from interacting with the background grid.

### Global Shortcuts

- **Mouse Support:** Clicks focus panels, switch Tabs, or expand the API Tester.
- `Tab` / `Shift+Tab`: Cycle focus sequentially across panels.
- `/`: Trigger Shared Fuzzy Search (Scope depends on Active Focus).
- `f`: Toggle "Auto-Tail" follow mode inside the Global Log viewport.
- `Ctrl+F`: Auto-Format shortcut. Formats the current JSON/YAML buffer (using `json.Indent` or `yaml` marshaler) to resolve unformatted minified pasting.
- `Shift+Left Arrow`: Trigger "Collapse All" across the entire Project Tree to quickly escape deeply nested folder mazes.
- `Ctrl+B`: Toggle (Hide/Show) Left Panel.
- `Ctrl+E`: Toggle (Hide/Show) Right Panel.
- **`Ctrl+\`:** Toggle (Expand/Collapse) the Bottom API Tester in the Middle Workspace.
- `Ctrl+S`: Save YAML edits.
- `Ctrl+W`: Close the active YAML Tab. **Cleanup Hook:** Closing a tab MUST trigger a `context.CancelFunc()` to instantly kill any background HTTP requests (API Tests) associated with that tab, preventing nil-pointer panics when the request returns to an unmounted component.
- `Ctrl+R` or `Ctrl+Enter`: Execute API Test request. **Pre-Flight Validation:** If testing with a JSON body, the TUI MUST pre-validate the JSON format before dispatching the request. Invalid JSON must show a Toast Error and abort, preventing garbage data transmission.
- `Ctrl+Backspace`: **Quick State Reset.** Instantly wipes the `MemoryStateStore` for the active project and regenerates fresh seed data from the OpenAPI spec, allowing developers to reset their tests without restarting the daemon.
- `Esc`: Unfocus inputs, return YAML to Read Mode, OR **Cancel an active API Test HTTP request**. **Mobile Debouncing:** In Termux, closing the soft-keyboard emits `Esc`. The TUI MUST debounce or safely differentiate this event to prevent accidental cancellation of active HTTP requests when the user simply hides the keyboard.
- `c`: Trigger the **Shared Clipboard Utility** (max 1MB).
- `d`: Trigger Project Deletion. **Catastrophic Protection:** To prevent accidental data loss, the deletion modal MUST require the user to explicitly type the project's exact name to confirm deletion.
- `R`: Attempt daemon restart (by calling `app.StartServer(ctx)`) if `DaemonCrashedModal` is active due to a freed port.
- `Up/Down Arrows` or `j/k`: Navigate lists and viewports.
- `?`: Toggle the Help Menu footer expansion.
- `Ctrl+C`: Context-Aware. If inside a YAML/Text Editor, it copies text to OS clipboard. Otherwise, it gracefully exits the TUI.
- `Ctrl+Q` or `q`: Gracefully exit TUI / close Full Screen Overlays. **Context-Aware Safety:** The `q` key MUST be strictly intercepted and ignored if the YAML Viewer (Read Mode) or any text input is focused, preventing accidental application termination when users attempt to quit a pager.

---

## 🧪 Testing Strategy

In compliance with **PKS-029 (Testing Strategy)**, the TUI MUST achieve >85% coverage via the following approach:
1. **Elm Architecture Unit Tests:** Pure `Update(msg)` functions MUST be tested by passing a mock `Model` and a specific `tea.Msg`, then asserting the returned new `Model` state and the returned `tea.Cmd`.
2. **End-to-End (E2E) UI Tests:** Use `charmbracelet/x/exp/teatest` to simulate user keypresses and validate the `View()` string output against pre-recorded **Golden Files**. This adheres to the E2E taxonomy rather than Integration testing.
3. **Test Locations:** `_test.go` files MUST reside adjacent to their components (e.g., `components/sidebars/right/sidebar_right_test.go`). Shared logic (e.g., `clipboard.go`) MUST have isolated unit tests.

---

## 🚀 Implementation Phases

### Phase 1: Framework Setup, Commands & Shared Utils

1. Install dependencies (`bubbletea`, `lipgloss`, `chroma`, `clipboard`, `sahilm/fuzzy`).
2. Create `mockctl studio` CLI command. **CRITICAL:** Reconfigure `slog` to write to `daemon.log` ONLY and never `os.Stdout`, to prevent terminal UI corruption.
3. Initialize BubbleTea with `tea.WithAltScreen()`, `tea.WithMouseCellMotion()`, and `tea.WithBracketedPaste()` to enable touch interactions.
4. Implement `internal/tui/shared/` utilities. (Includes EOL Normalizer).
5. **Security Check:** Implement the Absolute Path Daemon spawning and SSRF constraints.
6. **Testing Requirement:** Ensure any unit/integration tests written for the TUI components that spin up background services implement `goleak.VerifyNone(t)` to strictly align with PKS-029 goroutine leak prevention rules.

### Phase 2: VS Code Grid, Termux Shields & Themes

1. Define adaptive global themes. Read `NO_COLOR=1` to enforce Monochrome rendering. **Terminal Profile Fallback:** Query `lipgloss.ColorProfile()`; if TrueColor is unsupported, automatically fallback to a High-Contrast ANSI 16-color palette to prevent invisible text. **Light Mode Protection:** The TUI MUST query `lipgloss.HasDarkBackground()` on boot and swap to a dark-text palette if a light terminal background is detected. Ensure `runewidth` emoji padding.
2. Build the VS Code Mirror layout (`lipgloss.JoinHorizontal`/`Vertical`).
3. Build the Custom Collapsible Tree (Left) and Icons Tabs (Left/Right switching). Implement Kinetic Touch-Drag scrolling.
4. **Screen Size Obstruction (Blocking Overlay):** The TUI requires a minimum dimension to render the 3-pane layout properly. If `Height < 24` OR `Width < 120`, the TUI MUST completely collapse and render a centered, full-screen red warning modal: *"Terminal is too small. Please increase window size, zoom out, or rotate your device."* **Death Spiral Prevention:** If the terminal shrinks below an absolute hard-limit of `10x10`, the TUI MUST bypass `lipgloss` rendering entirely and output a bare `!` to prevent the word-wrapper from entering an infinite loop panic.

### Phase 3: Focus Management & Telemetry

1. Implement the `FocusRing`.
2. **The Blinking Colon:** 500ms `tea.Tick`.
3. **Zero-Allocation Telemetry:** Implement `SubscribeRaw()` channels. Ensure components explicitly call `UnsubscribeRaw()` to prevent memory leaks when navigating away. Implement ANSI/OSC sanitization for raw payload strings before rendering to prevent Terminal Injection.
4. **History Subsystem:** Expand `SystemStore` with `GetTelemetryHistory()` to power the Global Log and History overlays.

### Phase 4: Tabbed Workspace & API Tester Integrations

1. **Middle Workspace:** Implement the top Tab Bar, YAML viewer, and `Ctrl+\` expandable API Tester panel. Ensure `Drafts` persistence strictly in RAM. **Ghost Request Prevention:** Active HTTP requests triggered by the Tester MUST be executed with a `context.Context`. Pressing `Esc` MUST call the `context.CancelFunc()` to instantly kill the TCP socket and prevent background goroutine leaks.
2. **YAML Security:** Ensure YAML parsing disables anchor expansion.
3. **Header Overlays:** Implement the 5 Full Screen Overlays (`State Inspector`, `History`, `Global Log`, `Profiler`, `Settings`). The `Settings` overlay MUST be divided into 4 tabs: **1. Theme** (Dark/Light/Monochrome toggles & Termux Battery Saver to disable animations). **2. Network** (Global CORS toggle, Strict OpenAPI Validation mode, HTTP Timeout, and History Buffer limits). **3. Editor** (Tab Size spacing, Line Wrapping toggle, and Auto-Format on Save). **4. API Tester** (Follow HTTP Redirects toggle, and Default Headers injection). **Dynamic Re-Evaluation:** When Theme or Editor settings are toggled, the TUI MUST instantly trigger a `tea.ClearScreen` and re-evaluate `lipgloss.Style` or viewport bounds; it MUST NOT require an app restart. **Settings Persistence:** All preferences MUST be persisted explicitly into the `settings_bucket` of the BBolt `SystemStore`. The TUI MUST NOT fragment configurations into separate `studio.yaml` files. The `State Inspector` MUST display a fully interactive, syntax-highlighted JSON tree of the active project's `MemoryStateStore`. It MUST support **CRUD Operations**, allowing the developer to manually Insert, Edit, or Delete records in RAM to test how the mocked endpoints react. To prevent `lipgloss` rendering panics from extreme horizontal widths, the JSON serializer for the State Inspector MUST enforce a `MaxDepth: 5` limit, truncating deeply nested objects (e.g., `"{...}"`).
4. **Go Memory Profiler:** Implement the `Profiler` overlay. Because the Go backend will eventually be embedded inside the Flutter app via FFI (where leaks cause fatal crashes), this overlay MUST read Go's `runtime.MemStats` in real-time. It should display Heap Allocations, Goroutine counts, and GC pauses on a visual bar chart (via `bubbles/progress`) to rigorously battle-test the core engine's memory safety during heavy mocking.
4. **Global Log & History:** Global Log enforces a 1,000 line Ring Buffer with Auto-Tail. **Overlay Searchability:** Both Overlays MUST implement an inline Search/Filter component (`/` keybinding) to allow users to grep through the 1,000-line viewport without tedious manual scrolling. **Nested Esc Routing:** Pressing `Esc` while the search input is focused MUST only unfocus the input. A second `Esc` is required to close the overlay, preventing frustrating accidental modal closures.
5. **Dynamic Host Resolution:** Fetch daemon port from `/settings` for accurate Code Snippets in the Right Sidebar.

### 🧪 Phase 5: TUI Testing Strategy

Because the TUI is a complex, stateful Elm Architecture (BubbleTea) application, testing MUST adhere strictly to PKS-029 rules:

1. **Assertion Libraries & Parallelism:** All TUI tests MUST strictly use `testify/assert` and `testify/require` for evaluation. Furthermore, all isolated unit tests MUST call `t.Parallel()` at the top to ensure lightning-fast CI execution, preventing pipeline bottlenecks.
2. **View Rendering Tests (Golden Testing):** The `View()` function returns complex ANSI strings (colors, layouts, borders). Unit tests MUST use Golden Files (`.golden`) in a `testdata/` directory to capture and assert the exact ANSI string output of key components to prevent UI layout regressions.
3. **Model Update Tests (Event Simulation):** Unit tests MUST directly call the `Update(msg tea.Msg)` function by simulating keystrokes (e.g., `tea.KeyMsg{Type: tea.KeyEnter}`) and asserting that the resulting `Model` state and returned `tea.Cmd` match expectations.
4. **Isolated Component Testing:** Complex components (Sidebar, API Tester, Overlays) MUST be tested in isolation by instantiating their specific models, rather than always booting the full 3-pane root application.
5. **Headless Integration Tests (Build Tags):** Integration tests that boot the Mock Server background daemon MUST run the TUI in a headless mode, execute a simulated HTTP request via the API Tester, and assert the response. These heavy tests MUST be isolated using the `//go:build integration` build tag at the top of the file, and they MUST strictly call `goleak.VerifyNone(t)` to guarantee zero background goroutine leaks.
6. **Fallback & NO_COLOR Tests:** The TUI MUST have explicit tests ensuring that when `NO_COLOR=1` is set, or when `lipgloss.HasDarkBackground()` returns false, the UI correctly falls back to monochrome or dark-text mode without panicking. Tests MUST also verify the "Death Spiral Prevention" by simulating a terminal dimension of `< 10x10`.
7. **Time Determinism (FakeClock):** Components that display time-sensitive data (e.g., the Global Log or History overlays) MUST NOT use `time.Now()`. They MUST accept the `internal/shared/clock.go` interface, allowing tests to inject a `FakeClock` to perfectly freeze time and prevent flaky tests across CI timezones.
