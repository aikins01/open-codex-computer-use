# Open Computer Use Troubleshooting

Read this reference when setup, permission checks, app discovery, snapshots, or actions fail.

## First Checks

Start with:

```sh
open-computer-use -h
open-computer-use doctor
open-computer-use call list_apps
```

On macOS, `doctor` reports Accessibility and Screen Recording status. If either is missing, ask the user to approve the onboarding UI.

## App Not Found

If `get_app_state` cannot find an app:

1. Run `open-computer-use call list_apps`.
2. Use the app name or bundle identifier from that result.
3. Confirm the app is running and has a visible, non-minimized window.
4. On macOS, rerun `open-computer-use doctor`.

Do not silently switch to a different app when the requested target is not available.

## Empty Or Missing Snapshot

Common causes:

- The app has no visible window.
- The window is minimized, hidden, or on an unavailable desktop.
- macOS Screen Recording permission is missing.
- Windows or Linux commands are running outside the logged-in desktop session.
- Linux screenshot support is blocked by the compositor or desktop portal state.

Ask the user to bring the target app/window into a visible state when automation cannot do so safely.

## Truncated Text

Snapshot text is limited to 500 characters by default. If a visible chat message, email body, document paragraph, or form value ends with `...`, do not assume the page itself is missing content.

Request full accessibility text explicitly:

```sh
open-computer-use call get_app_state --args '{"app":"TextEdit","show_full_text":true}'
open-computer-use snapshot --show-full-text TextEdit
```

`show_full_text` only disables the text character limit. It does not remove node count, tree depth, screenshot size, permission, or desktop-session protections.

## Large Or Repeated Snapshots

If a host thread is growing quickly or repeated state checks are polluting context, poll with text-only change output:

```sh
open-computer-use call get_app_state --args '{"app":"TextEdit","include_image":false,"only_changes":true,"max_text_chars":20000}'
```

The first call returns capped state. Later calls return `There has been no change` or a compact diff from the previous accessibility tree. Request `include_image:true` or `force_image:true` only when visual state is needed.

## Element Action Fails

If an element-targeted action fails:

1. Re-run `get_app_state`.
2. Confirm the `element_index` still exists and refers to the intended UI element.
3. Prefer `set_value` for settable text/value controls.
4. Prefer `perform_secondary_action` only for actions exposed in the state result.
5. Use coordinate `click`, `scroll`, or `drag` only after the semantic route is unavailable.

If the action says `The element ID is no longer valid`, the UI changed enough that Open Computer Use could not safely refetch the old target. Run `get_app_state` again and choose the current element index.

## Desktop Session Issues

Windows UI Automation and Linux AT-SPI require a live user desktop. SSH sessions, CI jobs, launch daemons, or services often do not have access to the GUI session even when the CLI binary starts successfully.

If desktop access is missing, ask the user to run the command from the logged-in desktop session or start the target app visibly in that session.

## Permission And Safety Issues

- Do not bypass macOS TCC prompts.
- Do not enable global pointer fallbacks unless the user asks for low-level diagnostic behavior.
- Do not interact with password managers or sensitive apps unless the user explicitly requests it.
- Pause before submitting, sending, deleting, purchasing, or approving anything externally visible.
