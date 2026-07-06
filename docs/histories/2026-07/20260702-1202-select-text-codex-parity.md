## [2026-07-02 12:02] | Task: 对齐官方 `select_text` 工具

### Execution Context
* **Agent ID**: `Amp`
* **Base Model**: `Amp deep mode, model not exposed`
* **Runtime**: `macOS SwiftPM + Go tests`

### User Query
> 继续上一轮 Open Computer Use 改进，特别是从 Codex 学到的 `select_text`，并补完无效 image env 诊断、测试、文档和 Codex parity。

### Changes Overview
**Scope:** macOS Swift runtime, fixture/smoke, Windows runtime, Linux runtime, docs/skill

**Key Actions:**
- **[Official parity]**: 按官方 `computer-use` `1.0.857` 把 MCP tool surface 扩到 10 个 tools，新增 `select_text` schema、dispatcher 和跨平台实现。
- **[Native selection]**: macOS 使用 `kAXSelectedTextRangeAttribute`，Linux 使用 AT-SPI `Text` selection / caret API，Windows 使用显式 opt-in 的 UI Automation `TextPattern`，不走剪贴板或全局输入兜底。
- **[Fixture coverage]**: Fixture bridge 支持 `select_text` 并在 smoke suite 中验证 `prefix` / `suffix` 消歧后的选区结果。
- **[Diagnostics]**: macOS image capture env 的无效非空值现在会打印忽略诊断，空值和缺省仍静默使用默认值。
- **[Docs]**: 更新架构、质量水位、skill usage、release note，并新增 completed execution plan。

### Design Intent (Why)
官方 Codex `computer-use` 已从 9-tool surface 演进到 10-tool surface，`select_text` 是编辑类任务里比剪贴板、全局键盘或物理鼠标更安全的语义动作。把它落到三套 runtime 的原生 accessibility text API 上，可以保持本仓库“非侵入优先”的能力边界，同时减少 host 与官方工具面的提示词差异。

### Files Modified
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ToolDefinitions.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/MCPServer.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ComputerUseToolDispatcher.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ComputerUseService.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/FixtureBridge.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/AccessibilitySnapshot.swift`
- `packages/OpenComputerUseKit/Tests/OpenComputerUseKitTests/OpenComputerUseKitTests.swift`
- `apps/OpenComputerUseFixture/Sources/OpenComputerUseFixture/main.swift`
- `apps/OpenComputerUseSmokeSuite/Sources/OpenComputerUseSmokeSuite/main.swift`
- `apps/OpenComputerUseWindows/main.go`
- `apps/OpenComputerUseWindows/runtime.ps1`
- `apps/OpenComputerUseWindows/main_test.go`
- `apps/OpenComputerUseLinux/main.go`
- `apps/OpenComputerUseLinux/runtime.py`
- `apps/OpenComputerUseLinux/main_test.go`
- `docs/ARCHITECTURE.md`
- `docs/QUALITY_SCORE.md`
- `docs/releases/feature-release-notes.md`
- `skills/open-computer-use/SKILL.md`
- `skills/open-computer-use/references/usage.md`
- `docs/exec-plans/completed/20260702-select-text-codex-parity.md`
