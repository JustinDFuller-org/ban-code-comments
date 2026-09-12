package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
