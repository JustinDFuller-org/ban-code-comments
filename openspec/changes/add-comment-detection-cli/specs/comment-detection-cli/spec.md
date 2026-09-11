## Purpose

Provide a portable command-line policy check that finds code comments in common source and configuration files while avoiding false positives from comment-shaped text in executable content.

## ADDED Requirements

### Requirement: Scan supported source and configuration languages
The CLI SHALL recursively scan supported files beneath supplied paths, or the current directory when no path is supplied. It SHALL support Go; JavaScript, TypeScript, JSX, and TSX; Python; Rust; Java; C, C++, and C#; Kotlin; Swift; Ruby; PHP; shell; SQL; HTML and XML; CSS and SCSS; YAML; TOML; JSONC; HCL and Terraform; Dockerfile; Makefile; and INI. It SHALL report only syntactically recognized comments and SHALL NOT report comment delimiters contained in strings, raw strings, templates, or escaped literal content.

#### Scenario: Source comment is detected
- **WHEN** a supported scanned file contains an ordinary line or block comment outside literal content
- **THEN** the CLI produces one finding for that comment

#### Scenario: Comment-shaped literal content is ignored
- **WHEN** a supported scanned file contains a comment delimiter inside a string, raw string, template, or escaped literal
- **THEN** the CLI produces no finding for that delimiter

### Requirement: Apply practical comment categories
The CLI SHALL classify recognized comments as ordinary, documentation, header, or directive. By default it SHALL report ordinary and documentation comments and SHALL not report header or directive comments. The `--categories` option SHALL let users select the categories evaluated as findings.

#### Scenario: Default policy preserves directive comments
- **WHEN** a scanned file contains a recognized directive comment and the user supplies no category selection
- **THEN** the directive is not a finding

#### Scenario: Category selection includes a directive
- **WHEN** a user selects the directive category and a scanned file contains a recognized directive comment
- **THEN** the directive is a finding

### Requirement: Select files and languages from the command line
The CLI SHALL accept command-line options to select supported languages and to include or exclude files by glob. It SHALL use repository-aware discovery that honors `.gitignore`, excludes ordinary VCS, dependency, and build directories, and silently ignores unsupported or irrelevant files. Exclude globs SHALL take precedence over include globs. The `--debug` option SHALL emit skipped-path diagnostics to stderr without changing normal result output.

#### Scenario: Ignored files are not scanned
- **WHEN** a candidate file is matched by `.gitignore` or an exclude glob
- **THEN** the file is not scanned or reported as a finding

#### Scenario: Debug mode explains an ignored file
- **WHEN** `--debug` is enabled and a candidate file is skipped because it is unsupported or ignored
- **THEN** the CLI writes the path and skip reason to stderr

### Requirement: Produce machine-readable findings and enforcement status
The CLI SHALL write JSON to stdout by default containing findings and a scan summary. Each finding SHALL include its path, source range, detected language, category, and comment text. It SHALL support a human-readable text format through `--format text`. It SHALL exit with status 0 when no findings exist, status 1 when one or more findings exist, and status 2 for invalid options or scan failures.

#### Scenario: Clean scan succeeds
- **WHEN** all scanned files have no selected comment categories
- **THEN** the JSON result contains an empty findings collection and the process exits with status 0

#### Scenario: Policy violation fails CI
- **WHEN** at least one selected comment category is found
- **THEN** the JSON result includes the findings and the process exits with status 1

#### Scenario: Invalid invocation is distinguishable
- **WHEN** the user supplies an invalid option value or a requested path cannot be scanned
- **THEN** the CLI reports an invocation or scan error and exits with status 2

### Requirement: Distribute verified standalone binaries
The project SHALL publish version-tagged, checksummed standalone binaries for macOS arm64 and amd64, Linux arm64 and amd64, and Windows amd64 through GitHub Releases.

#### Scenario: Tagged release publishes supported artifacts
- **WHEN** a version tag is released from the project
- **THEN** GitHub Releases contains an archive and checksum information for each supported platform target
