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

Use `@v1` to receive compatible releases automatically. For reproducible workflows, pin an exact release such as `@v1.0.0` or pin the Action to a full commit SHA. The moving `v1` tag is maintained to the latest compatible release.

The Action accepts the same configuration as the CLI:

| Input | Default | Description |
| --- | --- | --- |
| `paths` | `.` | Files or directories to scan; separate multiple values with newlines. |
| `languages` | — | Languages to scan; separate values with commas or newlines. |
| `categories` | — | Comment categories to report; separate values with commas. |
| `include` | — | File globs to include; separate multiple values with newlines. |
| `exclude` | — | File globs to exclude; separate multiple values with newlines. |
| `format` | `json` | Output format: `json` or `text`. |
| `debug` | `false` | Set to `true` to write skipped-file diagnostics to stderr. |

The Action exits `0` for a clean scan, `1` when selected findings exist, and `2` for invalid options or scan failures. Findings and operational errors therefore fail the workflow step while remaining distinguishable in the log.

For direct automation outside GitHub Actions, install the released binary and run `ban-code-comments .` as a step. This repository dogfoods the current source tree in CI with `go run ./cmd/ban-code-comments --format text --exclude 'internal/scanner/testdata/fixtures/**' --exclude 'internal/hook/testdata/**' --exclude 'plugins/**' .`; the exclusions keep intentional fixtures and generated plugin bundles separate from production code. Claude hooks can invoke the same command against the changed repository or file paths.

## Codex plugins

This repository provides two separately installable Codex plugins: `ban-code-comments-warn` allows supported edits and adds guidance, while `ban-code-comments-hard-block` denies reconstructible supported edits that introduce findings and stops after opaque Bash findings.

Review plugin source and hook commands before trusting them, then add the repository marketplace and install one variant:

```sh
codex plugin marketplace add JustinDFuller/ban-code-comments
codex plugin add ban-code-comments-warn@ban-code-comments
```

Use `ban-code-comments-hard-block@ban-code-comments` instead when denial is preferred, and review the enabled hook in `/hooks` after installation. The shared scanner covers the supported source and configuration languages in the language registry; Markdown, README files, literals, directives, and unsupported paths are not code-comment findings.

Pre-tool evaluation reconstructs documented file-edit payloads and compares proposed findings with the file's existing findings, so unchanged legacy comments do not block unrelated edits. Opaque Bash commands are audited after execution, which reports new findings but cannot prevent the command from writing first.

The guidance skill directs agents to use Git history, pull-request descriptions, simplified code, nearby README files, and Markdown instead of explanatory source comments. Roll back by disabling or removing the plugin with `codex plugin remove <plugin>@ban-code-comments`; the existing CLI, GitHub Action, and repository files are unchanged.

## Development

Run the complete local checks with:

```sh
gofmt -w .
go test ./...
go vet ./...
npm ci
npm test
npm run coverage
```

## Coverage policy

CI measures all Go production packages and JavaScript files under `src/` using the existing test suites, converts both reports to Cobertura XML, and publishes separate `code-coverage/go` and `code-coverage/javascript` reports to GitHub Code Quality.

Run the reports locally with `go test ./... -coverprofile=coverage/go.out && go run github.com/boumenot/gocover-cobertura@v1.5.0 < coverage/go.out > coverage/go-cobertura.xml` and `npm run coverage`; the Go report must remain at or above 90 percent aggregate line coverage before the repository rule is activated.

GitHub Code Quality must be enabled for the repository before coverage reports can appear on pull requests or a `Restrict code coverage` ruleset can be evaluated.

The repository policy uses one aggregate line-coverage threshold across Go and JavaScript, with the active minimum set to the greater of 90 percent and the default-branch baseline and a maximum allowed drop of one percentage point.

Fork pull requests still run the test and report-generation job, but the privileged Code Quality upload is skipped because fork workflows cannot receive `code-quality: write`; skipped publication is reported as unavailable rather than successful.

To roll back enforcement, change the default-branch ruleset to Evaluate or disable its `Restrict code coverage` rule; report generation and pull-request visibility can remain enabled independently.

The isolated Codex smoke harness reports `unsupported` unless given a built hook CLI and explicit authenticated-test opt-in: `BAN_CODE_COMMENTS_EXECUTABLE=/path/to/ban-code-comments CODEX_SMOKE_REUSE_AUTH=1 npm run smoke:codex`. It uses a temporary Codex home and repository and reports live hook cases separately from unsupported coverage.

Semantic-version tags matching `vMAJOR.MINOR.PATCH` publish cross-platform archives and checksums through GoReleaser.
