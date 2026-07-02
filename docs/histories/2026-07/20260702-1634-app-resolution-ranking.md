## [2026-07-02 16:34] | Task: Fix app-name resolution picking windowless helper processes

### 🤖 Execution Context
* **Agent ID**: Claude Code (Amp)
* **Base Model**: Claude
* **Runtime**: macOS, swift test via Xcode-beta toolchain

### 📥 User Query
> get_app_state with `"app":"Safari"` fails with `Apple event error -10005: cgWindowNotFound` while `"app":"com.apple.Safari"` works; fix the underlying cause instead of documenting workarounds.

### 🛠 Changes Overview
**Scope:** packages/OpenComputerUseKit

**Key Actions:**
- **[Root cause]**: `AppDiscovery.resolvedRunningApp` matched apps by display name OR executable name with `first(where:)` over an active-first sorted list. Bitwarden's Safari-extension helper (`exec=safari`, activationPolicy=accessory, reports `isActive=true`) matched the query "Safari" via the executable fallback and won over the real Safari, then failed window capture with `cgWindowNotFound`.
- **[Fix]**: Extracted a pure `bestResolutionIndex(of:matching:)` ranking that prefers regular-activation-policy apps and display-name matches: regular name > regular executable > accessory name > accessory executable.
- **[Tests]**: Added three `bestResolutionIndex` unit tests including the Bitwarden/Safari regression case.

### 🧠 Design Intent (Why)
Name queries should never resolve to background helper processes when a regular app with a matching display name is running; ranking by activation policy and match kind fixes the class of bug instead of the single symptom.

### 📁 Files Modified
- `packages/OpenComputerUseKit/Sources/OpenComputerUseKit/AppDiscovery.swift`
- `packages/OpenComputerUseKit/Tests/OpenComputerUseKitTests/OpenComputerUseKitTests.swift`
