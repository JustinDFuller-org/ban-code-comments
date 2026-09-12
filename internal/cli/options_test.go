package cli

import (
	"testing"

	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

func TestParseDefaults(t *testing.T) {
	options, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(options.Paths) != 1 || options.Paths[0] != "." {
		t.Fatalf("paths = %#v, want [.]", options.Paths)
	}
	if !options.Categories[model.CategoryOrdinary] || !options.Categories[model.CategoryDocumentation] {
		t.Fatalf("categories = %#v, want practical defaults", options.Categories)
	}
	if options.Format != FormatJSON {
		t.Fatalf("format = %q, want json", options.Format)
	}
}

func TestParseFiltersAndFormat(t *testing.T) {
	options, err := Parse([]string{"--languages", "go,python", "--categories", "directive", "--include", "src/**", "--exclude", "vendor/**", "--format", "text", "pkg"})
	if err != nil {
		t.Fatal(err)
	}
	if !options.Languages[model.Language("go")] || !options.Languages[model.Language("python")] {
		t.Fatalf("languages = %#v", options.Languages)
	}
	if len(options.Categories) != 1 || !options.Categories[model.CategoryDirective] {
		t.Fatalf("categories = %#v", options.Categories)
	}
	if options.Format != FormatText || len(options.Paths) != 1 || options.Paths[0] != "pkg" {
		t.Fatalf("options = %#v", options)
	}
}

func TestParseRejectsInvalidValues(t *testing.T) {
	for _, args := range [][]string{{"--languages", "unknown"}, {"--categories", "unknown"}, {"--format", "yaml"}} {
		if _, err := Parse(args); err == nil {
			t.Errorf("Parse(%v) accepted invalid value", args)
		}
	}
}

func TestParseHandlesEmptyRepeatedAndShortFlags(t *testing.T) {
	options, err := Parse([]string{"--categories", ",,ordinary", "--include", "", "--exclude", "vendor/**", "--language", "go", "file.go"})
	if err != nil {
		t.Fatal(err)
	}
	if !options.Categories[model.CategoryOrdinary] || len(options.Paths) != 1 || options.Paths[0] != "file.go" {
		t.Fatalf("options = %#v", options)
	}
	if _, err := Parse([]string{"--language"}); err == nil {
		t.Fatal("missing language value accepted")
	}
}
