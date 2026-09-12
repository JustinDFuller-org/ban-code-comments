package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JustinDFuller-org/ban-code-comments/internal/model"
)

func TestHardPreToolUseBlocksNewFindingForEveryTraditionalTool(t *testing.T) {
	root := t.TempDir()
	inputs := map[string]json.RawMessage{
		"apply_patch": json.RawMessage(`{"patch":"*** Begin Patch\n*** Add File: main.go\n+package main\n+// finding\n*** End Patch\n"}`),
		"edit":        json.RawMessage(`{"path":"main.go","content":"package main\n// finding\n"}`),
		"write":       json.RawMessage(`{"path":"main.go","content":"package main\n// finding\n"}`),
		"write_file":  json.RawMessage(`{"path":"main.go","content":"package main\n// finding\n"}`),
		"file_write":  json.RawMessage(`{"path":"main.go","content":"package main\n// finding\n"}`),
	}
	for toolName, input := range inputs {
		response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: toolName, ToolInput: input}, ModeHard)
		if response.Decision != "block" || !strings.Contains(response.Reason, "main.go:2:1") {
			t.Fatalf("%s response = %#v", toolName, response)
		}
	}
}

func TestWarnPreToolUseReturnsSystemMessageWithoutBlocking(t *testing.T) {
	root := t.TempDir()
	input := `{"patch":"*** Begin Patch\n*** Add File: main.py\n+value = 1\n+# finding\n*** End Patch\n"}`
	response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(input)}, ModeWarn)
	if response.Decision != "" || !strings.Contains(response.SystemMessage, "main.py:2:1") || response.HookSpecificOutput == nil || response.HookSpecificOutput.AdditionalContext != response.SystemMessage {
		t.Fatalf("response = %#v", response)
	}
}

func TestLegacyFindingDoesNotBlockUnrelatedEdit(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	if err := os.WriteFile(path, []byte("package main\n// legacy\nvar value = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	input := `{"patch":"*** Begin Patch\n*** Update File: main.go\n@@\n package main\n // legacy\n-var value = 1\n+var value = 2\n*** End Patch\n"}`
	response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(input)}, ModeHard)
	if response.Decision != "" || response.Reason != "" {
		t.Fatalf("response = %#v", response)
	}
}

func TestProposedLiteralsAndMarkdownAreAllowed(t *testing.T) {
	root := t.TempDir()
	for name, source := range map[string]string{
		"main.go":   "package main\nvar text = \"// literal\"\n",
		"README.md": "# Documentation\n<!-- Markdown comments are allowed -->\n",
		"data.bin":  "// unsupported\n",
	} {
		input, err := json.Marshal(map[string]string{"path": name, "content": source})
		if err != nil {
			t.Fatal(err)
		}
		response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "write_file", ToolInput: input}, ModeHard)
		if response.Decision != "" || response.Reason != "" {
			t.Fatalf("%s response = %#v", name, response)
		}
	}
}

func TestRenameAndDeleteDoNotReportRemovedFinding(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "old.go")
	if err := os.WriteFile(path, []byte("package main\n// legacy\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	input := `{"patch":"*** Begin Patch\n*** Update File: old.go\n*** Move to: new.go\n@@\n package main\n // legacy\n*** End Patch\n"}`
	response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(input)}, ModeHard)
	if response.Decision != "" {
		t.Fatalf("rename response = %#v", response)
	}

	deleteInput := `{"patch":"*** Begin Patch\n*** Delete File: old.go\n*** End Patch\n"}`
	response = Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(deleteInput)}, ModeHard)
	if response.Decision != "" {
		t.Fatalf("delete response = %#v", response)
	}
}

func TestUnsupportedEventsAndToolsAreNoOps(t *testing.T) {
	root := t.TempDir()
	for _, event := range []Event{
		{CWD: root, EventName: "PostToolUse", ToolName: "apply_patch"},
		{CWD: root, EventName: "PreToolUse", ToolName: "Bash", ToolInput: json.RawMessage(`{"command":"echo // finding"}`)},
		{CWD: root, EventName: "PreToolUse", ToolName: "MCP", ToolInput: json.RawMessage(`{"path":"main.go","content":"package main\n// finding\n"}`)},
	} {
		if response := Process(event, ModeHard); response != (Response{}) {
			t.Fatalf("event %#v response = %#v", event, response)
		}
	}
}

func TestMalformedEventProducesModeSpecificOperationalResponse(t *testing.T) {
	var output strings.Builder
	if err := Run(strings.NewReader("not json"), &output, ModeHard); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"decision":"block"`) || !strings.Contains(output.String(), "invalid_event") {
		t.Fatalf("output = %s", output.String())
	}
}

func TestWarnJSONProtocolIncludesCompatibilityAndCodexContext(t *testing.T) {
	root := t.TempDir()
	event := Event{CWD: root, EventName: "PreToolUse", ToolName: "write_file", ToolInput: json.RawMessage(`{"path":"main.go","content":"package main\n// finding\n"}`)}
	input, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	if err := Run(strings.NewReader(string(input)), &output, ModeWarn); err != nil {
		t.Fatal(err)
	}
	var response map[string]any
	if err := json.Unmarshal([]byte(output.String()), &response); err != nil {
		t.Fatal(err)
	}
	if response["systemMessage"] == nil {
		t.Fatalf("response = %#v", response)
	}
	hookSpecificOutput, ok := response["hookSpecificOutput"].(map[string]any)
	if !ok || hookSpecificOutput["additionalContext"] != response["systemMessage"] {
		t.Fatalf("response = %#v", response)
	}
}

func TestParseModeAndProcessBranches(t *testing.T) {
	if got, err := ParseMode(" WARN "); err != nil || got != ModeWarn {
		t.Fatalf("ParseMode = %q, %v", got, err)
	}
	if _, err := ParseMode("other"); err == nil {
		t.Fatal("ParseMode accepted unsupported mode")
	}
	if response := Process(Event{EventName: "unknown"}, ModeHard); response != (Response{}) {
		t.Fatalf("unknown event response = %#v", response)
	}
	if response := Process(Event{EventName: "PreToolUse"}, Mode("other")); response.SystemMessage == "" || !strings.Contains(response.SystemMessage, "invalid_mode") {
		t.Fatalf("invalid mode response = %#v", response)
	}
}

func TestRunAndResponseFormattingBranches(t *testing.T) {
	var output bytes.Buffer
	if err := Run(strings.NewReader("{}"), &output, ModeWarn); err != nil || output.String() != "{}\n" {
		t.Fatalf("empty event output = %q, err = %v", output.String(), err)
	}
	findings := []model.Finding{
		{Path: "b.go", Range: model.Range{Start: model.Position{Line: 2, Column: 1}}, Category: model.CategoryOrdinary, Text: "// b"},
		{Path: "a.go", Range: model.Range{Start: model.Position{Line: 1, Column: 3}}, Category: model.CategoryDocumentation, Text: "/// a"},
	}
	if got := formatFindings(findings); !strings.Contains(got, "a.go:1:3") || !strings.Contains(got, "b.go:2:1") {
		t.Fatalf("formatted findings = %q", got)
	}
	if response := responseForFindings(ModeHard, findings); response.Decision != "block" || response.Reason == "" {
		t.Fatalf("hard response = %#v", response)
	}
	if response := responseForFindings(ModeWarn, findings); response.SystemMessage == "" || response.Decision != "" || response.HookSpecificOutput == nil {
		t.Fatalf("warn response = %#v", response)
	}
	if response := responseForError(ModeWarn, "code", "message"); response.SystemMessage == "" || response.Decision != "" {
		t.Fatalf("warn error response = %#v", response)
	}
}

func TestReconstructChangesSupportsInputForms(t *testing.T) {
	cases := []struct {
		name  string
		input string
		check func(*testing.T, []fileChange)
	}{
		{"string patch", `"*** Begin Patch\n*** Add File: main.go\n+package main\n*** End Patch\n"`, func(t *testing.T, changes []fileChange) {
			if len(changes) != 1 || changes[0].Path != "main.go" {
				t.Fatalf("changes = %#v", changes)
			}
		}},
		{"input patch", `{"input":"*** Begin Patch\n*** Delete File: main.go\n*** End Patch\n"}`, func(t *testing.T, changes []fileChange) {
			if len(changes) != 1 || !changes[0].Delete {
				t.Fatalf("changes = %#v", changes)
			}
		}},
		{"content", `{"filePath":"main.go","newContent":"package main\n"}`, func(t *testing.T, changes []fileChange) {
			if string(changes[0].Source) != "package main\n" {
				t.Fatalf("changes = %#v", changes)
			}
		}},
		{"replacement", `{"path":"main.go","oldString":"old","newString":"new"}`, func(t *testing.T, changes []fileChange) {
			if changes[0].OldText != "old" || changes[0].NewText != "new" {
				t.Fatalf("changes = %#v", changes)
			}
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			changes, err := reconstructChanges(json.RawMessage(testCase.input))
			if err != nil {
				t.Fatal(err)
			}
			testCase.check(t, changes)
		})
	}
	for _, input := range []string{"", "[]", `{}`, `{"path":"main.go"}`, `{"path":"main.go","content":1}`} {
		if _, err := reconstructChanges(json.RawMessage(input)); err == nil {
			t.Errorf("reconstructChanges(%q) accepted invalid input", input)
		}
	}
}

func TestPatchAndPathHelpers(t *testing.T) {
	updated, err := applyUpdatePatch([]byte("one\ntwo\nthree\n"), []string{"@@", " one", "-two", "+changed"})
	if err != nil || string(updated) != "one\nchanged\nthree\n" {
		t.Fatalf("updated = %q, err = %v", updated, err)
	}
	for _, lines := range [][]string{{"@@", "-missing", "+new"}, {"plain"}, {"@@"}} {
		if _, err := applyUpdatePatch([]byte("one\n"), lines); err == nil {
			t.Errorf("applyUpdatePatch(%#v) accepted invalid input", lines)
		}
	}
	root := t.TempDir()
	if path, err := resolvePath(root, "a/main.go"); err != nil || !strings.HasSuffix(path, "main.go") {
		t.Fatalf("resolvePath = %q, %v", path, err)
	}
	for _, name := range []string{"", "../escape"} {
		if _, err := resolvePath(root, name); err == nil {
			t.Errorf("resolvePath accepted %q", name)
		}
	}
	if got := displayPath(root, "nested/main.go"); got != "nested/main.go" {
		t.Fatalf("displayPath = %q", got)
	}
	if !equalLines([]string{"a"}, []string{"a"}) || equalLines([]string{"a"}, []string{"b"}) || equalLines([]string{"a"}, nil) {
		t.Fatal("equalLines mismatch")
	}
}
