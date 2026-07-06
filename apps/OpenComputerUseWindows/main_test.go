package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestToolDefinitionCount(t *testing.T) {
	if got := len(toolDefinitions()); got != 10 {
		t.Fatalf("toolDefinitions() count = %d, want 10", got)
	}
}

func TestSelectTextSchemaMatchesCodexSurface(t *testing.T) {
	tool := findToolDefinition(t, "select_text")
	properties := tool.InputSchema["properties"].(map[string]any)
	selection := properties["selection"].(map[string]any)
	required := tool.InputSchema["required"].([]string)

	if tool.Description != "Select text inside a text element, or place the text cursor before or after it. Provide text exactly as it appears in the accessibility tree, including any Markdown formatting. If the text is not unique, provide surrounding prefix or suffix text to disambiguate it." {
		t.Fatalf("unexpected select_text description: %q", tool.Description)
	}
	if strings.Join(required, ",") != "app,element_index,text" {
		t.Fatalf("required = %#v, want [app element_index text]", required)
	}
	if strings.Join(selection["enum"].([]string), ",") != "text,cursor_before,cursor_after" {
		t.Fatalf("selection enum = %#v", selection["enum"])
	}
}

func TestGetAppStateSchemaIncludesOutputControls(t *testing.T) {
	tool := findToolDefinition(t, "get_app_state")
	properties := tool.InputSchema["properties"].(map[string]any)
	showFullText := properties["show_full_text"].(map[string]any)
	if got := showFullText["type"]; got != "boolean" {
		t.Fatalf("show_full_text type = %v, want boolean", got)
	}
	for _, key := range []string{"include_image", "force_image", "only_changes"} {
		property := properties[key].(map[string]any)
		if got := property["type"]; got != "boolean" {
			t.Fatalf("%s type = %v, want boolean", key, got)
		}
	}
	maxTextChars := properties["max_text_chars"].(map[string]any)
	if got := maxTextChars["type"]; got != "integer" {
		t.Fatalf("max_text_chars type = %v, want integer", got)
	}
	if got := maxTextChars["minimum"]; got != 0 {
		t.Fatalf("max_text_chars minimum = %v, want 0", got)
	}
	required := tool.InputSchema["required"].([]string)
	if len(required) != 1 || required[0] != "app" {
		t.Fatalf("required = %#v, want [app]", required)
	}
}

func TestAppStateResultOutputControls(t *testing.T) {
	maxChars := 40
	snapshot := &appSnapshot{
		App:                 appDescriptor{Name: "Notepad", BundleIdentifier: "notepad", PID: 123},
		WindowTitle:         "Draft",
		ScreenshotPNGBase64: "image-a",
		TreeLines:           []string{"0123456789abcdef"},
	}

	result, current := snapshot.appStateResult(appStateOutputOptions{IncludeImage: false, MaxTextChars: &maxChars}, nil)
	if len(result.Content) != 1 || result.Content[0].Type != "text" {
		t.Fatalf("content = %#v, want text only", result.Content)
	}
	if !strings.Contains(result.Content[0].Text, "[truncated after 40 characters]") {
		t.Fatalf("text was not capped: %q", result.Content[0].Text)
	}
	if got := len([]rune(result.Content[0].Text)); got > maxChars {
		t.Fatalf("text length = %d, want at most %d", got, maxChars)
	}

	imageAfterTextOnly, currentWithImage := snapshot.appStateResult(defaultAppStateOutputOptions(), &current)
	if len(imageAfterTextOnly.Content) != 2 || imageAfterTextOnly.Content[1].Type != "image" {
		t.Fatalf("image after text-only content = %#v, want text and image", imageAfterTextOnly.Content)
	}
	imageAfterTextOnlyChangePoll, _ := snapshot.appStateResult(appStateOutputOptions{IncludeImage: true, OnlyChanges: true}, &current)
	if len(imageAfterTextOnlyChangePoll.Content) != 2 || imageAfterTextOnlyChangePoll.Content[0].Text != appStateNoChangeMessage || imageAfterTextOnlyChangePoll.Content[1].Type != "image" {
		t.Fatalf("image after text-only change poll content = %#v, want no-change text and image", imageAfterTextOnlyChangePoll.Content)
	}
	deduped, _ := snapshot.appStateResult(defaultAppStateOutputOptions(), &currentWithImage)
	if len(deduped.Content) != 1 {
		t.Fatalf("deduped content = %#v, want text only", deduped.Content)
	}
	forced, _ := snapshot.appStateResult(appStateOutputOptions{IncludeImage: true, ForceImage: true}, &currentWithImage)
	if len(forced.Content) != 2 || forced.Content[1].Type != "image" {
		t.Fatalf("forced content = %#v, want text and image", forced.Content)
	}
	unchanged, _ := snapshot.appStateResult(appStateOutputOptions{IncludeImage: true, OnlyChanges: true}, &currentWithImage)
	if unchanged.Content[0].Text != appStateNoChangeMessage {
		t.Fatalf("unchanged text = %q, want no-change message", unchanged.Content[0].Text)
	}
	screenshotOnlyChanged := *snapshot
	screenshotOnlyChanged.ScreenshotPNGBase64 = "image-b"
	screenshotOnlyChangedResult, screenshotOnlyChangedCache := screenshotOnlyChanged.appStateResult(appStateOutputOptions{IncludeImage: true, OnlyChanges: true}, &currentWithImage)
	if len(screenshotOnlyChangedResult.Content) != 1 || screenshotOnlyChangedResult.Content[0].Text != appStateNoChangeMessage {
		t.Fatalf("screenshot-only change content = %#v, want no-change text only", screenshotOnlyChangedResult.Content)
	}
	repeatedScreenshotOnlyChangedResult, _ := screenshotOnlyChanged.appStateResult(appStateOutputOptions{IncludeImage: true, OnlyChanges: true}, &screenshotOnlyChangedCache)
	if len(repeatedScreenshotOnlyChangedResult.Content) != 1 || repeatedScreenshotOnlyChangedResult.Content[0].Text != appStateNoChangeMessage {
		t.Fatalf("repeated screenshot-only change content = %#v, want no-change text only", repeatedScreenshotOnlyChangedResult.Content)
	}
	fullAfterScreenshotOnlyChanged, _ := screenshotOnlyChanged.appStateResult(defaultAppStateOutputOptions(), &screenshotOnlyChangedCache)
	if len(fullAfterScreenshotOnlyChanged.Content) != 2 || fullAfterScreenshotOnlyChanged.Content[1].Type != "image" {
		t.Fatalf("full state after screenshot-only change content = %#v, want text and image", fullAfterScreenshotOnlyChanged.Content)
	}
	changedSnapshot := *snapshot
	changedSnapshot.TreeLines = []string{"changed"}
	changed, _ := changedSnapshot.appStateResult(appStateOutputOptions{IncludeImage: true, OnlyChanges: true}, &currentWithImage)
	if !strings.Contains(changed.Content[0].Text, appStateDiffHeader) || !strings.Contains(changed.Content[0].Text, "~ 0123456789abcdef -> changed") {
		t.Fatalf("changed text = %q, want compact diff", changed.Content[0].Text)
	}
}

func TestMaxTextCharsRejectsInvalidValues(t *testing.T) {
	for _, value := range []any{json.Number("0"), json.Number("10"), float64(2), 3, int64(4)} {
		if _, ok := optionalNonNegativeInt(map[string]any{"max_text_chars": value}, "max_text_chars"); !ok {
			t.Fatalf("max_text_chars %v should be valid", value)
		}
	}
	for _, value := range []any{json.Number("-1"), json.Number("1.5"), json.Number("9223372036854775808"), float64(2.2), true, "10"} {
		if _, ok := optionalNonNegativeInt(map[string]any{"max_text_chars": value}, "max_text_chars"); ok {
			t.Fatalf("max_text_chars %v should be invalid", value)
		}
	}
}

func TestParseSnapshotArgsSupportsShowFullText(t *testing.T) {
	app, showFullText, err := parseSnapshotArgs([]string{"--show-full-text", "Notepad"})
	if err != nil {
		t.Fatal(err)
	}
	if app != "Notepad" || !showFullText {
		t.Fatalf("parseSnapshotArgs = (%q, %v), want (Notepad, true)", app, showFullText)
	}

	app, showFullText, err = parseSnapshotArgs([]string{"Notepad"})
	if err != nil {
		t.Fatal(err)
	}
	if app != "Notepad" || showFullText {
		t.Fatalf("parseSnapshotArgs default = (%q, %v), want (Notepad, false)", app, showFullText)
	}
}

func TestCallSequenceStopsAfterFirstToolError(t *testing.T) {
	output, hasError, err := runCallCommand([]string{
		"--calls",
		`[{"tool":"not_a_tool"},{"tool":"list_apps"}]`,
	}, newService())
	if err != nil {
		t.Fatal(err)
	}
	if !hasError {
		t.Fatal("expected hasError")
	}
	items, ok := output.([]map[string]any)
	if !ok {
		t.Fatalf("output type = %T", output)
	}
	if len(items) != 1 {
		t.Fatalf("sequence output count = %d, want 1", len(items))
	}
}

func TestReadArgumentsAcceptsJSONObject(t *testing.T) {
	args, err := readArguments(`{"app":"Notepad","pages":2}`, "")
	if err != nil {
		t.Fatal(err)
	}
	if args["app"] != "Notepad" {
		t.Fatalf("app = %v", args["app"])
	}
	if args["pages"].(json.Number).String() != "2" {
		t.Fatalf("pages = %v", args["pages"])
	}
}

func TestElementIndexAcceptsStringAndJSONNumber(t *testing.T) {
	args, err := readArguments(`{"app":"Notepad","element_index":0}`, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := optionalElementIndex(args); got != "0" {
		t.Fatalf("numeric element_index = %q, want 0", got)
	}
	if got := optionalElementIndex(map[string]any{"element_index": "14"}); got != "14" {
		t.Fatalf("string element_index = %q, want 14", got)
	}
	if got := optionalElementIndex(map[string]any{"element_index": json.Number("1.5")}); got != "" {
		t.Fatalf("fractional element_index = %q, want empty", got)
	}
}

func TestMCPInitializeResponseContainsToolsCapability(t *testing.T) {
	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      float64(1),
		"method":  "initialize",
		"params":  map[string]any{},
	}
	response := handleMCPRequest(request, newService())
	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatalf("missing result: %#v", response)
	}
	capabilities := result["capabilities"].(map[string]any)
	if _, ok := capabilities["tools"]; !ok {
		t.Fatalf("missing tools capability: %#v", capabilities)
	}
}

func TestMCPTurnEndedClearsRuntimeState(t *testing.T) {
	svc := newService()
	svc.snapshots["app"] = &appSnapshot{App: appDescriptor{Name: "App", PID: 1}}
	svc.appStateOutputs["app"] = appStateOutputCache{RenderedText: "old"}

	response := handleMCPRequest(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/turn-ended",
		"params":  map[string]any{"type": "agent-turn-complete"},
	}, svc)

	if response != nil {
		t.Fatalf("turn-ended response = %#v, want nil", response)
	}
	if len(svc.snapshots) != 0 || len(svc.appStateOutputs) != 0 {
		t.Fatalf("runtime state was not cleared: snapshots=%d outputs=%d", len(svc.snapshots), len(svc.appStateOutputs))
	}
}

func TestCLIHelpMentionsWindowsRuntime(t *testing.T) {
	var out bytes.Buffer
	if err := runCLI([]string{"--help"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Open Computer Use for Windows") {
		t.Fatalf("help text did not mention Windows runtime:\n%s", out.String())
	}
}

func TestWindowsRuntimeForegroundActionsRequireOptIn(t *testing.T) {
	if !strings.Contains(windowsRuntimeScript, "OPEN_COMPUTER_USE_WINDOWS_ALLOW_APP_LAUNCH") {
		t.Fatal("Windows app launch fallback must remain opt-in")
	}
	if !strings.Contains(windowsRuntimeScript, "OPEN_COMPUTER_USE_WINDOWS_ALLOW_FOCUS_ACTIONS") {
		t.Fatal("Windows SetFocus action must remain opt-in")
	}
	if !strings.Contains(windowsRuntimeScript, "OPEN_COMPUTER_USE_WINDOWS_ALLOW_UIA_TEXT_FALLBACK") {
		t.Fatal("Windows UIA text fallback must remain opt-in")
	}
	if !strings.Contains(windowsRuntimeScript, "OPEN_COMPUTER_USE_WINDOWS_ALLOW_UIA_TEXT_SELECTION") {
		t.Fatal("Windows UIA text selection must remain opt-in")
	}
	if !strings.Contains(serverInstructions, "does not auto-launch apps, perform SetFocus, use UIA text fallback, or perform UIA text selection by default") {
		t.Fatal("MCP instructions must document the Windows background-focus policy")
	}
}

func TestUTF8EncodingInPowerShellScript(t *testing.T) {
	// Verify that the PowerShell script sets UTF-8 encoding
	if !strings.Contains(windowsRuntimeScript, "$OutputEncoding = [System.Text.Encoding]::UTF8") {
		t.Fatal("PowerShell script must set $OutputEncoding to UTF-8 for proper non-ASCII character handling")
	}
	if !strings.Contains(windowsRuntimeScript, "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8") {
		t.Fatal("PowerShell script must set [Console]::OutputEncoding to UTF-8 for proper non-ASCII character handling")
	}
}

func TestWindowsRuntimeTextLimitSupportsFullTextMode(t *testing.T) {
	if !strings.Contains(windowsRuntimeScript, "$TextCharacterLimit = 500") {
		t.Fatal("Windows runtime should define the shared 500 character text limit")
	}
	if !strings.Contains(windowsRuntimeScript, "Build-Snapshot $operation.app ([bool]$operation.show_full_text)") {
		t.Fatal("Windows get_app_state should pass show_full_text into snapshot rendering")
	}
	if !strings.Contains(windowsRuntimeScript, "$maxLength = if ($ShowFullText) { -1 } else { $script:TextCharacterLimit + 1 }") {
		t.Fatal("Windows selected text should use full UIA text only in full-text mode")
	}
}

func findToolDefinition(t *testing.T, name string) toolDefinition {
	t.Helper()
	for _, tool := range toolDefinitions() {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("missing tool definition %q", name)
	return toolDefinition{}
}
