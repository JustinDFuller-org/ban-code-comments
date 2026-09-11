package discovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func mkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}

func writeFile(path string, contents ...string) error {
	content := "content\n"
	if len(contents) > 0 {
		content = contents[0]
	}
	return os.WriteFile(filepath.Clean(path), []byte(content), 0o644)
}

func TestGlobMatchSupportsRecursiveAndBasenamePatterns(t *testing.T) {
	cases := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"src/**", "src/nested/main.go", true},
		{"*.go", "src/main.go", true},
		{"vendor/**", "src/vendor/main.go", false},
		{"**/*.generated.go", "deep/path/file.generated.go", true},
	}
	for _, testCase := range cases {
		if got := globMatch(testCase.pattern, testCase.path); got != testCase.want {
			t.Errorf("globMatch(%q, %q) = %v, want %v", testCase.pattern, testCase.path, got, testCase.want)
		}
	}
}

func TestDiscoverHonorsGitignoreAndGlobPrecedence(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, directory := range []string{"src", "vendor"} {
		if err := mkdirAll(directory); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"keep.go", "ignored.go", "src/excluded.go", "vendor/dependency.go", "notes.txt"} {
		if err := writeFile(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeFile(".gitignore", "ignored.go\n"); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("git", "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	discovered, err := Discover(Config{Paths: []string{"."}, Includes: []string{"**/*.go"}, Excludes: []string{"src/excluded.go"}, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(discovered.Candidates) != 1 || discovered.Candidates[0].RelPath != "keep.go" {
		t.Fatalf("candidates = %#v, debug = %#v, want keep.go only", discovered.Candidates, discovered.Debug)
	}
	if discovered.Skipped < 3 {
		t.Fatalf("skipped = %d, want ignored, excluded, and vendor files", discovered.Skipped)
	}
	joined := make([]string, 0, len(discovered.Debug))
	for _, diagnostic := range discovered.Debug {
		joined = append(joined, diagnostic.Path+": "+diagnostic.Reason)
	}
	debugOutput := strings.Join(joined, "\n")
	if !strings.Contains(debugOutput, "ignored.go: matched .gitignore") || !strings.Contains(debugOutput, "src/excluded.go: exclude glob") {
		t.Fatalf("debug diagnostics = %q", debugOutput)
	}
}
