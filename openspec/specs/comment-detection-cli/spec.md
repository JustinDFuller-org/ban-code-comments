# comment-detection-cli Specification

## Purpose

Provide a publishable Node.js command-line package and programmatic API that detects code comments consistently across supported source and configuration languages.

## Requirements

### Requirement: Provide a Node.js CLI and programmatic API

The package SHALL target Node.js 24 or newer, publish under the `@justindfuller/ban-code-comments` npm package name, and provide a `ban-code-comments` executable. It SHALL expose programmatic operations for source scanning, repository discovery, complete checks, result reporting, and Codex hook evaluation. The CLI and programmatic operations SHALL use the same behavior and result model.

#### Scenario: Installed package exposes the CLI

- **WHEN** a user installs the published package on Node.js 24 or newer
- **THEN** the `ban-code-comments` executable is available and scans the requested paths

#### Scenario: Programmatic and CLI checks agree

- **WHEN** the same repository and options are supplied through the programmatic check and CLI entrypoint
- **THEN** both produce equivalent findings, summary values, reports, and enforcement status

### Requirement: Scan the supported language set with literal-safe detection

The package SHALL support Go; JavaScript, TypeScript, JSX, and TSX; Python; Rust; Java; C, C++, and C#; Kotlin; Swift; Ruby; PHP; shell; SQL; HTML and XML; CSS and SCSS; YAML; TOML; JSONC; HCL and Terraform; Dockerfile; Makefile; and INI. It SHALL recognize the same extensions, special filenames, aliases, comment categories, and lexical comment rules as the existing CLI. It SHALL report recognized comments outside executable literals and SHALL ignore comment-shaped text inside strings, raw strings, templates, escaped literals, heredocs, and triple-quoted literals.

#### Scenario: Supported source comment is detected

- **WHEN** a supported file contains an ordinary or documentation comment outside literal content
- **THEN** the package returns one finding with path, language, category, source range, and comment text

#### Scenario: Literal comment text is ignored

- **WHEN** a supported file contains comment delimiters inside a recognized literal or heredoc form
- **THEN** the package returns no finding for those delimiters

### Requirement: Apply category selection and defaults

The package SHALL classify recognized comments as ordinary, documentation, header, or directive. It SHALL report ordinary and documentation categories by default. A category selection SHALL replace the default selection and SHALL accept the same comma-separated and repeated command-line forms as the existing CLI.

#### Scenario: Default categories exclude headers and directives

- **WHEN** a scan contains ordinary, documentation, header, and directive comments without an explicit category selection
- **THEN** only ordinary and documentation findings are returned

#### Scenario: Explicit categories replace defaults

- **WHEN** a user selects the directive category
- **THEN** directive comments are returned and ordinary/documentation comments are not returned unless explicitly selected

### Requirement: Discover files with repository-aware selection

The package SHALL recursively scan supplied files or directories, defaulting to the current directory. It SHALL honor `.gitignore`, fixed VCS/dependency/build exclusions, language filters, include globs, and exclude globs. Exclude globs SHALL take precedence over include globs. Unsupported and irrelevant files SHALL be skipped, and debug mode SHALL report skipped paths and reasons to stderr without changing normal results.

#### Scenario: Selection filters determine candidates

- **WHEN** a repository contains supported, unsupported, ignored, included, and excluded paths
- **THEN** only supported paths surviving all selection rules are scanned

#### Scenario: Debug exposes skip decisions

- **WHEN** debug mode is enabled and a path is skipped
- **THEN** stderr contains the path and skip reason while stdout remains the selected report

### Requirement: Produce stable reports and enforcement statuses

The CLI SHALL write JSON to stdout by default with `findings` and `summary` fields, and SHALL support the existing human-readable text format. JSON findings SHALL include `path`, `language`, `category`, `range`, and `text`; summaries SHALL include `files_scanned`, `files_skipped`, and `findings`. The process SHALL exit with status 0 for no findings, 1 for selected findings, and 2 for invalid options or scan failures.

#### Scenario: Clean invocation succeeds

- **WHEN** no selected comments are found
- **THEN** the CLI writes an empty findings collection and exits with status 0

#### Scenario: Finding invocation fails by policy

- **WHEN** one or more selected comments are found
- **THEN** the CLI writes all findings and exits with status 1

#### Scenario: Invalid or unreadable invocation fails operationally

- **WHEN** an option is invalid or a requested path cannot be scanned
- **THEN** the CLI reports an operational error and exits with status 2

### Requirement: Publish a self-contained JavaScript package

The package SHALL be distributable through the public npm registry without requiring Go, a standalone native executable, or a runtime download of a platform-specific release artifact. The package release SHALL include the executable, programmatic modules, and documentation for supported options, outputs, statuses, and integrations.

#### Scenario: Consumer installs without Go

- **WHEN** a consumer installs the published npm package on a supported Node.js 24 runtime
- **THEN** the consumer can run the CLI without installing Go or downloading a separate CLI binary

#### Scenario: Package release is reproducible

- **WHEN** a release version is published
- **THEN** its npm package metadata, executable entrypoint, and bundled integration artifacts identify the same version
