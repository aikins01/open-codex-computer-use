## [2026-07-07 23:26] | Task: 修复 Safari tab action 与前台快捷键路径

### 🤖 Execution Context
* **Agent ID**: `Amp`
* **Base Model**: `API agent`
* **Runtime**: `macOS / SwiftPM`

### 📥 User Query
> 调查 Codex bundled Computer Use 是否有可借鉴实现，并把 Safari 临时 tab 无法通过 `perform_secondary_action("close tab")` 或 `press_key("super+w")` 关闭的问题修到本仓库 fork 中。

### 🛠 Changes Overview
**Scope:** `packages/OpenComputerUseKit`, `docs`

**Key Actions:**
- **[Safari custom action]**: 将 `Name:close tab ...` 这类 AppKit custom action descriptor 渲染为短名称，并让 executor 用同一套归一化逻辑把短名称映射回原始 AX action。
- **[Action matching correctness]**: 修正 `rawActions` 和 `prettyActions` 数组长度不一致时的错配风险，避免 rendered action 被错误映射到前面的 filtered raw action（例如 `AXPress`）。
- **[Foreground keyboard opt-in]**: 新增 `OPEN_COMPUTER_USE_ALLOW_GLOBAL_KEYBOARD_INPUT=1`，让 `press_key` 在显式 opt-in 时先激活目标 app，再通过 `.cghidEventTap` 发送快捷键，覆盖 Safari `Command-W` 这类 PID-targeted keyboard event 不可靠的场景。
- **[Regression coverage]**: 添加 secondary-action 名称归一化、raw/display action 对齐和新环境变量的单测覆盖。

### 🧠 Design Intent (Why)
Codex bundled Computer Use 暴露出更完整的 focus/event-tap 设计线索；本仓库保持非侵入默认行为不变，只把 Safari custom AX action 修成语义 action 路径，并把前台 HID keyboard 作为显式 opt-in 能力提供给需要真实菜单快捷键路由的场景。

### 📁 Files Modified
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/AccessibilitySnapshot.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ComputerUseService.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/InputSimulation.swift`
- `packages/OpenComputerUseKit/Tests/OpenComputerUseKitTests/OpenComputerUseKitTests.swift`
- `docs/ARCHITECTURE.md`
- `docs/releases/feature-release-notes.md`
