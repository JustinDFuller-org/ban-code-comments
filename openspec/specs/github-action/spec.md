# github-action Specification

## Purpose

Provide a versioned GitHub Action that lets repositories run the bundled JavaScript `ban-code-comments` implementation without installing Go or downloading a platform-specific binary.

## Requirements

### Requirement: Run the matching released CLI on supported runners

The Action SHALL run the JavaScript `ban-code-comments` implementation bundled with the Action release on Linux amd64 and arm64, macOS amd64 and arm64, and Windows amd64 runners supported by Node.js 24. It SHALL reject an unsupported runner before scanning. It SHALL not download, verify, cache, or execute a separate Go binary or platform archive.

#### Scenario: Supported runner executes the matching release

- **WHEN** a checked-out repository invokes a released Action on a supported runner
- **THEN** the Action runs its bundled JavaScript implementation against the requested workspace paths

#### Scenario: Unsupported runner is rejected

- **WHEN** the Action runs on an operating-system and architecture combination unsupported by Node.js 24 or the package
- **THEN** the Action reports an actionable operational error and exits unsuccessfully without attempting a scan

#### Scenario: Bundled distribution integrity fails

- **WHEN** the bundled JavaScript Action distribution is incomplete or fails package integrity validation
- **THEN** the Action rejects the distribution, reports an operational error, and does not execute a scan

### Requirement: Expose the CLI configuration through Action inputs

The Action SHALL expose inputs corresponding to the CLI's paths, languages, categories, include globs, exclude globs, output format, and debug mode. Omitted inputs SHALL retain the CLI defaults: the workspace path, JSON output, ordinary and documentation categories, no language/include/exclude filter, and debug disabled. Paths SHALL accept newline-separated entries, and list-like filter inputs SHALL accept the comma-separated forms supported by the CLI; include and exclude inputs SHALL also accept newline-separated entries.

#### Scenario: Default inputs match direct CLI behavior

- **WHEN** a repository invokes the Action without configuration inputs
- **THEN** the Action runs the CLI against the workspace with JSON output, the default comment categories, and no additional filters

#### Scenario: Configured filters reach the CLI

- **WHEN** a repository supplies paths, languages, categories, include globs, exclude globs, format, or debug inputs
- **THEN** the Action passes the equivalent values to the CLI without shell expansion or loss of spaces and glob characters

#### Scenario: Invalid configuration is rejected by the CLI contract

- **WHEN** an input contains an unsupported language, category, or output format
- **THEN** the Action reports the CLI validation error and exits with the CLI's operational failure status

### Requirement: Preserve CLI reports and enforcement statuses

The Action SHALL stream the CLI's standard output and standard error to the workflow log and SHALL preserve its exit status: `0` for a clean scan, `1` when selected findings exist, and `2` for invalid options or scan failures. The Action SHALL not add a separate configuration file, finding policy, or alternate report format.

#### Scenario: Clean scan passes

- **WHEN** the CLI finds no selected comments
- **THEN** the Action emits the selected report and completes successfully with exit status `0`

#### Scenario: Findings fail the workflow step

- **WHEN** the CLI reports one or more selected findings
- **THEN** the Action emits the selected report and completes unsuccessfully with exit status `1`

#### Scenario: Scan failure remains distinguishable

- **WHEN** the CLI cannot parse the inputs, read a requested path, or complete the scan
- **THEN** the Action emits the error on standard error and exits with status `2`

### Requirement: Provide a permission-minimal and documented integration

The Action SHALL require no GitHub token, repository write permission, or runtime network access. Documentation SHALL show the required checkout step, Action invocation with a release reference, every supported input, Node.js 24 runner requirements, exit behavior, npm package usage, and guidance for immutable commit-SHA pinning.

#### Scenario: Checked-out repository is scanned without write access

- **WHEN** a workflow grants only `contents: read`, checks out the repository, and invokes the Action
- **THEN** the Action scans the checked-out files without requiring additional permissions or mutating the repository

#### Scenario: Consumer can reproduce a release

- **WHEN** a consumer invokes an exact Action release reference
- **THEN** the Action uses the bundled JavaScript implementation coupled to that release and does not resolve an unrelated runtime version
