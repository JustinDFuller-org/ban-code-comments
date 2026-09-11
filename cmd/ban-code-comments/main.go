package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/JustinDFuller/ban-code-comments/internal/cli"
	"github.com/JustinDFuller/ban-code-comments/internal/discovery"
	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

func main() {
	options, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(model.ExitCode(nil, err))
	}
	discovered, err := discovery.Discover(discovery.Config{Paths: options.Paths, Languages: options.Languages, Includes: options.Includes, Excludes: options.Excludes, Debug: options.Debug})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(model.ExitCode(nil, err))
	}
	for _, diagnostic := range discovered.Debug {
		fmt.Fprintf(os.Stderr, "debug: %s: %s\n", diagnostic.Path, diagnostic.Reason)
	}
	result := model.Result{Findings: []model.Finding{}, Summary: model.Summary{FilesScanned: len(discovered.Candidates), FilesSkipped: discovered.Skipped}}
	if options.Format == cli.FormatText {
		if len(result.Findings) == 0 {
			return
		}
		for _, finding := range result.Findings {
			fmt.Printf("%s:%d:%d: %s\n", finding.Path, finding.Range.Start.Line, finding.Range.Start.Column, finding.Text)
		}
		return
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(model.ExitCode(nil, err))
	}
}
