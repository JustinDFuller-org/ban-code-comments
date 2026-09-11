# ban-code-comments

`ban-code-comments` detects code comments so teams can enforce a no-comments policy locally and in CI. It is a standalone Go binary with no language runtime or repository configuration file required.

## Install

Download the archive for your platform from the [GitHub Releases](https://github.com/JustinDFuller/ban-code-comments/releases) page, verify it with `checksums.txt`, and place `ban-code-comments` on your `PATH`. Releases include macOS arm64/amd64, Linux arm64/amd64, and Windows amd64 binaries.

For local development, run `go install github.com/JustinDFuller/ban-code-comments/cmd/ban-code-comments@latest`.

## Usage

Run `ban-code-comments` to scan the current repository, or pass one or more files or directories:

```sh
ban-code-comments .
ban-code-comments src scripts
```

The default output is JSON with `findings` and `summary` fields. Use `--format text` for concise path, line, column, and comment output.

```sh
ban-code-comments --format text .
ban-code-comments --language go,python --include 'src/**' --exclude 'src/vendor/**' .
```

The practical default reports ordinary and documentation comments. Select `ordinary`, `documentation`, `header`, or `directive` categories with `--categories`; selecting categories replaces the default set. Repeat or comma-separate `--language`, `--include`, and `--exclude` values. `--debug` writes reasons for skipped files to stderr.

The process exits with `0` when no selected comments are found, `1` when findings exist, and `2` for invalid options, unreadable paths, or scan failures.

## Supported languages

Go; JavaScript, TypeScript, JSX, and TSX; Python; Rust; Java; C, C++, and C#; Kotlin; Swift; Ruby; PHP; shell; SQL; HTML and XML; CSS and SCSS; YAML; TOML; JSONC; HCL and Terraform; Dockerfile; Makefile; and INI. Standard JSON files are recognized and scanned as comment-free JSON.

The scanner ignores comment-shaped text inside strings, raw strings, templates, escaped literals, heredocs, and triple-quoted literals. Unsupported or irrelevant files are skipped; use `--debug` to inspect those decisions.

## GitHub Actions

The Action requires a checkout step and runs with read-only repository access. It downloads the exact CLI version coupled to the Action release, verifies the published checksum, caches the runner-specific binary, and preserves the CLI's report and exit status.

```yaml
name: Ban code comments

on: [push, pull_request]

permissions:
  contents: read

jobs:
  comments:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: JustinDFuller/ban-code-comments@v1
        with:
          format: text
          languages: go,python,typescript
          exclude: |
            vendor/**
            internal/scanner/testdata/fixtures/**
```

Use an exact release such as `@v1.0.0` or a full commit SHA when reproducibility is more important than automatically receiving compatible updates. The moving major tag is maintained to the latest compatible release.

Action inputs are `paths` (newline-separated, default `.`), `languages` (comma- or newline-separated), `categories` (comma-separated), `include` and `exclude` (newline-separated globs), `format` (`json` or `text`, default `json`), and `debug` (`true` or `false`, default `false`). Their behavior matches the direct CLI options.

The Action exits `0` for a clean scan, `1` when selected findings exist, and `2` for invalid options or scan failures. Findings and operational errors therefore fail the workflow step while remaining distinguishable in the log.

For direct automation outside GitHub Actions, install the released binary and run `ban-code-comments .` as a step. This repository dogfoods the current source tree in CI with `go run ./cmd/ban-code-comments --format text --exclude 'internal/scanner/testdata/fixtures/**' .`; the fixture exclusion keeps intentional scanner inputs separate from production code. Claude hooks can invoke the same command against the changed repository or file paths.

## Development

Run the complete local checks with:

```sh
gofmt -w .
go test ./...
go vet ./...
```

Semantic-version tags matching `vMAJOR.MINOR.PATCH` publish cross-platform archives and checksums through GoReleaser.
