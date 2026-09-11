## Context

The repository currently has only a product README and OpenSpec bootstrap. The proposed CLI needs broad multi-language detection, CI-safe exit semantics, repository-aware traversal, and cross-platform distribution without requiring users to install another runtime or native parser toolchain.

## Goals / Non-Goals

**Goals:**

- Deliver a single Go binary with deterministic, scriptable output and exit behavior.
- Detect actual comments across the supported language set while excluding delimiters in literal content.
- Make practical policy selection and file/language scope entirely flag-driven.
- Produce tagged release artifacts for the selected desktop and CI platforms.

**Non-Goals:**

- A checked-in configuration format, generic heuristics for unknown files, text-regex exemptions, editor integration, a hosted service, a packaged GitHub Action, or a packaged Claude hook.
- Full syntax parsing, automatic fixing, comment rewriting, or support for languages outside the documented initial set.

## Decisions

### Use Go with pure-Go lexical scanner families

The command will be a Go module with a small CLI layer, discovery layer, language registry, comment classifier, scanner-family implementations, report encoder, and release configuration. Scanner families will model lexical states required for their syntax instead of matching delimiters with regular expressions: C-style line/block comments, hash comments, shell heredocs, Python-style triples, HTML comments, SQL comments, and format-specific forms. The language registry will map extensions and special filenames to a scanner and category classifier. Tree-sitter is not used because broad CGo-backed grammar packaging would undermine simple native cross-compilation; simple regular-expression scanning is rejected because it cannot reliably distinguish strings and templates from comments.

### Make category policy explicit and grammar-owned

Each recognized comment receives one of `ordinary`, `documentation`, `header`, or `directive`. The effective category set defaults to ordinary and documentation; `--categories` replaces that set with a comma-separated selected set. A language-specific classifier identifies semantically required directives and leading header/generated markers. This enables strict policies without unreviewable comment-text exemptions.

### Make file selection repository-aware but overrideable

Paths are traversed recursively, using `.gitignore` semantics and a fixed exclusion set for VCS, dependency, and build directories. `--language`, `--include`, and `--exclude` are repeatable or comma-separated filters; include globs narrow candidates and exclude globs always win. Files without a registry mapping are not examined. Debug diagnostics go only to stderr so stdout remains valid JSON.

### Define a stable reporting boundary

The default stdout value will be a JSON object with `findings` and `summary`; every finding carries path, start/end line and column, language, category, and text. Text mode derives from the same finding model. Operational messages and debug diagnostics use stderr. The process status distinguishes a clean result, a policy failure, and a command failure.

### Release from version tags

GoReleaser will build archives for macOS arm64/amd64, Linux arm64/amd64, and Windows amd64, generate checksums, and publish them to GitHub Releases from a tag-triggered GitHub Actions workflow. This keeps installation to a downloadable archive and deliberately defers Homebrew, Scoop, and other package-manager maintenance.

## Risks / Trade-offs

- [Lexical scanners do not provide a full AST] → Define scanner state machines and per-language fixtures for literals, templates, nesting, and special comment forms before claiming support.
- [Broad language coverage can drift] → Keep a registry-owned support matrix and require fixtures for every extension and category rule.
- [Ignore behavior may hide an unexpectedly unsupported file] → Preserve silent normal scans but make every skip observable through `--debug` and document the language list.
- [Release automation is a supply-chain boundary] → Limit permissions, pin actions, generate checksums, and validate release configuration in CI.

## Migration Plan

1. Add the CLI and fixture suite without changing existing repository behavior.
2. Add documentation examples for direct invocation and external automation.
3. Validate GoReleaser configuration and publish the first version-tagged GitHub Release.
4. Roll back a faulty release by publishing a corrected patch version; no persistent data or repository migration is introduced.
