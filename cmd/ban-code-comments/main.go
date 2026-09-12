package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/JustinDFuller-org/ban-code-comments/internal/cli"
	"github.com/JustinDFuller-org/ban-code-comments/internal/discovery"
	"github.com/JustinDFuller-org/ban-code-comments/internal/hook"
	"github.com/JustinDFuller-org/ban-code-comments/internal/model"
	"github.com/JustinDFuller-org/ban-code-comments/internal/report"
	"github.com/JustinDFuller-org/ban-code-comments/internal/scanner"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, input io.Reader, output, errorOutput io.Writer) int {
	if len(args) > 0 && (args[0] == "hook" || args[0] == "--hook") {
		return runHook(args[1:], input, output, errorOutput)
	}
	options, err := cli.Parse(args)
	if err != nil {
		fmt.Fprintln(errorOutput, err)
		return model.ExitCode(nil, err)
	}
	discovered, err := discovery.Discover(discovery.Config{Paths: options.Paths, Languages: options.Languages, Includes: options.Includes, Excludes: options.Excludes, Debug: options.Debug})
	if err != nil {
		fmt.Fprintln(errorOutput, err)
		return model.ExitCode(nil, err)
	}
	for _, diagnostic := range discovered.Debug {
		fmt.Fprintf(errorOutput, "debug: %s: %s\n", diagnostic.Path, diagnostic.Reason)
	}
	findings := make([]model.Finding, 0)
	for _, candidate := range discovered.Candidates {
		source, err := os.ReadFile(candidate.Path)
		if err != nil {
			fmt.Fprintf(errorOutput, "%s: %v\n", candidate.RelPath, err)
			return model.ExitCode(nil, err)
		}
		findings = append(findings, scanner.Scan(candidate.RelPath, candidate.Language, source, options.Categories)...)
	}
	result := model.Result{Findings: findings, Summary: model.Summary{FilesScanned: len(discovered.Candidates), FilesSkipped: discovered.Skipped, Findings: len(findings)}}
	if options.Format == cli.FormatText {
		if err := report.Text(output, result.Findings); err != nil {
			fmt.Fprintln(errorOutput, err)
			return model.ExitCode(nil, err)
		}
		return model.ExitCode(findings, nil)
	}
	if err := report.JSON(output, result); err != nil {
		fmt.Fprintln(errorOutput, err)
		return model.ExitCode(nil, err)
	}
	return model.ExitCode(findings, nil)
}

func runHook(args []string, input io.Reader, output, errorOutput io.Writer) int {
	mode := hook.ModeHard
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--mode":
			if index+1 >= len(args) {
				fmt.Fprintln(errorOutput, "missing value for --mode")
				return 2
			}
			parsed, err := hook.ParseMode(args[index+1])
			if err != nil {
				fmt.Fprintln(errorOutput, err)
				return 2
			}
			mode = parsed
			index++
		case "--state-dir":
			if index+1 >= len(args) {
				fmt.Fprintln(errorOutput, "missing value for --state-dir")
				return 2
			}
			index++
		default:
			if strings.HasPrefix(args[index], "-") {
				fmt.Fprintf(errorOutput, "unsupported hook option %q\n", args[index])
				return 2
			}
		}
	}
	if err := hook.Run(input, output, mode); err != nil {
		fmt.Fprintln(errorOutput, err)
		return 2
	}
	return 0
}
