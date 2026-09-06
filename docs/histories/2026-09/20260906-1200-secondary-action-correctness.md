## [2026-09-06 12:00] | Task: 修正 secondary action 映射

### 📥 User Query
> 独立提取 macOS secondary-action correctness 变更，支持 Safari custom action description，并修复过滤后的 raw/display action 错配。

### 🛠 Changes Overview
**Scope:** OpenComputerUseKit macOS accessibility snapshot 与 action execution

**Key Actions:**
- **动作渲染**: 解析 AppKit `Name:...` action descriptor，并输出有意义的短名称。
- **名称冲突**: 短名称冲突时显示原始 descriptor，保证每个输出 selector 都可唯一映射回对应动作。
- **动作执行**: 保持原始 AX action 精确匹配的兼容性；短名称基于同一组 role-filtered raw actions 匹配，拒绝歧义短名称。
- **回归测试**: 覆盖 Safari close-tab、过滤对齐、无效 descriptor 与歧义名称。

### 🧠 Design Intent (Why)
*渲染与执行必须共享过滤和命名语义，否则压缩后的 action 数组会与原始 action 索引错位并执行错误动作。*

### 📁 Files Modified
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/AccessibilitySnapshot.swift`
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/ComputerUseService.swift`
- `packages/OpenComputerUseKit/Tests/OpenComputerUseKitTests/OpenComputerUseKitTests.swift`
- `docs/ARCHITECTURE.md`
- `docs/releases/feature-release-notes.md`
