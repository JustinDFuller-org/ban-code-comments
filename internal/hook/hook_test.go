package hook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
