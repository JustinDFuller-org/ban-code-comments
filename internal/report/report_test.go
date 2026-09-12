package report

import (
	"bytes"
	"errors"
	"testing"

	"github.com/JustinDFuller-org/ban-code-comments/internal/model"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

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

func TestReportsReturnWriterErrors(t *testing.T) {
	if err := JSON(failingWriter{}, model.Result{}); err == nil {
		t.Fatal("JSON swallowed writer error")
	}
	if err := Text(failingWriter{}, []model.Finding{{Path: "main.go"}}); err == nil {
		t.Fatal("Text swallowed writer error")
	}
}
