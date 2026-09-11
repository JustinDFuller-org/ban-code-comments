package model

type Language string

type Category string

const (
	CategoryOrdinary      Category = "ordinary"
	CategoryDocumentation Category = "documentation"
	CategoryHeader        Category = "header"
	CategoryDirective     Category = "directive"
)

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Finding struct {
	Path     string   `json:"path"`
	Language Language `json:"language"`
	Category Category `json:"category"`
	Range    Range    `json:"range"`
	Text     string   `json:"text"`
}

type Summary struct {
	FilesScanned int `json:"files_scanned"`
	FilesSkipped int `json:"files_skipped"`
	Findings     int `json:"findings"`
}

type Result struct {
	Findings []Finding `json:"findings"`
	Summary  Summary   `json:"summary"`
}

func ExitCode(findings []Finding, err error) int {
	if err != nil {
		return 2
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}
