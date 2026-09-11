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

## Automation

In GitHub Actions, install the released binary and run `ban-code-comments .` as a step. Preserve its exit status so findings fail the job. Claude hooks can invoke the same command against the changed repository or file paths.

## Development

Run the complete local checks with:

```sh
gofmt -w .
go test ./...
go vet ./...
```

Semantic-version tags matching `vMAJOR.MINOR.PATCH` publish cross-platform archives and checksums through GoReleaser.
