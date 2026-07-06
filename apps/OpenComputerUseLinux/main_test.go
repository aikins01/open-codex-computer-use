package main

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
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
		App:                 appDescriptor{Name: "Text Editor", BundleIdentifier: "org.gnome.TextEditor", PID: 123},
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
	app, showFullText, err := parseSnapshotArgs([]string{"--show-full-text", "Text Editor"})
	if err != nil {
		t.Fatal(err)
	}
	if app != "Text Editor" || !showFullText {
		t.Fatalf("parseSnapshotArgs = (%q, %v), want (Text Editor, true)", app, showFullText)
	}

	app, showFullText, err = parseSnapshotArgs([]string{"Text Editor"})
	if err != nil {
		t.Fatal(err)
	}
	if app != "Text Editor" || showFullText {
		t.Fatalf("parseSnapshotArgs default = (%q, %v), want (Text Editor, false)", app, showFullText)
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
	args, err := readArguments(`{"app":"Text Editor","pages":2}`, "")
	if err != nil {
		t.Fatal(err)
	}
	if args["app"] != "Text Editor" {
		t.Fatalf("app = %v", args["app"])
	}
	if args["pages"].(json.Number).String() != "2" {
		t.Fatalf("pages = %v", args["pages"])
	}
}

func TestElementIndexAcceptsStringAndJSONNumber(t *testing.T) {
	args, err := readArguments(`{"app":"Text Editor","element_index":0}`, "")
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

func TestCLIHelpMentionsLinuxRuntime(t *testing.T) {
	var out bytes.Buffer
	if err := runCLI([]string{"--help"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Open Computer Use for Linux") {
		t.Fatalf("help text did not mention Linux runtime:\n%s", out.String())
	}
}

func TestLinuxRuntimeDocumentsATSPIAndFallbackBoundary(t *testing.T) {
	if !strings.Contains(linuxRuntimeScript, "Atspi") {
		t.Fatal("Linux runtime must use AT-SPI")
	}
	if !strings.Contains(linuxRuntimeScript, "generate_mouse_event") {
		t.Fatal("Linux runtime should keep coordinate input explicit and visible in the bridge")
	}
	if !strings.Contains(serverInstructions, "not a universal Wayland background input model") {
		t.Fatal("MCP instructions must document the Linux background-input boundary")
	}
}

func TestLinuxRuntimeTextLimitSupportsFullTextMode(t *testing.T) {
	if !strings.Contains(linuxRuntimeScript, "TEXT_CHARACTER_LIMIT = 500") {
		t.Fatal("Linux runtime should define the shared 500 character text limit")
	}
	if !strings.Contains(linuxRuntimeScript, "show_full_text=bool(operation.get(\"show_full_text\", False))") {
		t.Fatal("Linux get_app_state should pass show_full_text into snapshot rendering")
	}
	if !strings.Contains(linuxRuntimeScript, "TEXT_CHARACTER_LIMIT + 1") {
		t.Fatal("Linux default truncation should read one extra character so it can append ellipsis")
	}
}

func TestLinuxRuntimeEnvironmentDiscoversDesktopSession(t *testing.T) {
	runtimeDir := shortTempDir(t)
	listenUnixSocket(t, filepath.Join(runtimeDir, "bus"))
	listenUnixSocket(t, filepath.Join(runtimeDir, "wayland-0"))

	env := envSliceToMap(linuxRuntimeEnvironmentFrom(
		[]string{"PATH=/usr/bin"},
		os.Getuid(),
		[]map[string]string{{
			"XDG_RUNTIME_DIR":     runtimeDir,
			"DISPLAY":             ":1",
			"XAUTHORITY":          "/tmp/open-computer-use-xauth",
			"XDG_SESSION_TYPE":    "wayland",
			"XDG_CURRENT_DESKTOP": "GNOME",
		}},
	))

	if got := env["XDG_RUNTIME_DIR"]; got != runtimeDir {
		t.Fatalf("XDG_RUNTIME_DIR = %q, want %q", got, runtimeDir)
	}
	if got, want := env["DBUS_SESSION_BUS_ADDRESS"], "unix:path="+filepath.Join(runtimeDir, "bus"); got != want {
		t.Fatalf("DBUS_SESSION_BUS_ADDRESS = %q, want %q", got, want)
	}
	if got := env["WAYLAND_DISPLAY"]; got != "wayland-0" {
		t.Fatalf("WAYLAND_DISPLAY = %q, want wayland-0", got)
	}
	if got := env["DISPLAY"]; got != ":1" {
		t.Fatalf("DISPLAY = %q, want :1", got)
	}
	if got := env["XDG_CURRENT_DESKTOP"]; got != "GNOME" {
		t.Fatalf("XDG_CURRENT_DESKTOP = %q, want GNOME", got)
	}
}

func TestLinuxRuntimeEnvironmentCanonicalizesRuntimeBus(t *testing.T) {
	runtimeDir := shortTempDir(t)
	listenUnixSocket(t, filepath.Join(runtimeDir, "bus"))

	env := envSliceToMap(linuxRuntimeEnvironmentFrom(
		[]string{
			"XDG_RUNTIME_DIR=" + runtimeDir,
			"DBUS_SESSION_BUS_ADDRESS=unix:path=" + filepath.Join(runtimeDir, "bus") + ",guid=stale",
		},
		os.Getuid(),
		nil,
	))

	if got, want := env["DBUS_SESSION_BUS_ADDRESS"], "unix:path="+filepath.Join(runtimeDir, "bus"); got != want {
		t.Fatalf("DBUS_SESSION_BUS_ADDRESS = %q, want %q", got, want)
	}
}

func listenUnixSocket(t *testing.T, path string) {
	t.Helper()
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatalf("listen unix socket %s: %v", path, err)
	}
	t.Cleanup(func() {
		_ = listener.Close()
		_ = os.Remove(path)
	})
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

func shortTempDir(t *testing.T) string {
	t.Helper()
	path, err := os.MkdirTemp("/tmp", "ocu-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(path)
	})
	return path
}
