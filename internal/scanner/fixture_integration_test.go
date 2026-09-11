package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
				if wantFinding {
					finding := findings[0]
					if finding.Path != filepath.ToSlash(path) || finding.Language != model.Language(fixture.name) || finding.Category != model.CategoryOrdinary || !strings.Contains(finding.Text, "finding") {
						t.Fatalf("finding = %#v, want fixture path/language/ordinary category/text", finding)
					}
					if finding.Range.Start.Line < 1 || finding.Range.Start.Column < 1 || finding.Range.End.Line < finding.Range.Start.Line {
						t.Fatalf("finding range = %#v, want positive ordered range", finding.Range)
					}
				}
				if !wantFinding && len(findings) != 0 {
					t.Fatalf("%s fixture findings = %#v, want none", kind, findings)
				}
			}
		})
	}
}

func TestFixtureRegistryCoversEverySupportedExtension(t *testing.T) {
	cases := map[string]model.Language{
		"fixture.go": "go",
		"fixture.js": "javascript", "fixture.jsx": "javascript", "fixture.mjs": "javascript", "fixture.cjs": "javascript",
		"fixture.ts": "typescript", "fixture.tsx": "typescript", "fixture.mts": "typescript", "fixture.cts": "typescript",
		"fixture.py": "python", "fixture.pyw": "python", "fixture.rs": "rust", "fixture.java": "java",
		"fixture.c": "c", "fixture.h": "c", "fixture.cc": "cpp", "fixture.cpp": "cpp", "fixture.cxx": "cpp", "fixture.hh": "cpp", "fixture.hpp": "cpp", "fixture.hxx": "cpp",
		"fixture.cs": "csharp", "fixture.kt": "kotlin", "fixture.kts": "kotlin", "fixture.swift": "swift",
		"fixture.rb": "ruby", "fixture.rake": "ruby", "fixture.php": "php",
		"fixture.sh": "shell", "fixture.bash": "shell", "fixture.zsh": "shell", "fixture.fish": "shell", "fixture.ksh": "shell", "fixture.csh": "shell",
		"fixture.sql": "sql", "fixture.html": "html", "fixture.htm": "html", "fixture.xhtml": "html", "fixture.xml": "xml", "fixture.svg": "xml",
		"fixture.css": "css", "fixture.scss": "scss", "fixture.sass": "scss", "fixture.yaml": "yaml", "fixture.yml": "yaml", "fixture.toml": "toml",
		"fixture.json": "json", "fixture.jsonc": "jsonc", "fixture.hcl": "hcl", "fixture.tf": "terraform", "fixture.tfvars": "terraform",
		"fixture.mk": "makefile", "fixture.mak": "makefile", "fixture.ini": "ini", "fixture.cfg": "ini", "fixture.conf": "ini",
		"Dockerfile": "dockerfile", "Makefile": "makefile",
	}
	for path, want := range cases {
		if got, ok := languages.Lookup(path); !ok || got != want {
			t.Errorf("Lookup(%q) = %q, %v; want %q, true", path, got, ok, want)
		}
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
