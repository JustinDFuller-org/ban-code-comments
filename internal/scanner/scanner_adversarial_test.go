package scanner

import (
	"testing"

	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

func TestAdversarialRawAndQuotedLiteralCases(t *testing.T) {
	selected := map[model.Category]bool{model.CategoryOrdinary: true}
	cases := []struct {
		name     string
		language model.Language
		source   string
		findings int
	}{
		{name: "rust raw internal quote", language: "rust", source: `let value = r#"text " // literal"#;`},
		{name: "cpp raw internal quote", language: "cpp", source: `const char* value = R"tag(text " // literal)tag";`},
		{name: "csharp verbatim doubled quote", language: "csharp", source: `var value = @"text "" // literal"";`},
		{name: "csharp reversed interpolated verbatim", language: "csharp", source: `var value = @$"text "" // literal"";`},
		{name: "shell quoted heredoc", language: "shell", source: "cat <<'EOF'\n# literal\nEOF\necho value # actual\n", findings: 1},
		{name: "shell inline quoted heredoc", language: "shell", source: "printf x; cat <<'EOF'\n# literal\nEOF\necho value # actual\n", findings: 1},
		{name: "shell escaped heredoc marker", language: "shell", source: "cat <<\\EOF\n# literal\nEOF\necho value # actual\n", findings: 1},
		{name: "shell quoted heredoc text", language: "shell", source: "echo \"literal <<EOF\"\necho value # actual\n", findings: 1},
		{name: "shell comment heredoc text", language: "shell", source: "# note <<EOF\nprintf x # actual\n", findings: 2},
		{name: "shell multiple heredocs", language: "shell", source: "cat <<A <<B\n# A\nA\n# B\nB\n", findings: 0},
		{name: "shell hyphenated heredoc", language: "shell", source: "cat <<END-OF-FILE\n# literal\nEND-OF-FILE\necho ok # real\n", findings: 1},
		{name: "shell hash in word", language: "shell", source: "echo foo#bar\n"},
		{name: "shell escaped hash", language: "shell", source: `echo value \# literal`},
		{name: "yaml block scalar", language: "yaml", source: "message: |\n  # literal\nnext: value\n"},
		{name: "yaml plain scalar hash", language: "yaml", source: "value: plain#scalar\n"},
		{name: "swift extended string", language: "swift", source: `let value = #"text " // literal"#`},
		{name: "php heredoc", language: "php", source: "<?php\n$value = <<<TXT\n// literal\nTXT;\n$actual = 1; // actual\n", findings: 1},
		{name: "php heredoc text", language: "php", source: "<?php\n$value = \"<<<TXT\";\n$actual = 1; // actual\n", findings: 1},
		{name: "php block heredoc text", language: "php", source: "/* <<<TXT\nTXT */\n$actual = 1; // actual\n", findings: 2},
		{name: "php heredoc after block", language: "php", source: "<?php\n/* header */ $value = <<<TXT\n// literal\nTXT;\n$actual = 1; // actual\n", findings: 2},
		{name: "yaml indicator comment", language: "yaml", source: "message: | # actual\n  # literal\n", findings: 1},
		{name: "csharp raw four quotes", language: "csharp", source: `var value = """"text // literal""""; // actual`, findings: 1},
		{name: "sql hash operator", language: "sql", source: "SELECT data #> '{a}' FROM table_name;\n"},
		{name: "escaped template backtick", language: "javascript", source: "const value = `text \\` // literal`;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if findings := Scan("fixture", testCase.language, []byte(testCase.source), selected); len(findings) != testCase.findings {
				t.Fatalf("findings = %#v, want %d", findings, testCase.findings)
			}
		})
	}
}

func TestAdversarialShellShebangIsDirective(t *testing.T) {
	findings := Scan("fixture", "shell", []byte("#!/bin/sh\necho value\n"), map[model.Category]bool{model.CategoryOrdinary: true})
	if len(findings) != 0 {
		t.Fatalf("findings = %#v, want shebang suppressed by ordinary policy", findings)
	}
	findings = Scan("fixture", "shell", []byte("#!/bin/sh\necho value\n"), map[model.Category]bool{model.CategoryDirective: true})
	if len(findings) != 1 || findings[0].Category != model.CategoryDirective {
		t.Fatalf("directive findings = %#v, want shebang directive", findings)
	}
}

func TestAdversarialCRLFCommentText(t *testing.T) {
	findings := Scan("fixture", "go", []byte("value := 1 // comment\r\n"), map[model.Category]bool{model.CategoryOrdinary: true})
	if len(findings) != 1 || findings[0].Text != "// comment" {
		t.Fatalf("findings = %#v, want CRLF-free comment text", findings)
	}
}
