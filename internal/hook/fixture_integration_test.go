package hook

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/JustinDFuller/ban-code-comments/internal/languages"
	"github.com/JustinDFuller/ban-code-comments/internal/model"
	"github.com/JustinDFuller/ban-code-comments/internal/scanner"
)

func TestCheckedInHookEventsDecode(t *testing.T) {
	for _, name := range []string{"pre-tool-use.json", "post-tool-use.json"} {
		contents, err := os.ReadFile(filepath.Join(hookTestdataRoot(t), "events", name))
		if err != nil {
			t.Fatal(err)
		}
		var event Event
		if err := json.Unmarshal(contents, &event); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if event.EventName == "" || event.ToolName == "" || len(event.ToolInput) == 0 && event.EventName == "PreToolUse" {
			t.Fatalf("%s decoded incomplete event: %#v", name, event)
		}
	}
}

func TestHookProposedFixturesCoverEveryRegisteredLanguage(t *testing.T) {
	fixtureRoot := scannerFixtureRoot(t)
	for _, language := range languages.Supported() {
		language := language
		t.Run(string(language), func(t *testing.T) {
			findingPath := findLanguageFixture(t, fixtureRoot, language, "finding")
			cleanPath := findLanguageFixture(t, fixtureRoot, language, "clean")
			falsePositivePath := findLanguageFixture(t, fixtureRoot, language, "false-positive")
			for _, fixture := range []struct {
				name        string
				path        string
				wantFinding bool
			}{
				{name: "finding", path: findingPath, wantFinding: language != model.Language("json")},
				{name: "clean", path: cleanPath},
				{name: "false-positive", path: falsePositivePath},
			} {
				source, err := os.ReadFile(fixture.path)
				if err != nil {
					t.Fatal(err)
				}
				if got, ok := languages.Lookup(fixture.path); !ok || got != language {
					t.Fatalf("Lookup(%q) = %q, %v; want %q, true", fixture.path, got, ok, language)
				}
				input, err := json.Marshal(map[string]any{
					"path":    filepath.Base(fixture.path),
					"content": string(source),
				})
				if err != nil {
					t.Fatal(err)
				}
				response := Process(Event{CWD: t.TempDir(), EventName: "PreToolUse", ToolName: "write_file", ToolInput: input}, ModeHard, t.TempDir())
				if fixture.wantFinding && response.Decision != "block" {
					t.Fatalf("%s response = %#v; scanner = %#v", fixture.name, response, scanner.Scan(filepath.Base(fixture.path), language, source, selectedCategories))
				}
				if !fixture.wantFinding && response.Decision != "" {
					t.Fatalf("%s response = %#v", fixture.name, response)
				}
			}
		})
	}
}

func TestHookProposedFixturesCoverNonCodeAndLegacyCases(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"README.md", "unsupported.txt"} {
		source, err := os.ReadFile(filepath.Join(hookTestdataRoot(t), "proposed", name))
		if err != nil {
			t.Fatal(err)
		}
		input, err := json.Marshal(map[string]any{"path": name, "content": string(source)})
		if err != nil {
			t.Fatal(err)
		}
		response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "write_file", ToolInput: input}, ModeHard, t.TempDir())
		if response.Decision != "" {
			t.Fatalf("%s response = %#v", name, response)
		}
	}

	legacyPath := filepath.Join(root, "legacy.go")
	legacy, err := os.ReadFile(filepath.Join(hookTestdataRoot(t), "proposed", "legacy.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	input, err := json.Marshal(map[string]any{"path": "legacy.go", "old_string": "var value = 1", "new_string": "var value = 2"})
	if err != nil {
		t.Fatal(err)
	}
	response := Process(Event{CWD: root, EventName: "PreToolUse", ToolName: "edit", ToolInput: input}, ModeHard, t.TempDir())
	if response.Decision != "" {
		t.Fatalf("legacy response = %#v", response)
	}
}

func TestOpaquePostAuditCoversCleanWriteAndMissingState(t *testing.T) {
	root := t.TempDir()
	stateDirectory := t.TempDir()
	event := Event{SessionID: "session", TurnID: "turn", ToolCallID: "call", CWD: root, EventName: "PreToolUse", ToolName: "Bash"}
	if response := Process(event, ModeHard, stateDirectory); response.Decision != "" {
		t.Fatalf("pre response = %#v", response)
	}
	if err := os.WriteFile(filepath.Join(root, "clean.go"), []byte("package clean\nvar value = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	event.EventName = "PostToolUse"
	if response := Process(event, ModeHard, stateDirectory); response.Decision != "" {
		t.Fatalf("clean post response = %#v", response)
	}
	if response := Process(event, ModeHard, stateDirectory); response.Decision != "block" || !strings.Contains(response.Reason, "workspace_state_missing") {
		t.Fatalf("missing state response = %#v", response)
	}
}

func TestHookEventFixtureRunsThroughJSONProtocol(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join(hookTestdataRoot(t), "events", "pre-tool-use.json"))
	if err != nil {
		t.Fatal(err)
	}
	var event Event
	if err := json.Unmarshal(contents, &event); err != nil {
		t.Fatal(err)
	}
	event.CWD = t.TempDir()
	var output bytes.Buffer
	if err := Run(bytes.NewReader(contentsWithEventCWD(t, event)), &output, ModeHard, t.TempDir()); err != nil {
		t.Fatal(err)
	}
	var response Response
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Decision != "block" || !strings.Contains(response.Reason, "main.go") {
		t.Fatalf("response = %#v", response)
	}
}

func findLanguageFixture(t *testing.T, root string, language model.Language, kind string) string {
	directory := filepath.Join(root, string(language), kind)
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(paths)
	if len(paths) != 1 {
		t.Fatalf("%s/%s fixture files = %v; want one", language, kind, paths)
	}
	return paths[0]
}

func scannerFixtureRoot(t *testing.T) string {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(sourceFile), "..", "scanner", "testdata", "fixtures")
}

func hookTestdataRoot(t *testing.T) string {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Join(filepath.Dir(sourceFile), "testdata")
}

func contentsWithEventCWD(t *testing.T, event Event) []byte {
	contents, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}
