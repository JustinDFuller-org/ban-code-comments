## Purpose

Provide installable Claude Code plugins that enforce the repository's no-code-comments policy before supported file edits are applied, while preserving the existing scanner and CI enforcement boundary.

## ADDED Requirements

### Requirement: Enforce the comment policy before supported Claude edits

The plugins SHALL inspect Claude Code `PreToolUse` events for the `Edit` and `Write` tools. They SHALL reconstruct the proposed file content, evaluate it with the shared JavaScript language registry, scanner, and default comment categories, and compare proposed findings with findings already present in the file. A hard-block plugin SHALL deny a proposed edit that introduces a finding before mutation. A warn plugin SHALL allow the edit and provide concise model-visible guidance.

#### Scenario: Hard mode denies a newly introduced comment through Write

- **WHEN** Claude proposes a `Write` call whose supported source content introduces an ordinary or documentation comment
- **THEN** the plugin returns a Claude `PreToolUse` deny decision with the affected path and finding details, and the write does not occur

#### Scenario: Hard mode denies a newly introduced comment through Edit

- **WHEN** Claude proposes an `Edit` call whose replacement introduces an ordinary or documentation comment
- **THEN** the plugin returns a Claude `PreToolUse` deny decision with the affected path and finding details, and the edit does not occur

#### Scenario: Warn mode allows the same proposed edit

- **WHEN** the warn plugin receives either supported edit with a newly introduced finding
- **THEN** the edit proceeds and Claude receives concise guidance identifying the finding and approved documentation alternatives

#### Scenario: Existing findings do not block unrelated edits

- **WHEN** a supported edit preserves a finding already present in the target file without adding or changing a finding
- **THEN** both plugin modes allow the edit without policy output

#### Scenario: Edit replacement semantics are preserved

- **WHEN** an `Edit` event supplies `replace_all` as either `false` or `true`
- **THEN** the plugin evaluates the content produced by replacing the first matching occurrence or all matching occurrences respectively

### Requirement: Preserve non-code and unsupported write behavior

The plugins SHALL not deny or warn on Markdown, README files, unsupported file types, literal comment-shaped text, or clean supported edits. They SHALL not register policy hooks for `Bash`, `PowerShell`, `NotebookEdit`, MCP tools, read-only tools, `PostToolUse`, or other unsupported events.

#### Scenario: Markdown remains writable

- **WHEN** Claude writes Markdown content containing documentation comments
- **THEN** the tool call proceeds without comment-policy output

#### Scenario: Literal comment-shaped text remains writable

- **WHEN** Claude adds comment-shaped text inside a recognized literal, raw string, template, escaped literal, heredoc, or triple-quoted literal
- **THEN** the tool call proceeds without comment-policy output

#### Scenario: Unsupported tools remain outside the boundary

- **WHEN** Claude uses Bash, PowerShell, NotebookEdit, MCP, or another tool not covered by the plugin matcher
- **THEN** the plugin performs no comment-policy enforcement for that event

### Requirement: Provide distinct installable Claude plugin variants

The project SHALL provide separately installable hard-block and warn-mode Claude Code plugins. Each plugin SHALL include valid Claude metadata, an anchored `PreToolUse` matcher for `Edit` and `Write`, a self-contained bundled launcher, and the no-code-comments guidance skill. The plugins SHALL work without downloading a separate executable at hook runtime.

#### Scenario: User installs the hard-block plugin

- **WHEN** a user enables the Claude hard-block plugin
- **THEN** Claude Code loads its pre-edit denial behavior for `Edit` and `Write`

#### Scenario: User installs the warn plugin

- **WHEN** a user enables the Claude warn plugin
- **THEN** Claude Code loads the same pre-edit scanner coverage with warning behavior and no denial for findings

#### Scenario: Bundled launcher runs offline

- **WHEN** an enabled plugin receives a supported hook event without network access
- **THEN** it evaluates the event using its bundled JavaScript implementation

### Requirement: Explain compliant alternatives through a skill

The bundled guidance skill SHALL state that supported code comments are prohibited and SHALL direct agents to use Git history, pull-request descriptions, simplified code or nearby README files, and Markdown for appropriate explanations. It SHALL explain that unsupported write paths are enforced by the repository scanner and GitHub Action in CI.

#### Scenario: Agent receives replacement guidance

- **WHEN** the skill is loaded for a task involving a prohibited comment
- **THEN** it gives concise alternatives for history, rationale, complexity, and general documentation

### Requirement: Enforce unsupported write paths in CI

The repository scanner and GitHub Action SHALL remain the authoritative final-state enforcement for comments introduced through Bash, PowerShell, NotebookEdit, MCP, generators, redirection, or other unsupported write paths. Claude plugins SHALL document this boundary and SHALL not claim to prevent those writes.

#### Scenario: CI catches an unsupported write

- **WHEN** an unsupported Claude tool introduces a comment into the repository
- **THEN** the repository scanner or GitHub Action reports the final-state finding even though the Claude plugin did not pre-enforce the tool event
