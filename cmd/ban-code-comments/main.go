package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/JustinDFuller/ban-code-comments/internal/cli"
	"github.com/JustinDFuller/ban-code-comments/internal/discovery"
	"github.com/JustinDFuller/ban-code-comments/internal/hook"
	"github.com/JustinDFuller/ban-code-comments/internal/model"
	"github.com/JustinDFuller/ban-code-comments/internal/report"
	"github.com/JustinDFuller/ban-code-comments/internal/scanner"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "hook" || args[0] == "--hook") {
		runHook(args[1:])
		return
	}
	options, err := cli.Parse(args)
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
		if err := report.Text(os.Stdout, result.Findings); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(model.ExitCode(nil, err))
		}
		os.Exit(model.ExitCode(findings, nil))
	}
	if err := report.JSON(os.Stdout, result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(model.ExitCode(nil, err))
	}
	os.Exit(model.ExitCode(findings, nil))
}

func runHook(args []string) {
	mode := hook.ModeHard
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--mode":
			if index+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "missing value for --mode")
				os.Exit(2)
			}
			parsed, err := hook.ParseMode(args[index+1])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			mode = parsed
			index++
		default:
			if strings.HasPrefix(args[index], "-") {
				fmt.Fprintf(os.Stderr, "unsupported hook option %q\n", args[index])
				os.Exit(2)
			}
		}
	}
	if err := hook.Run(os.Stdin, os.Stdout, mode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
