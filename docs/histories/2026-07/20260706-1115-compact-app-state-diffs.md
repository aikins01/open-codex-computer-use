## [2026-07-06 11:15] | Task: Compact app-state diffs for agent polling

### Execution Context
* **Agent ID**: `Codex`
* **Base Model**: `GPT-5`
* **Runtime**: `macOS SwiftPM + Linux/Windows Go tests`

### User Query
> Improve Open Computer Use by learning efficiency behavior from the Codex app and updating the Amp-facing skill/plugin guidance.

### Changes Overview
**Scope:** macOS `OpenComputerUseKit`, Linux/Windows runtimes, skill docs, plugin metadata, and release/history docs

**Key Actions:**
- **[State diffs]**: Changed `only_changes` so repeated `get_app_state` calls return `There has been no change` when the accessibility tree is stable, or a compact accessibility-tree diff when it changed.
- **[Context control]**: Added default internal caps for change diffs so host threads do not receive an entire large tree after every small UI change.
- **[Action results]**: Changed successful action tools to return Codex-style compact acknowledgment text instead of a refreshed screenshot/tree result, while still refreshing the runtime's internal snapshot cache.
- **[Turn boundary]**: Wired `notifications/turn-ended` to clear snapshot and app-state output caches so stale element IDs and previous diff baselines do not leak into the next turn.
- **[Stale element refetch]**: Added macOS AX notification dirtying and conservative element refetch before element-index actions. The runtime reuses a refreshed element only when identifier or rendered tree line plus frame match narrowly; otherwise it asks the host to call `get_app_state` again.
- **[Cross-platform parity]**: Applied the same state-diff policy to macOS, Linux, and Windows runtimes.
- **[Amp guidance]**: Updated the skill and plugin prompts to recommend `include_image:false`, `only_changes:true`, and `max_text_chars` for repeated checks in long-running agent threads.

### Design Intent (Why)
Codex bundled Computer Use exposes official-style accessibility tree diff language and no-change responses. Matching that host-facing behavior makes Open Computer Use more useful in long Amp/Codex threads because the runtime can return small deltas before the host records tool output into model context.

Codex Computer Use client strings also show successful actions returning `Action completed. Call get_app_state to fetch the updated UI state.` That is a better default for long threads than embedding a full updated UI tree after every action; callers still get fresh state by explicitly calling `get_app_state`.

Codex service strings include `RefetchableSkyshotAXTree`, `RefetchableUIElement`, AX notification observers, and stale element messages. The open runtime now mirrors the safe part of that behavior: it observes app-level AX changes, marks cached snapshots dirty, and performs one conservative refetch for element actions instead of blindly trusting old AXUIElement references.

### Files Modified
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ComputerUseService.swift`
- `packages/OpenComputerUseKit/Tests/OpenComputerUseKitTests/OpenComputerUseKitTests.swift`
- `apps/OpenComputerUseLinux/main.go`
- `apps/OpenComputerUseLinux/main_test.go`
- `apps/OpenComputerUseWindows/main.go`
- `apps/OpenComputerUseWindows/main_test.go`
- `skills/open-computer-use/SKILL.md`
- `skills/open-computer-use/references/usage.md`
- `skills/open-computer-use/references/troubleshooting.md`
- `skills/open-computer-use/agents/openai.yaml`
- `plugins/open-computer-use/.codex-plugin/plugin.json`
- `docs/ARCHITECTURE.md`
- `docs/releases/feature-release-notes.md`
