# ban-code-comments

`ban-code-comments` detects code comments so teams can enforce a no-comments policy locally and in CI. It is a self-contained Node.js package with no repository configuration file required.

## Install

Install the published package with Node.js 24 or newer:

```sh
npm install --global @justindfuller/ban-code-comments
```

The package includes the `ban-code-comments` executable and the programmatic API. It does not require Go, a native executable, or a runtime download.

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

The Action requires a checkout step and runs with read-only repository access. Its release contains the matching JavaScript implementation, so it does not download or cache a native CLI at runtime.

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
      - uses: JustinDFuller-org/ban-code-comments@v1
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

For direct automation outside GitHub Actions, install the npm package and run `ban-code-comments .` as a step. The CLI and Action use the same implementation, options, reports, and exit statuses.

The package also exposes plain-data operations:

```js
import { check, scanSource } from "@justindfuller/ban-code-comments";

const sourceFindings = scanSource("const value = 1; // finding\n", "fixture.js", "javascript");
const repositoryResult = await check(["."], { categories: new Set(["ordinary", "documentation"]) });
```

The API exports `scanSource`, `discover`, `check`, `runCLI`, `evaluateHook`, `renderJSON`, `renderText`, `lookup`, and the shared model and category helpers.

## Codex plugins

This repository provides two separately installable Codex plugins: `ban-code-comments-warn` allows supported traditional file edits and adds guidance, while `ban-code-comments-hard-block` denies supported traditional file edits that introduce findings.

Review plugin source and hook commands before trusting them, then add the repository marketplace and install one variant:

```sh
codex plugin marketplace add JustinDFuller-org/ban-code-comments
codex plugin add ban-code-comments-warn@ban-code-comments
```

Use `ban-code-comments-hard-block@ban-code-comments` instead when denial is preferred, and review the enabled hook in `/hooks` after installation. The shared scanner covers the supported source and configuration languages in the language registry; Markdown, README files, literals, directives, and unsupported paths are not code-comment findings.

Pre-tool evaluation reconstructs documented file-edit payloads from `apply_patch`, `edit`, `write`, `write_file`, and `file_write`, then compares proposed findings with the file's existing findings, so unchanged legacy comments do not block unrelated edits. Bash, shell, exec, MCP, generators, redirection, and other opaque write paths are outside the plugin boundary; the repository scanner and GitHub Action enforce the final state in CI.

The guidance skill directs agents to use Git history, pull-request descriptions, simplified code, nearby README files, and Markdown instead of explanatory source comments. The plugin bundles are self-contained and work offline after installation. Roll back by disabling or removing the plugin with `codex plugin remove <plugin>@ban-code-comments`.

## Claude Code plugins

This repository provides two separately installable Claude Code plugins: `ban-code-comments-claude-warn` allows supported `Edit` and `Write` calls and adds model-visible guidance, while `ban-code-comments-claude-hard-block` denies supported calls that introduce findings.

Review plugin source and hook commands before trusting them, then add the repository marketplace and install one variant:

```sh
claude plugin marketplace add JustinDFuller-org/ban-code-comments
claude plugin install ban-code-comments-claude-warn@ban-code-comments
claude plugin enable ban-code-comments-claude-warn@ban-code-comments
```

Use `ban-code-comments-claude-hard-block@ban-code-comments` instead when denial is preferred. Roll back with `claude plugin disable <plugin>@ban-code-comments` or `claude plugin uninstall <plugin>@ban-code-comments`. The plugins require Node.js 24 or newer for their bundled launchers.

Claude pre-tool evaluation covers only `PreToolUse` `Edit` and `Write` events. Bash, PowerShell, MCP, NotebookEdit, generators, redirection, and other opaque writes remain outside the plugin boundary; the repository scanner and GitHub Action enforce their final state in CI. Markdown, README files, literals, directives, and unsupported paths remain outside the comment finding policy.

## Development

Run the complete local checks with Node.js 24 or newer:

```sh
npm ci
npm test
npm run coverage
npm run validate:plugins
npm run build
npm run build:plugins
```

## Coverage policy

CI runs the Node test suite, enforces at least 90 percent aggregate JavaScript line coverage, validates both Codex and Claude Code plugin variants, rebuilds generated distributions, runs the package CLI, and validates OpenSpec.

## Releases

Releases are manual administrator actions. To publish a version, an administrator creates the protected `vMAJOR.MINOR.PATCH` tag on the current `main` commit, selects that tag when manually dispatching the `Release` workflow, and waits for the `npm-publish` environment approval from `JustinDFuller`. The workflow validates the tag, source commit, and package version before publishing through the repository's npm trusted publisher with GitHub Actions OIDC.

After npm publication succeeds, an administrator manually advances the protected floating Action tag, such as `v1`, to the released commit. The release workflow never creates or updates floating tags. Use an exact version such as `@v1.0.0` or a full commit SHA when reproducibility is required.
