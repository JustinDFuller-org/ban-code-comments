package languages

import (
	"testing"

	"github.com/JustinDFuller-org/ban-code-comments/internal/model"
)

func TestLookupCoversDocumentedLanguageExamples(t *testing.T) {
	cases := map[string]model.Language{
		"main.go": model.Language("go"), "app.tsx": model.Language("typescript"), "script.py": model.Language("python"),
		"main.rs": model.Language("rust"), "Main.java": model.Language("java"), "main.cpp": model.Language("cpp"),
		"app.cs": model.Language("csharp"), "Main.kt": model.Language("kotlin"), "main.swift": model.Language("swift"),
		"app.rb": model.Language("ruby"), "index.php": model.Language("php"), "run.sh": model.Language("shell"),
		"query.sql": model.Language("sql"), "index.html": model.Language("html"), "doc.xml": model.Language("xml"),
		"site.scss": model.Language("scss"), "config.yaml": model.Language("yaml"), "config.toml": model.Language("toml"),
		"data.jsonc": model.Language("jsonc"), "main.tf": model.Language("terraform"), "Dockerfile": model.Language("dockerfile"),
		"Makefile": model.Language("makefile"), "settings.ini": model.Language("ini"),
	}
	for path, want := range cases {
		got, ok := Lookup(path)
		if !ok || got != want {
			t.Errorf("Lookup(%q) = %q, %v; want %q, true", path, got, ok, want)
		}
	}
}

func TestParseSelectionRejectsUnknownLanguage(t *testing.T) {
	if _, err := ParseSelection([]string{"not-a-language"}); err == nil {
		t.Fatal("ParseSelection accepted an unknown language")
	}
}

func TestParseSelectionHandlesAliasesAndEmptyValues(t *testing.T) {
	selection, err := ParseSelection([]string{" js, c++ ", "", "tf"})
	if err != nil {
		t.Fatal(err)
	}
	if !selection[model.Language("javascript")] || !selection[model.Language("cpp")] || !selection[model.Language("terraform")] {
		t.Fatalf("selection = %#v", selection)
	}
	if got, ok := Lookup("Dockerfile"); !ok || got != model.Language("dockerfile") {
		t.Fatalf("Dockerfile = %q, %v", got, ok)
	}
	if _, ok := Lookup("unknown.xyz"); ok {
		t.Fatal("unknown extension was recognized")
	}
}
