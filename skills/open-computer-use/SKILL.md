---
name: open-computer-use
description: Platform-neutral guidance for using Open Computer Use, the open-source Computer Use MCP server and CLI for macOS, Linux, and Windows. Use when an agent needs to install, verify, troubleshoot, configure, or operate Open Computer Use through its native CLI, stdio MCP server, or direct Computer Use tool calls.
---

# Open Computer Use

## Overview

Open Computer Use exposes Computer Use as a local CLI and stdio MCP server. It is not Codex.app-specific; adapt the commands and MCP config to the agent runtime you are operating in.

It supports the same core tool surface across macOS, Linux, and Windows:
`list_apps`, `get_app_state`, `click`, `perform_secondary_action`, `scroll`,
`select_text`, `drag`, `type_text`, `press_key`, and `set_value`.

## Core Workflow

1. Check the CLI is installed with `open-computer-use -h` or `ocu -h`. If installation or setup is missing, read [references/installation.md](references/installation.md).
2. On macOS, run `open-computer-use doctor` before the first real GUI task. If permissions are missing, ask the user to approve Accessibility and Screen Recording in the onboarding UI.
3. Inspect available apps before acting: `open-computer-use call list_apps`.
4. Capture current UI state with `open-computer-use call get_app_state --args '{"app":"TextEdit","include_image":false,"only_changes":true,"max_text_chars":20000}'` when polling or operating inside a long agent thread. The first call returns capped state; later calls return either `There has been no change` or a compact accessibility-tree diff.
5. Use the default screenshot-bearing `get_app_state` only when visual inspection is needed. When the task needs complete long text, such as chat history, email bodies, document text, or long form content, call `get_app_state` with `show_full_text: true`.
6. Prefer element-targeted actions using `element_index` from the latest `get_app_state` result.
7. After successful action tools, call `get_app_state` again when the next step depends on updated UI state.
8. For multi-step CLI work, use `open-computer-use call --calls '<json-array>'` so one process can reuse the latest element index mapping.
9. For agent runtimes that support local MCP servers, configure `open-computer-use mcp` or `ocu mcp` and call the exposed Computer Use tools directly. Read [references/usage.md](references/usage.md).
10. If communication, permission, or desktop-session access fails, read [references/troubleshooting.md](references/troubleshooting.md).

## Operating Rules

- Treat the target desktop as the user's real session. Do not inspect password managers, unrelated private content, or sensitive apps unless the user explicitly asked for that task.
- Ask before sending, deleting, purchasing, approving, uploading, or making other externally visible changes.
- Do not assume Codex.app plugin helpers are available. Use the installed `open-computer-use` / `ocu` CLI or an explicit MCP config.
- Always run `get_app_state` before using `element_index`; do not guess indexes across sessions or after large UI changes.
- For repeated checks, use `include_image:false`, `only_changes:true`, and a reasonable `max_text_chars` cap unless the screenshot or full text is necessary.
- Prefer semantic actions, `select_text`, and `set_value` for editable controls. Use coordinate `click`, `scroll`, and `drag` only when the element tree does not expose a safer target.
- On macOS, do not enable `OPEN_COMPUTER_USE_ALLOW_GLOBAL_POINTER_FALLBACKS=1` unless the user explicitly wants diagnostic behavior that may move the real pointer.
- On Windows and Linux, confirm the command is running inside the logged-in desktop session before assuming GUI automation is available.

## Common CLI Actions

```sh
open-computer-use -h
ocu -h
open-computer-use doctor
open-computer-use call list_apps
ocu call list_apps
open-computer-use call get_app_state --args '{"app":"TextEdit"}'
open-computer-use call get_app_state --args '{"app":"TextEdit","include_image":false,"only_changes":true,"max_text_chars":20000}'
open-computer-use call get_app_state --args '{"app":"TextEdit","show_full_text":true}'
```

For a short sequence that reuses state in one process:

```sh
open-computer-use call --calls '[
  {"tool":"get_app_state","args":{"app":"TextEdit"}},
  {"tool":"click","args":{"app":"TextEdit","element_index":"1"}},
  {"tool":"type_text","args":{"app":"TextEdit","text":"Hello from Open Computer Use"}},
  {"tool":"press_key","args":{"app":"TextEdit","key":"Return"}}
]'
```

## MCP Usage

For runtimes that can launch local MCP servers over stdio, use:

```toml
[mcp_servers.open_computer_use]
command = "open-computer-use"
args = ["mcp"]
```

Read [references/usage.md](references/usage.md) for JSON config examples, direct tool-call patterns, and platform notes.

## References

- [references/installation.md](references/installation.md): one-time CLI install, agent MCP install commands, and macOS permissions.
- [references/usage.md](references/usage.md): MCP config, direct CLI calls, sequencing, and platform behavior.
- [references/troubleshooting.md](references/troubleshooting.md): permission, desktop-session, app discovery, and action failures.
