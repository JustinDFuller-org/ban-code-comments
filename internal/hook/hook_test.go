package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

func TestParseModeAndProcessBranches(t *testing.T) {
	if got, err := ParseMode(" WARN "); err != nil || got != ModeWarn {
		t.Fatalf("ParseMode = %q, %v", got, err)
	}
	if _, err := ParseMode("other"); err == nil {
		t.Fatal("ParseMode accepted unsupported mode")
	}
	if response := Process(Event{EventName: "unknown"}, ModeHard, t.TempDir()); response != (Response{}) {
		t.Fatalf("unknown event response = %#v", response)
	}
	if response := Process(Event{EventName: "PreToolUse"}, Mode("other"), t.TempDir()); response.SystemMessage == "" || !strings.Contains(response.SystemMessage, "invalid_mode") {
		t.Fatalf("invalid mode response = %#v", response)
	}
}

func TestRunAndResponseFormattingBranches(t *testing.T) {
	var output bytes.Buffer
	if err := Run(strings.NewReader("{}"), &output, ModeWarn, t.TempDir()); err != nil || output.String() != "{}\n" {
		t.Fatalf("empty event output = %q, err = %v", output.String(), err)
	}
	findings := []model.Finding{
		{Path: "b.go", Range: model.Range{Start: model.Position{Line: 2, Column: 1}}, Category: model.CategoryOrdinary, Text: "// b"},
		{Path: "a.go", Range: model.Range{Start: model.Position{Line: 1, Column: 3}}, Category: model.CategoryDocumentation, Text: "/// a"},
	}
	if got := formatFindings(findings); !strings.Contains(got, "a.go:1:3") || !strings.Contains(got, "b.go:2:1") {
		t.Fatalf("formatted findings = %q", got)
	}
	if response := responseForFindings(ModeHard, findings, true); response.Continue == nil || response.StopReason == "" {
		t.Fatalf("post hard response = %#v", response)
	}
	if response := responseForFindings(ModeWarn, findings, false); response.SystemMessage == "" || response.Decision != "" {
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
			t.Errorf("applyUpdatePatch(%#v) accepted invalid patch", lines)
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
	if got := displayPath(root, "../escape"); got != "../escape" {
		t.Fatalf("displayPath invalid = %q", got)
	}
	if !equalLines([]string{"a"}, []string{"a"}) || equalLines([]string{"a"}, []string{"b"}) || equalLines([]string{"a"}, nil) {
		t.Fatal("equalLines mismatch")
	}
}

func TestWorkspaceStateAndComparisonBranches(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "vendor"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "vendor", "ignored.go"), []byte("package ignored\n// finding\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	state, err := captureWorkspace(root)
	if err != nil || len(state.Files) != 1 {
		t.Fatalf("state = %#v, err = %v", state, err)
	}
	event := Event{SessionID: "s", ToolCallID: "c", CWD: root}
	directory := t.TempDir()
	if err := saveState(directory, event, state); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadState(directory, event, root)
	if err != nil || len(loaded.Files) != 1 {
		t.Fatalf("loaded = %#v, err = %v", loaded, err)
	}
	if _, err := loadState(directory, event, t.TempDir()); err == nil {
		t.Fatal("loadState accepted mismatched root")
	}
	removeState(directory, event)
	if _, err := loadState(directory, event, root); err == nil {
		t.Fatal("removed state still loaded")
	}
	before := workspaceState{Files: map[string][]byte{"main.go": []byte("package main\n")}}
	after := workspaceState{Files: map[string][]byte{"main.go": []byte("package main\n// new\n"), "new.py": []byte("# finding\n")}}
	findings, err := compareWorkspace(before, after)
	if err != nil || len(findings) != 2 {
		t.Fatalf("comparison = %#v, err = %v", findings, err)
	}
	if skippedDirectory("vendor") != true || skippedDirectory("src") {
		t.Fatal("skippedDirectory mismatch")
	}
}

func TestProcessReportsProposalAndWorkspaceErrors(t *testing.T) {
	root := t.TempDir()
	for _, event := range []Event{
		{CWD: root, EventName: "PreToolUse", ToolName: "write_file", ToolInput: json.RawMessage(`{"path":"main.go"}`)},
		{CWD: root, EventName: "PreToolUse", ToolName: "write_file", ToolInput: json.RawMessage(`[]`)},
		{CWD: root, EventName: "PostToolUse", ToolName: "Bash", SessionID: "missing", ToolCallID: "missing"},
	} {
		response := Process(event, ModeHard, t.TempDir())
		if response.Decision != "block" || response.Reason == "" {
			t.Errorf("event %#v response = %#v", event, response)
		}
	}
	if response := processPre(Event{ToolInput: json.RawMessage(`{"path":"main.go","old_string":"old","new_string":"new"}`)}, "write", root, ModeHard, t.TempDir()); response.Decision != "block" {
		t.Fatalf("replacement error response = %#v", response)
	}
}

func TestHookParsingAndWorkspaceEdgeCases(t *testing.T) {
	if _, err := parsePatch("*** Begin Patch\nunknown\n*** End Patch\n"); err == nil {
		t.Fatal("unsupported patch accepted")
	}
	if _, err := applyUpdatePatch([]byte("old\n"), []string{"@@", "plain"}); err == nil {
		t.Fatal("unsupported hunk accepted")
	}
	if got := statePath("", Event{}); got == "" {
		t.Fatal("empty state path")
	}
	if got := displayPath(t.TempDir(), "/absolute.go"); got != "/absolute.go" {
		t.Fatalf("absolute display path = %q", got)
	}
	if got := displayPath(t.TempDir(), "../outside.go"); got != "../outside.go" {
		t.Fatalf("outside display path = %q", got)
	}
	if _, err := reconstructChanges(json.RawMessage(`{"path":1,"content":"x"}`)); err == nil {
		t.Fatal("invalid path type accepted")
	}
	if got := rawString(map[string]json.RawMessage{"value": json.RawMessage("1")}, "value"); got != "" {
		t.Fatalf("rawString = %q", got)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".ban-code-comments-hook-state.json"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "node_modules"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := captureWorkspace(root); err != nil {
		t.Fatal(err)
	}
	changes := []fileChange{{Path: "missing.go", Source: []byte("package main\n// note\n")}}
	if findings, err := evaluateChanges(root, changes); err != nil || len(findings) != 1 {
		t.Fatalf("new file findings = %#v, %v", findings, err)
	}
}

func TestHardPreToolUseBlocksNewApplyPatchFinding(t *testing.T) {
	root := t.TempDir()
	input := `{"patch":"*** Begin Patch\n*** Add File: main.go\n+package main\n+// finding\n*** End Patch\n"}`
	response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(input)}, ModeHard, t.TempDir())
	if response.Decision != "block" || !strings.Contains(response.Reason, "main.go:2:1") {
		t.Fatalf("response = %#v", response)
	}
}

func TestWarnPreToolUseReturnsSystemMessageWithoutBlocking(t *testing.T) {
	root := t.TempDir()
	input := `{"patch":"*** Begin Patch\n*** Add File: main.py\n+value = 1\n+# finding\n*** End Patch\n"}`
	response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(input)}, ModeWarn, t.TempDir())
	if response.Decision != "" || !strings.Contains(response.SystemMessage, "main.py:2:1") {
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
	response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(input)}, ModeHard, t.TempDir())
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
		response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "write_file", ToolInput: input}, ModeHard, t.TempDir())
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
	response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(input)}, ModeHard, t.TempDir())
	if response.Decision != "" {
		t.Fatalf("rename response = %#v", response)
	}

	deleteInput := `{"patch":"*** Begin Patch\n*** Delete File: old.go\n*** End Patch\n"}`
	response = Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "apply_patch", ToolInput: json.RawMessage(deleteInput)}, ModeHard, t.TempDir())
	if response.Decision != "" {
		t.Fatalf("delete response = %#v", response)
	}
}

func TestOpaqueBashPostAuditUsesPreToolWorkspaceBaseline(t *testing.T) {
	root := t.TempDir()
	stateDirectory := t.TempDir()
	event := Event{SessionID: "session", TurnID: "turn", ToolCallID: "call", CWD: root, EventName: "PreToolUse", ToolName: "Bash", ToolInput: json.RawMessage(`{"command":"generator"}`)}
	if response := Process(event, ModeHard, stateDirectory); response.Decision != "" {
		t.Fatalf("pre response = %#v", response)
	}
	if err := os.WriteFile(filepath.Join(root, "generated.go"), []byte("package main\n// generated finding\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	event.EventName = "PostToolUse"
	response := Process(event, ModeHard, stateDirectory)
	if response.Decision != "block" || response.StopReason == "" || response.Continue == nil || *response.Continue {
		t.Fatalf("post response = %#v", response)
	}
}

func TestOpaqueBashWarnAllowsContinuation(t *testing.T) {
	root := t.TempDir()
	stateDirectory := t.TempDir()
	event := Event{SessionID: "session", ToolCallID: "call", CWD: root, EventName: "PreToolUse", ToolName: "Bash", ToolInput: json.RawMessage(`{"command":"generator"}`)}
	Process(event, ModeWarn, stateDirectory)
	if err := os.WriteFile(filepath.Join(root, "generated.go"), []byte("package main\n// generated finding\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	event.EventName = "PostToolUse"
	response := Process(event, ModeWarn, stateDirectory)
	if response.Decision != "" || !strings.Contains(response.SystemMessage, "newly introduced") {
		t.Fatalf("post response = %#v", response)
	}
}

func TestMalformedEventProducesModeSpecificOperationalResponse(t *testing.T) {
	var output strings.Builder
	if err := Run(strings.NewReader("not json"), &output, ModeHard, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"decision":"block"`) || !strings.Contains(output.String(), "invalid_event") {
		t.Fatalf("output = %s", output.String())
	}
}
