package scanner

import (
	"testing"

	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

func practicalCategories() map[model.Category]bool {
	return map[model.Category]bool{model.CategoryOrdinary: true, model.CategoryDocumentation: true}
}

func TestScanDetectsCommentAcrossEverySupportedLanguage(t *testing.T) {
	cases := map[model.Language]string{
		"go": "value := 1 // comment\n", "javascript": "const value = 1; // comment\n", "typescript": "const value = 1; // comment\n", "java": "int value = 1; // comment\n",
		"c": "int value = 1; // comment\n", "cpp": "int value = 1; // comment\n", "csharp": "int value = 1; // comment\n", "kotlin": "val value = 1 // comment\n", "swift": "let value = 1 // comment\n", "rust": "let value = 1; // comment\n",
		"python": "value = 1 # comment\n", "ruby": "value = 1 # comment\n", "php": "<?php $value = 1; // comment\n", "shell": "value=1 # comment\n", "yaml": "value: 1 # comment\n", "toml": "value = 1 # comment\n", "dockerfile": "FROM alpine # comment\n", "makefile": "all: # comment\n", "ini": "value=1 ; comment\n",
		"sql": "SELECT 1 -- comment\n", "html": "<p><!-- comment --></p>\n", "xml": "<p><!-- comment --></p>\n", "css": "body { color: red; /* comment */ }\n", "scss": "body { color: red; // comment\n}\n", "jsonc": "{\"value\": 1} // comment\n", "hcl": "value = 1 # comment\n", "terraform": "value = 1 # comment\n",
	}
	for language, source := range cases {
		findings := Scan("fixture", language, []byte(source), practicalCategories())
		if len(findings) != 1 {
			t.Errorf("%s findings = %#v, want one", language, findings)
			continue
		}
		if findings[0].Language != language || findings[0].Category != model.CategoryOrdinary {
			t.Errorf("%s finding metadata = %#v", language, findings[0])
		}
	}
	if findings := Scan("fixture", "json", []byte(`{"value": "// literal"}`), practicalCategories()); len(findings) != 0 {
		t.Fatalf("standard JSON findings = %#v, want none", findings)
	}
}

func TestScanIgnoresLiteralCommentDelimiters(t *testing.T) {
	source := []byte("var a = \"// literal\"; var b = `/* literal */`\nvalue := 1 // actual\n")
	findings := Scan("main.go", "go", source, practicalCategories())
	if len(findings) != 1 || findings[0].Text != "// actual" {
		t.Fatalf("findings = %#v, want actual comment only", findings)
	}

	python := []byte("value = ''' # literal '''\nvalue = 1 # actual\n")
	findings = Scan("main.py", "python", python, practicalCategories())
	if len(findings) != 1 || findings[0].Text != "# actual" {
		t.Fatalf("python findings = %#v, want actual comment only", findings)
	}

	shell := []byte("cat <<EOF\n# literal\nEOF\nvalue=1 # actual\n")
	findings = Scan("run.sh", "shell", shell, practicalCategories())
	if len(findings) != 1 || findings[0].Text != "# actual" {
		t.Fatalf("shell findings = %#v, want actual comment only", findings)
	}
}

func TestScanReportsSourceRangeAndCategories(t *testing.T) {
	source := []byte("package main\n/// docs\n//go:generate tool\n// ordinary\n")
	findings := Scan("main.go", "go", source, map[model.Category]bool{model.CategoryDocumentation: true, model.CategoryDirective: true, model.CategoryOrdinary: true})
	if len(findings) != 3 {
		t.Fatalf("findings = %#v, want three", findings)
	}
	if findings[0].Range.Start.Line != 2 || findings[0].Range.Start.Column != 1 || findings[0].Range.End.Line != 2 {
		t.Fatalf("range = %#v, want line 2", findings[0].Range)
	}
	if findings[0].Category != model.CategoryDocumentation || findings[1].Category != model.CategoryDirective || findings[2].Category != model.CategoryOrdinary {
		t.Fatalf("categories = %#v", findings)
	}
	defaultFindings := Scan("main.go", "go", source, practicalCategories())
	if len(defaultFindings) != 2 || defaultFindings[0].Category != model.CategoryDocumentation || defaultFindings[1].Category != model.CategoryOrdinary {
		t.Fatalf("default findings = %#v", defaultFindings)
	}
}

func TestScanSupportsNestedRustBlocksAndRubyBlocks(t *testing.T) {
	rustFindings := Scan("main.rs", "rust", []byte("/* outer /* inner */ outer */\n"), practicalCategories())
	if len(rustFindings) != 1 || rustFindings[0].Text != "/* outer /* inner */ outer */" {
		t.Fatalf("rust findings = %#v", rustFindings)
	}
	rubyFindings := Scan("README.rb", "ruby", []byte("=begin\n# body\n=end\n"), practicalCategories())
	if len(rubyFindings) != 1 {
		t.Fatalf("ruby findings = %#v", rubyFindings)
	}
}
