# `select_text` Codex parity

## 目标

把官方 Codex `computer-use` `1.0.857` 新增的 `select_text` 工具补到 Open Computer Use 的 macOS、Windows、Linux runtime 中，并让 schema、dispatcher、fixture smoke、文档和历史记录同步进入 10-tool surface。

## 范围

- 包含：macOS Swift runtime、fixture bridge、smoke suite、Swift schema tests、Windows Go/PowerShell runtime、Linux Go/Python runtime、架构/skill/release note/history 文档。
- 不包含：移除本仓库已有的 `show_full_text` 扩展、重写旧版 9-tool 历史逆向文档、发布版本号或替换用户已安装的 app bundle。

## 背景

- 相关文档：`docs/ARCHITECTURE.md`、`docs/HISTORY_GUIDE.md`、`docs/QUALITY_SCORE.md`、`skills/open-computer-use/references/usage.md`。
- 相关代码路径：`packages/OpenComputerUseKit`、`apps/OpenComputerUseFixture`、`apps/OpenComputerUseSmokeSuite`、`apps/OpenComputerUseWindows`、`apps/OpenComputerUseLinux`。
- 已知约束：保持非侵入优先，不用剪贴板、全局物理指针或前台抢占兜底实现文本选区。

## 风险

- 风险：不同平台文本 API 的索引单位不同，可能导致 emoji / 组合字符附近选区偏移。
- 缓解方式：macOS 使用 `NSString` / `NSRange` 与 AX 的 UTF-16 range 语义对齐；Windows 使用 UIA `TextPatternRange` 字符端点；Linux 使用 AT-SPI `Text` offset。
- 风险：目标文本重复时错误选中错误位置。
- 缓解方式：要求唯一匹配，重复时提示补 `prefix` / `suffix` 消歧。

## 里程碑

1. 官方 surface 探测与 schema 收敛。
2. macOS / fixture / smoke 实现。
3. Windows / Linux runtime parity。
4. 验证、文档、history 收尾。

## 验证方式

- 命令：`swift build --target OpenComputerUseKit`
- 命令：`swift build --product OpenComputerUseSmokeSuite`
- 命令：`swift build --product OpenComputerUse`
- 命令：`swift build --product OpenComputerUseFixture`
- 命令：`OPEN_COMPUTER_USE_VISUAL_CURSOR=0 .build/out/Products/Debug/OpenComputerUseSmokeSuite`
- 命令：`.build/out/Products/Debug/OpenComputerUseSmokeSuite --cursor-idle-only`
- 命令：`go test ./...` in `apps/OpenComputerUseWindows`
- 命令：`go test ./...` in `apps/OpenComputerUseLinux`
- 命令：`python3 -m py_compile apps/OpenComputerUseLinux/runtime.py`

## 进度记录

- [x] 确认官方 `computer-use` `1.0.857` 暴露 10 个 tools，并记录 `select_text` schema。
- [x] macOS tool definition、dispatcher、service 和 AX selected range 实现完成。
- [x] Fixture bridge 和 smoke suite 增加 `select_text` 覆盖。
- [x] Windows UI Automation `TextPattern` 和 Linux AT-SPI text selection parity 完成。
- [x] 文档、quality score、skill usage、history 同步完成。

## 决策记录

- 2026-07-02：保留本仓库已有 `show_full_text` 扩展；官方 `1.0.857` 未继续暴露该字段，但它已是本仓库面向长文本读取的既有能力。
- 2026-07-02：`select_text` 不使用剪贴板、全局按键或物理指针兜底；不支持原生文本 selection API 的元素直接返回能力边界错误。
- 2026-07-02：旧 reverse-engineering 样本和早期 9-tool execution plan 保持历史语境，不批量重写；当前架构、skill、release note 和新 plan/history 记录 10-tool 状态。
