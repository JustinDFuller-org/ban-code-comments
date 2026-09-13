# codex-comment-hooks Specification

## Purpose

Provide installable Codex plugins that catch newly introduced code comments before supported traditional file-write tools mutate files and either prevent or explain the policy violation.

## Requirements

### Requirement: Codex hook operational failures fail open
The Codex plugins SHALL allow the tool call when hook input cannot be decoded or a proposed edit cannot be evaluated, while returning a visible non-blocking diagnostic that identifies the operational error.

#### Scenario: Normal stdin event is decoded
- **WHEN** a Codex plugin receives a valid JSON `PreToolUse` event on stdin
- **THEN** it evaluates the event rather than treating the stdin stream as the event value

#### Scenario: Unevaluable event does not block
- **WHEN** a Codex plugin cannot decode or evaluate a hook event or proposed edit
- **THEN** it emits a diagnostic without a hard-block decision and the underlying tool call proceeds

### Requirement: Enforce the comment policy before supported proposed edits

The plugins SHALL inspect reconstructible Codex `PreToolUse` file-edit events from the traditional file-write tools `apply_patch`, `edit`, `write`, `write_file`, and `file_write`. They SHALL evaluate the proposed result with the shared JavaScript `ban-code-comments` language registry, scanner, and default comment categories. A hard-block plugin SHALL deny a proposed edit that introduces a finding before the tool mutates the file, while a warn plugin SHALL allow the edit and return concise model-visible guidance. Findings already present and unchanged in the target file SHALL not cause an unrelated edit to be denied.

#### Scenario: Hard mode denies a newly introduced supported comment

- **WHEN** an agent proposes an edit through any traditional file-write tool that adds an ordinary comment to a supported source or configuration file
- **THEN** the hard-block plugin denies the tool call with the affected path and finding details, and the proposed edit is not applied

#### Scenario: Warn mode allows the same proposed comment

- **WHEN** an agent proposes the same edit through the warn plugin
- **THEN** the plugin allows the tool call and returns concise guidance identifying the finding and an approved documentation alternative through Codex's model-visible hook context, while retaining a compatibility warning field

#### Scenario: Existing comments do not block unrelated edits

- **WHEN** an agent edits a supported file through a traditional file-write tool that already contains an unchanged legacy finding without adding or modifying a finding
- **THEN** both plugin modes allow the edit

#### Scenario: Literal comment-shaped text is allowed

- **WHEN** an agent adds comment-shaped text inside a recognized string, raw string, template, escaped literal, heredoc, or triple-quoted literal through a traditional file-write tool
- **THEN** both plugin modes allow the edit

### Requirement: Preserve non-code and unsupported write behavior

The plugins SHALL not deny or warn on writes to Markdown, README files, unsupported file types, or paths that do not introduce a finding recognized by the shared JavaScript scanner. They SHALL not register enforcement hooks for Bash, shell, exec, MCP, or other unsupported tools. Read-only tool calls and unrelated commands SHALL continue without policy output.

#### Scenario: Markdown documentation remains writable

- **WHEN** an agent writes explanatory content or Markdown comments to a `.md` file through a traditional file-write tool
- **THEN** the tool call proceeds without a code-comment policy denial or warning

#### Scenario: Clean supported code remains writable

- **WHEN** an agent proposes a supported-language edit through a traditional file-write tool with no newly introduced findings
- **THEN** the tool call proceeds without a policy warning or denial

#### Scenario: Unsupported tools remain outside plugin scope

- **WHEN** an agent invokes Bash, shell, exec, MCP, or another tool that is not one of the five traditional file-write tools
- **THEN** the plugin does not perform comment-policy enforcement for that tool event

### Requirement: Provide distinct installable plugin variants

The project SHALL provide separately installable hard-block and warn-mode Codex plugins. Each plugin SHALL include valid Codex metadata, a `PreToolUse` hook configuration limited to the five traditional file-write tools, the `no-code-comments` guidance skill, and a self-contained launcher bundled from the shared JavaScript package. The launcher SHALL not download or execute a separate Go binary at runtime.

#### Scenario: User installs the hard-block variant

- **WHEN** a user installs and trusts the hard-block plugin
- **THEN** Codex loads its traditional-file-write pre-tool denial behavior without a runtime CLI download

#### Scenario: User installs the warn variant

- **WHEN** a user installs and trusts the warn plugin
- **THEN** Codex loads the same traditional-file-write policy coverage with warning behavior and no pre-tool denial

#### Scenario: Bundled launcher is available offline

- **WHEN** a plugin receives a supported hook event without network access
- **THEN** it evaluates the event using its bundled JavaScript implementation

#### Scenario: CLI bootstrap integrity fails

- **WHEN** the installed plugin bundle is incomplete or fails package integrity validation
- **THEN** it refuses to execute the hook evaluator and reports an actionable hook failure

### Requirement: Explain compliant alternatives through a skill

The bundled guidance skill SHALL state that code comments are prohibited under the supported scanner policy and SHALL direct agents to put history in Git, rationale and context in pull-request descriptions, complexity explanations in simplified code or nearby README files, general documentation in Markdown, and no redundant restatement of readable code in comments. It SHALL explain that unsupported write paths are enforced by the repository scanner and GitHub Action in CI.

#### Scenario: Agent receives replacement guidance

- **WHEN** the skill is loaded for a task involving a prohibited comment
- **THEN** it gives the agent concise alternatives for history, rationale, complexity, and documentation

### Requirement: Enforce unsupported write paths in CI

The repository scanner and GitHub Action SHALL remain the authoritative final-state enforcement for comments introduced through Bash, shell, exec, MCP, generators, redirection, or other unsupported write paths. The Codex plugins SHALL document this boundary and SHALL not claim to prevent those writes.

#### Scenario: CI catches an unsupported write

- **WHEN** an unsupported tool introduces a comment into the repository
- **THEN** the repository scanner or GitHub Action reports the final-state finding even though the Codex plugin did not enforce the tool event
