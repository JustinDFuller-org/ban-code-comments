package discovery

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JustinDFuller-org/ban-code-comments/internal/model"
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

func TestDiscoverDeduplicatesOverlappingPaths(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := writeFile("main.go", "package main\n// finding\n"); err != nil {
		t.Fatal(err)
	}
	discovered, err := Discover(Config{Paths: []string{".", "main.go"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(discovered.Candidates) != 1 {
		t.Fatalf("candidates = %#v, want one unique candidate", discovered.Candidates)
	}
}

func TestDiscoverUsesSuppliedRepositoryRoot(t *testing.T) {
	repository := t.TempDir()
	outside := t.TempDir()
	if err := mkdirAll(filepath.Join(repository, "src")); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(repository, ".gitignore"), "src/ignored.go\n"); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(repository, "src", "ignored.go"), "package main\n// ignored\n"); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("git", "-C", repository, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	t.Chdir(outside)
	discovered, err := Discover(Config{Paths: []string{repository}})
	if err != nil {
		t.Fatal(err)
	}
	if len(discovered.Candidates) != 0 {
		t.Fatalf("discovered = %#v, want ignored file skipped", discovered)
	}
}

func TestDiscoverExcludesExplicitFixedDirectoryRoot(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := mkdirAll("vendor"); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join("vendor", "dependency.go"), "package dependency\n"); err != nil {
		t.Fatal(err)
	}
	discovered, err := Discover(Config{Paths: []string{"vendor"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(discovered.Candidates) != 0 || discovered.Skipped != 1 {
		t.Fatalf("discovered = %#v, want vendor root skipped", discovered)
	}
}

func TestDiscoverFollowsExplicitDirectorySymlink(t *testing.T) {
	t.Chdir(t.TempDir())
	target := filepath.Join(t.TempDir(), "source")
	if err := mkdirAll(target); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(target, "main.go"), "package main\n"); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, "linked"); err != nil {
		t.Fatal(err)
	}
	discovered, err := Discover(Config{Paths: []string{"linked"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(discovered.Candidates) != 1 {
		t.Fatalf("discovered = %#v, want symlink target file", discovered)
	}
}

func TestDiscoverUsesNestedRepositoryIgnoreRules(t *testing.T) {
	repository := t.TempDir()
	nested := filepath.Join(repository, "nested")
	if err := mkdirAll(nested); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(repository, "outer.go"), "package main\n"); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(nested, ".gitignore"), "ignored.go\n"); err != nil {
		t.Fatal(err)
	}
	if err := writeFile(filepath.Join(nested, "ignored.go"), "package nested\n"); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("git", "-C", repository, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("outer git init: %v: %s", err, output)
	}
	if output, err := exec.Command("git", "-C", nested, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("nested git init: %v: %s", err, output)
	}
	discovered, err := Discover(Config{Paths: []string{repository}})
	if err != nil {
		t.Fatal(err)
	}
	for _, candidate := range discovered.Candidates {
		if filepath.Base(candidate.Path) == "ignored.go" {
			t.Fatalf("nested ignored candidate = %#v", candidate)
		}
	}
}

func TestCandidateAndIgnoreHelpers(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "main.go")
	if err := writeFile(path, "package main\n"); err != nil {
		t.Fatal(err)
	}
	if candidate, reason := candidateFor(path, root, "", Config{Languages: map[model.Language]bool{"python": true}}); candidate != nil || reason != "language filter" {
		t.Fatalf("language filter = %#v, %q", candidate, reason)
	}
	if candidate, reason := candidateFor(path, root, "", Config{Includes: []string{"other/**"}}); candidate != nil || reason != "include glob" {
		t.Fatalf("include filter = %#v, %q", candidate, reason)
	}
	if candidate, reason := candidateFor(filepath.Join(root, "notes.txt"), root, "", Config{}); candidate != nil || reason != "unsupported file type" {
		t.Fatalf("unsupported = %#v, %q", candidate, reason)
	}
	if ignored, err := ignoredPaths(nil, ""); err != nil || len(ignored) != 0 {
		t.Fatalf("empty ignored paths = %#v, %v", ignored, err)
	}
}

func TestGlobHelpersHandleEmptyAndQuestionPatterns(t *testing.T) {
	if globMatch("", "file.go") {
		t.Fatal("empty glob matched")
	}
	if !globMatch("src/file.?o", "src/file.go") || globMatch("src/file.?o", "src/file.goo") {
		t.Fatal("question glob mismatch")
	}
	if matchesAny([]string{"", "src/*.go"}, "src/main.go") == false {
		t.Fatal("matchesAny missed pattern")
	}
}
