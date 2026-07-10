## [2026-07-02 18:37] | Task: Compact app state output

### Execution Context
* **Agent ID**: `Amp`
* **Base Model**: `Amp deep mode, model not exposed`
* **Runtime**: `macOS SwiftPM + Linux/Windows Go tests`

### User Query
> Apply Codex-derived context-saving behavior to the open computer use fork, build it, and replace the local installed package with a release build.

### Changes Overview
**Scope:** `OpenComputerUseKit` macOS MCP runtime, Linux/Windows runtimes, and release docs

**Key Actions:**
- **[get_app_state options]**: Added `include_image`, `force_image`, `max_text_chars`, and `only_changes` to the macOS, Linux, and Windows `get_app_state` schema and dispatcher.
- **[Runtime output policy]**: Added per-app output cache so repeated screenshots can be omitted, app-state text can be capped after rendering, and unchanged state can return `There has been no change`.
- **[Regression coverage]**: Added unit tests for the output policy, integer argument parsing, and schema fields.
- **[Docs]**: Updated architecture, release notes, and plugin package metadata to record the host-facing behavior.

### Design Intent (Why)
Installed Codex artifacts expose state-context concepts such as `screenshotNeededForContext`, `There has been no change`, and screenshot / AX text presence flags. Moving the compacting controls into `open-computer-use` keeps the MCP result smaller before host wrappers see it, while preserving explicit knobs for clients that still need full images or uncapped text.

### Files Modified
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ComputerUseService.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ComputerUseToolDispatcher.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ToolDefinitions.swift`
- `packages/OpenComputerUseKit/Tests/OpenComputerUseKitTests/OpenComputerUseKitTests.swift`
- `apps/OpenComputerUseLinux/main.go`
- `apps/OpenComputerUseWindows/main.go`
- `docs/ARCHITECTURE.md`
- `docs/releases/feature-release-notes.md`
- `plugins/open-computer-use/.codex-plugin/plugin.json`
