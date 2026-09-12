# codex-comment-hooks Specification

## Purpose

Provide installable Codex plugins that catch newly introduced code comments at the agent tool boundary and either prevent or explain the policy violation.

## Requirements

### Requirement: Enforce the comment policy before supported proposed edits

The plugins SHALL inspect supported Codex pre-tool file-edit events and evaluate the proposed result with the existing `ban-code-comments` language registry, scanner, and default comment categories. A hard-block plugin SHALL deny a proposed edit that introduces a finding, while a warn plugin SHALL allow the edit and return concise model-visible guidance. Findings already present and unchanged in the target file SHALL not cause an unrelated edit to be denied.

#### Scenario: Hard mode denies a newly introduced supported comment

- **WHEN** an agent proposes an `apply_patch` edit that adds an ordinary comment to a supported source or configuration file
- **THEN** the hard-block plugin denies the tool call with the affected path and finding details, and the proposed edit is not applied

#### Scenario: Warn mode allows the same proposed comment

- **WHEN** an agent proposes the same edit through the warn plugin
- **THEN** the plugin allows the tool call and returns concise guidance identifying the finding and an approved documentation alternative

#### Scenario: Existing comments do not block unrelated edits

- **WHEN** an agent edits a supported file that already contains an unchanged legacy finding without adding or modifying a finding
- **THEN** both plugin modes allow the edit

#### Scenario: Literal comment-shaped text is allowed

- **WHEN** an agent adds comment-shaped text inside a recognized string, raw string, template, escaped literal, heredoc, or triple-quoted literal
- **THEN** both plugin modes allow the edit

### Requirement: Preserve non-code and unsupported write behavior

The plugins SHALL not deny or warn on writes to Markdown, README files, unsupported file types, or paths that do not introduce a finding recognized by the existing scanner. Read-only tool calls and unrelated commands SHALL continue without policy output.

#### Scenario: Markdown documentation remains writable

- **WHEN** an agent writes explanatory content or Markdown comments to a `.md` file
- **THEN** the tool call proceeds without a code-comment policy denial

#### Scenario: Clean supported code remains writable

- **WHEN** an agent proposes a supported-language edit with no newly introduced findings
- **THEN** the tool call proceeds without a policy warning or denial

### Requirement: Audit opaque Bash writes after execution

The plugins SHALL support Bash tool events. When a Bash command's resulting supported-file content cannot be reconstructed before execution, the plugins SHALL compare the affected workspace state before and after the command and report newly introduced findings after execution. Hard mode SHALL stop the current agent turn after such a finding is detected, while warn mode SHALL allow continuation with guidance. This post-execution audit SHALL not claim to prevent an opaque command from writing before it is inspected.

#### Scenario: Opaque Bash violation is reported after execution

- **WHEN** an agent uses an opaque Bash command to add a comment to a supported file
- **THEN** the hard plugin reports the new finding and stops the current turn after the command, while the warn plugin reports the finding and allows continuation

#### Scenario: Opaque Bash clean write is allowed

- **WHEN** an opaque Bash command changes a supported file without introducing a finding
- **THEN** both plugins complete without a policy warning or denial

### Requirement: Provide distinct installable plugin variants

The project SHALL provide separately installable hard-block and warn-mode Codex plugins. Each plugin SHALL include valid Codex metadata, lifecycle hook configuration, the `no-code-comments` guidance skill, and a launcher that obtains the exact CLI release coupled to the plugin. The launcher SHALL verify the published checksum before caching or executing the CLI.

#### Scenario: User installs the hard-block variant

- **WHEN** a user installs and trusts the hard-block plugin
- **THEN** Codex loads its pre-tool denial and post-tool audit hooks with hard-block behavior

#### Scenario: User installs the warn variant

- **WHEN** a user installs and trusts the warn plugin
- **THEN** Codex loads the same policy coverage with warning behavior and no pre-tool denial

#### Scenario: CLI bootstrap integrity fails

- **WHEN** the launcher downloads an archive that does not match the published checksum
- **THEN** it refuses to execute the CLI and reports an actionable hook failure

### Requirement: Explain compliant alternatives through a skill

The bundled guidance skill SHALL state that code comments are prohibited under the supported scanner policy and SHALL direct agents to put history in Git, rationale and context in pull-request descriptions, complexity explanations in simplified code or nearby README files, general documentation in Markdown, and no redundant restatement of readable code in comments.

#### Scenario: Agent receives replacement guidance

- **WHEN** the skill is loaded for a task involving a prohibited comment
- **THEN** it gives the agent concise alternatives for history, rationale, complexity, and documentation
