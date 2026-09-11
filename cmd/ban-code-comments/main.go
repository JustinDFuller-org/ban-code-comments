package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/JustinDFuller/ban-code-comments/internal/cli"
	"github.com/JustinDFuller/ban-code-comments/internal/discovery"
	"github.com/JustinDFuller/ban-code-comments/internal/model"
	"github.com/JustinDFuller/ban-code-comments/internal/scanner"
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
	findings := make([]model.Finding, 0)
	for _, candidate := range discovered.Candidates {
		source, err := os.ReadFile(candidate.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", candidate.RelPath, err)
			os.Exit(model.ExitCode(nil, err))
		}
		findings = append(findings, scanner.Scan(candidate.RelPath, candidate.Language, source, options.Categories)...)
	}
	result := model.Result{Findings: findings, Summary: model.Summary{FilesScanned: len(discovered.Candidates), FilesSkipped: discovered.Skipped, Findings: len(findings)}}
	if options.Format == cli.FormatText {
		if len(result.Findings) == 0 {
			os.Exit(0)
		}
		for _, finding := range result.Findings {
			fmt.Printf("%s:%d:%d: %s\n", finding.Path, finding.Range.Start.Line, finding.Range.Start.Column, finding.Text)
		}
		os.Exit(model.ExitCode(findings, nil))
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(model.ExitCode(nil, err))
	}
	os.Exit(model.ExitCode(findings, nil))
}
