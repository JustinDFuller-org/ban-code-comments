package report

import (
	"bytes"
	"testing"

	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

func TestJSONProducesStableResult(t *testing.T) {
	result := model.Result{Findings: []model.Finding{{Path: "main.go", Language: "go", Category: model.CategoryOrdinary, Range: model.Range{Start: model.Position{Line: 2, Column: 1}, End: model.Position{Line: 2, Column: 10}}, Text: "// note"}}, Summary: model.Summary{FilesScanned: 1, Findings: 1}}
	var output bytes.Buffer
	if err := JSON(&output, result); err != nil {
		t.Fatal(err)
	}
	want := `{"findings":[{"path":"main.go","language":"go","category":"ordinary","range":{"start":{"line":2,"column":1},"end":{"line":2,"column":10}},"text":"// note"}],"summary":{"files_scanned":1,"files_skipped":0,"findings":1}}
`
	if output.String() != want {
		t.Fatalf("JSON = %q, want %q", output.String(), want)
	}
}

func TestTextProducesLineOutput(t *testing.T) {
	findings := []model.Finding{{Path: "main.go", Range: model.Range{Start: model.Position{Line: 2, Column: 1}}, Text: "// note"}}
	var output bytes.Buffer
	if err := Text(&output, findings); err != nil {
		t.Fatal(err)
	}
	if got, want := output.String(), "main.go:2:1: // note\n"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
}
