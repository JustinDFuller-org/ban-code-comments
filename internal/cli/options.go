package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/JustinDFuller-org/ban-code-comments/internal/languages"
	"github.com/JustinDFuller-org/ban-code-comments/internal/model"
)

type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

type Options struct {
	Paths      []string
	Languages  map[model.Language]bool
	Categories map[model.Category]bool
	Includes   []string
	Excludes   []string
	Format     Format
	Debug      bool
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func Parse(args []string) (Options, error) {
	var languageValues stringList
	var categoryValues stringList
	var includes stringList
	var excludes stringList
	var format string
	var debug bool

	flags := flag.NewFlagSet("ban-code-comments", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.Var(&languageValues, "language", "language to scan; repeat or use comma-separated values")
	flags.Var(&languageValues, "languages", "languages to scan; repeat or use comma-separated values")
	flags.Var(&categoryValues, "categories", "comment categories to report")
	flags.Var(&includes, "include", "file glob to include")
	flags.Var(&excludes, "exclude", "file glob to exclude")
	flags.StringVar(&format, "format", string(FormatJSON), "output format: json or text")
	flags.BoolVar(&debug, "debug", false, "write skipped-file diagnostics to stderr")
	if err := flags.Parse(args); err != nil {
		return Options{}, err
	}

	selectedLanguages, err := languages.ParseSelection(languageValues)
	if err != nil {
		return Options{}, err
	}
	selectedCategories, err := parseCategories(categoryValues)
	if err != nil {
		return Options{}, err
	}
	outputFormat := Format(strings.ToLower(format))
	if outputFormat != FormatJSON && outputFormat != FormatText {
		return Options{}, fmt.Errorf("unsupported format %q", format)
	}
	paths := flags.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}
	return Options{Paths: paths, Languages: selectedLanguages, Categories: selectedCategories, Includes: includes, Excludes: excludes, Format: outputFormat, Debug: debug}, nil
}

func parseCategories(values []string) (map[model.Category]bool, error) {
	selected := make(map[model.Category]bool)
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			category := model.Category(strings.ToLower(strings.TrimSpace(item)))
			if category == "" {
				continue
			}
			switch category {
			case model.CategoryOrdinary, model.CategoryDocumentation, model.CategoryHeader, model.CategoryDirective:
				selected[category] = true
			default:
				return nil, fmt.Errorf("unsupported category %q", item)
			}
		}
	}
	if len(selected) == 0 {
		selected[model.CategoryOrdinary] = true
		selected[model.CategoryDocumentation] = true
	}
	return selected, nil
}
