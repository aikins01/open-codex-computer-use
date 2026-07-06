# Open Computer Use Usage

Read this reference when the task requires direct Computer Use tool calls, MCP configuration, or platform-specific behavior.

## MCP Server

For MCP clients that support stdio servers:

```toml
[mcp_servers.open_computer_use]
command = "open-computer-use"
args = ["mcp"]
```

Supported npm packages also expose `ocu` as a short alias, so `ocu mcp` is equivalent when available.

Equivalent JSON shape:

```json
{
  "mcpServers": {
    "open-computer-use": {
      "command": "open-computer-use",
      "args": ["mcp"]
    }
  }
}
```

The MCP server exposes:

```text
list_apps
get_app_state
click
perform_secondary_action
scroll
select_text
drag
type_text
press_key
set_value
```

## Direct CLI Tool Calls

Use `call` for one-off state checks:

```sh
open-computer-use call list_apps
ocu call list_apps
open-computer-use call get_app_state --args '{"app":"TextEdit"}'
```

Use `--calls` for short action sequences that need to reuse the same process state:

```sh
open-computer-use call --calls '[
  {"tool":"get_app_state","args":{"app":"TextEdit"}},
  {"tool":"set_value","args":{"app":"TextEdit","element_index":"1","value":"Draft"}},
  {"tool":"click","args":{"app":"TextEdit","element_index":"1"}},
  {"tool":"type_text","args":{"app":"TextEdit","text":"Hello"}}
]'
```

Use `--calls-file` when the sequence is too large for a readable shell command:

```sh
open-computer-use call --calls-file examples/textedit-overlay-seq.json --sleep 0.5
```

## Full Text Snapshots

Snapshot text is truncated to 500 characters by default and ends with `...` when truncation happens. This keeps normal UI state compact for agent planning and element-targeted actions.

Use full-text mode when the task depends on complete semantic text, such as chat histories, email bodies, document text, or long form content:

```sh
open-computer-use call get_app_state --args '{"app":"TextEdit","show_full_text":true}'
open-computer-use snapshot --show-full-text TextEdit
```

The same `show_full_text` tool argument and `--show-full-text` snapshot flag apply on macOS, Linux, and Windows.

For context-budgeted hosts, `get_app_state` also accepts output controls on macOS, Linux, and Windows:

```sh
open-computer-use call get_app_state --args '{"app":"TextEdit","include_image":false,"max_text_chars":20000}'
open-computer-use call get_app_state --args '{"app":"TextEdit","include_image":false,"only_changes":true,"max_text_chars":20000}'
open-computer-use call get_app_state --args '{"app":"TextEdit","include_image":true,"force_image":true}'
```

Use `include_image:false` to omit screenshot content, `force_image:true` to return a repeated screenshot, and `max_text_chars:0` for no post-render cap. Use `only_changes:true` for repeated checks: the first call records the app state, later calls return `There has been no change` when the accessibility tree is stable or a compact accessibility-tree diff when it changed. For long Amp, Claude, Codex, or other budget-sensitive threads, prefer `include_image:false`, `only_changes:true`, and `max_text_chars:20000` unless the image is required for the next action.

Action tools return `Action completed. Call \`get_app_state\` to fetch the updated UI state.` after success and refresh the runtime's internal snapshot cache. Run `get_app_state` again after an action when the next step depends on updated UI state; use `show_full_text:true` only when complete long text is needed.

## Choosing Targets

- Prefer app names or bundle identifiers returned by `list_apps`.
- Run `get_app_state` immediately before element-targeted actions.
- Re-run `get_app_state` after navigation, modal changes, page reloads, or failed actions.
- Use coordinate actions only when the rendered tree does not expose the target as an element.

## Platform Notes

### macOS

The macOS runtime uses Accessibility, ScreenCaptureKit, and targeted input events. It normally avoids moving the user's real pointer. The visual cursor overlay is part of the Open Computer Use experience and can be disabled by the surrounding runtime only when needed.

### Windows

The Windows runtime uses UI Automation and Win32 message fallbacks. It must run in a logged-in desktop session. A detached SSH or service context may start the CLI but fail to see top-level windows.

Windows `select_text` uses UIA `TextPattern` only when `OPEN_COMPUTER_USE_WINDOWS_ALLOW_UIA_TEXT_SELECTION=1` is set, because that selection operation can bring the target app to the foreground.

### Linux

The Linux runtime uses AT-SPI2 through the desktop session bus. It must run in a logged-in graphical session with usable accessibility services. Wayland screenshot and coordinate input support is compositor-dependent and best-effort.

## Safety

Pause and ask the user before actions that affect external systems or sensitive local state, including sending messages, submitting forms, deleting files, approving prompts, uploading files, or interacting with password managers.
