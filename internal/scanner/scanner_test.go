package scanner

import (
	"testing"

	"github.com/JustinDFuller-org/ban-code-comments/internal/model"
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

func TestScannerHelpersCoverUnsupportedAndEdgeCases(t *testing.T) {
	if _, ok := syntaxFor("unknown"); ok {
		t.Fatal("unknown syntax was accepted")
	}
	findings := []model.Finding{}
	appendFinding(&findings, "main.go", "go", []byte(""), 0, 0, practicalCategories())
	appendFinding(&findings, "main.go", "go", []byte("// note"), 0, 7, map[model.Category]bool{})
	if len(findings) != 1 || findings[0].Category != model.CategoryOrdinary {
		t.Fatalf("findings = %#v", findings)
	}
	for text, want := range map[string]model.Category{
		"/// docs": model.CategoryDocumentation, "//! docs": model.CategoryDocumentation, "/** docs */": model.CategoryDocumentation,
		"#!/bin/sh": model.CategoryDirective, "# syntax=foo": model.CategoryDirective, "// eslint-disable": model.CategoryDirective,
		"// SPDX-License-Identifier: MIT": model.CategoryHeader, "// generated by tool": model.CategoryHeader, "// ordinary": model.CategoryOrdinary,
	} {
		if got := classify(text); got != want {
			t.Errorf("classify(%q) = %q, want %q", text, got, want)
		}
	}
	if lineLength([]byte("one\ntwo"), 0) != 3 || lineLength([]byte("one"), 0) != 3 {
		t.Fatal("lineLength mismatch")
	}
	if end, ok := skipString([]byte("f\"value\""), 0, "python"); !ok || end != len("f\"value\"") {
		t.Fatalf("prefixed string = %d, %v", end, ok)
	}
	if end, ok := skipString([]byte("@\"value\""), 0, "csharp"); !ok || end != len("@\"value\"") {
		t.Fatalf("verbatim string = %d, %v", end, ok)
	}
	if end, ok := skipString([]byte("#\"value\""), 0, "swift"); !ok || end != len("#\"value\"") {
		t.Fatalf("hash string = %d, %v", end, ok)
	}
	if end, ok := skipString([]byte("R\"tag(comment)tag\""), 0, "cpp"); !ok || end != len("R\"tag(comment)tag\"") {
		t.Fatalf("cpp raw string = %d, %v", end, ok)
	}
	if end, ok := skipCSharpRawString([]byte(`"""value"""`), 0); !ok || end != len(`"""value"""`) {
		t.Fatalf("csharp raw string = %d, %v", end, ok)
	}
}

func TestScannerDelimiterHelpers(t *testing.T) {
	if !isShellCommentStart([]byte("x #"), 2) || isShellCommentStart([]byte("x#"), 1) {
		t.Fatal("shell comment boundary mismatch")
	}
	for _, source := range []string{"# comment", "x # comment", "x;# comment", "x|# comment"} {
		if source == "x # comment" {
			if !isShellCommentStart([]byte(source), 2) {
				t.Fatal("space should start comment")
			}
		}
	}
	if !isYAMLCommentStart([]byte("x #"), 2) || isYAMLCommentStart([]byte("x#"), 1) {
		t.Fatal("yaml boundary mismatch")
	}
	for _, operator := range []string{"#>", "#<", "#-", "#?"} {
		if !isSQLHashOperator([]byte(operator), 0) {
			t.Fatalf("operator %q not recognized", operator)
		}
	}
	if isSQLHashOperator([]byte("#x"), 0) || isSQLHashOperator([]byte("#"), 0) {
		t.Fatal("invalid SQL operator recognized")
	}
	if end := findRubyBlockEnd([]byte("=begin\nbody\n=end\n"), 0); end != len("=begin\nbody\n=end\n") {
		t.Fatalf("ruby end = %d", end)
	}
	if !atLineStart([]byte("a\nb"), 2) || atLineStart([]byte("ab"), 1) {
		t.Fatal("line start mismatch")
	}
}

func TestScannerHeredocAndBlockScalarHelpers(t *testing.T) {
	if end, ok := skipShellHeredoc([]byte("cat <<-EOF\n\t# text\nEOF\n# after\n"), 0); !ok || end >= len("cat <<-EOF\n\t# text\nEOF\n# after\n") {
		t.Fatalf("shell heredoc = %d, %v", end, ok)
	}
	if marker, strip, ok := shellHeredocTerminator([]byte("cat <<-EOF")); !ok || marker != "EOF" || !strip {
		t.Fatalf("shell terminator = %q, %v, %v", marker, strip, ok)
	}
	if marker, _, ok := shellHeredocTerminator([]byte("echo value")); ok || marker != "" {
		t.Fatalf("unexpected shell terminator = %q, %v", marker, ok)
	}
	if end, ok := skipPHPHeredoc([]byte("value = <<<LABEL\ntext\nLABEL;\n"), 0); !ok || end != len("value = <<<LABEL\ntext\nLABEL;\n") {
		t.Fatalf("php heredoc = %d, %v", end, ok)
	}
	if end, ok := skipPHPHeredoc([]byte("value = <<<\"LABEL\"\ntext\nLABEL\n"), 0); !ok || end != len("value = <<<\"LABEL\"\ntext\nLABEL\n") {
		t.Fatalf("php quoted heredoc = %d, %v", end, ok)
	}
	if _, ok := skipPHPHeredoc([]byte("value = 1\n"), 0); ok {
		t.Fatal("non-heredoc accepted")
	}
	if end, ok := skipYAMLBlockScalar([]byte("value: |\n  text\nnext: 1\n"), 0); !ok || end <= 0 {
		t.Fatalf("yaml scalar = %d, %v", end, ok)
	}
	if _, ok := skipYAMLBlockScalar([]byte("value: 1\n"), 0); ok {
		t.Fatal("plain YAML value accepted as scalar")
	}
	for _, source := range [][]byte{[]byte("\n"), []byte("# note\n"), []byte("- |\n  text"), []byte("value: |\n  text")} {
		skipYAMLBlockScalar(source, 0)
	}
	if _, ok := skipShellHeredoc([]byte("cat <<\n"), 0); ok {
		t.Fatal("invalid shell heredoc accepted")
	}
	if _, ok := findShellHeredocEnd([]byte("body"), 0, "END", false); ok {
		t.Fatal("missing shell terminator accepted")
	}
	if _, ok := findPHPHeredocEnd([]byte("body"), 0, "END"); !ok {
		t.Fatal("PHP unterminated heredoc not consumed")
	}
}

func TestScannerUnterminatedAndPrefixedStrings(t *testing.T) {
	for _, testCase := range []struct {
		source   string
		language model.Language
	}{
		{`"unterminated`, "go"}, {"'unterminated", "python"}, {"`unterminated", "javascript"},
		{`"""unterminated`, "python"}, {`r#"unterminated`, "rust"}, {`#"unterminated`, "swift"},
		{`R"tag(unterminated`, "cpp"}, {`"""unterminated`, "csharp"}, {`@$"value"`, "csharp"},
	} {
		if end, ok := skipString([]byte(testCase.source), 0, testCase.language); !ok || end == 0 {
			t.Errorf("skipString(%q, %q) = %d, %v", testCase.source, testCase.language, end, ok)
		}
	}
	if end, ok := skipHashString([]byte(`##"value"##`), 0); !ok || end == 0 {
		t.Fatalf("hash string = %d, %v", end, ok)
	}
	if end, ok := skipRawString([]byte(`r###"value"###`), 0, "rust"); !ok || end == 0 {
		t.Fatalf("rust raw string = %d, %v", end, ok)
	}
	if end, ok := skipCSharpRawString([]byte(`"""unterminated`), 0); !ok || end == 0 {
		t.Fatalf("C# unterminated = %d, %v", end, ok)
	}
	if _, ok := skipHashString([]byte("#value"), 0); ok {
		t.Fatal("invalid hash string accepted")
	}
	if _, ok := skipRawString([]byte("R\"bad space(value)bad space\""), 0, "cpp"); ok {
		t.Fatal("invalid C++ raw string accepted")
	}
	if _, ok := skipRawString([]byte("r\"unterminated"), 0, "rust"); !ok {
		t.Fatal("unterminated Rust raw string not consumed")
	}
	if _, ok := skipString([]byte("\"line\nnext"), 0, "go"); !ok {
		t.Fatal("unterminated line string not consumed")
	}
}
