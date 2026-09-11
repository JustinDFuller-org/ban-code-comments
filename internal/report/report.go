package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

func JSON(writer io.Writer, result model.Result) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(result)
}

func Text(writer io.Writer, findings []model.Finding) error {
	for _, finding := range findings {
		if _, err := fmt.Fprintf(writer, "%s:%d:%d: %s\n", finding.Path, finding.Range.Start.Line, finding.Range.Start.Column, finding.Text); err != nil {
			return err
		}
	}
	return nil
}
