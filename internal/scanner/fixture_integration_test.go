package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/JustinDFuller/ban-code-comments/internal/languages"
	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

type fixtureLanguage struct {
	name string
	ext  string
}

func TestFixtureFilesCoverEverySupportedLanguage(t *testing.T) {
	fixtures := []fixtureLanguage{
		{name: "go", ext: "go"},
		{name: "javascript", ext: "js"},
		{name: "typescript", ext: "ts"},
		{name: "python", ext: "py"},
		{name: "rust", ext: "rs"},
		{name: "java", ext: "java"},
		{name: "c", ext: "c"},
		{name: "cpp", ext: "cpp"},
		{name: "csharp", ext: "cs"},
		{name: "kotlin", ext: "kt"},
		{name: "swift", ext: "swift"},
		{name: "ruby", ext: "rb"},
		{name: "php", ext: "php"},
		{name: "shell", ext: "sh"},
		{name: "sql", ext: "sql"},
		{name: "html", ext: "html"},
		{name: "xml", ext: "xml"},
		{name: "css", ext: "css"},
		{name: "scss", ext: "scss"},
		{name: "yaml", ext: "yaml"},
		{name: "toml", ext: "toml"},
		{name: "json", ext: "json"},
		{name: "jsonc", ext: "jsonc"},
		{name: "hcl", ext: "hcl"},
		{name: "terraform", ext: "tf"},
		{name: "dockerfile", ext: "Dockerfile"},
		{name: "makefile", ext: "Makefile"},
		{name: "ini", ext: "ini"},
	}
	for _, fixture := range fixtures {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			for _, kind := range []string{"finding", "clean", "false-positive"} {
				path := fixturePath(t, fixture.name, kind, fixture.ext)
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				language, ok := languages.Lookup(path)
				if !ok || language != model.Language(fixture.name) {
					t.Fatalf("Lookup(%q) = %q, %v; want %q, true", path, language, ok, fixture.name)
				}
				findings := Scan(filepath.ToSlash(path), language, source, allCategories())
				wantFinding := kind == "finding" && fixture.name != "json"
				if wantFinding && len(findings) != 1 {
					t.Fatalf("%s fixture findings = %#v, want one", kind, findings)
				}
				if !wantFinding && len(findings) != 0 {
					t.Fatalf("%s fixture findings = %#v, want none", kind, findings)
				}
			}
		})
	}
}

func allCategories() map[model.Category]bool {
	return map[model.Category]bool{
		model.CategoryOrdinary:      true,
		model.CategoryDocumentation: true,
		model.CategoryHeader:        true,
		model.CategoryDirective:     true,
	}
}

func fixturePath(t *testing.T, language, kind, extension string) string {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	base := extension
	if extension != "Dockerfile" && extension != "Makefile" {
		base = "fixture." + extension
	}
	return filepath.Join(filepath.Dir(sourceFile), "testdata", "fixtures", language, kind, base)
}
