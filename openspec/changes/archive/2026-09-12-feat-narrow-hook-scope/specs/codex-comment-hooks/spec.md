## MODIFIED Requirements

### Requirement: Enforce the comment policy before supported proposed edits

The plugins SHALL inspect reconstructible Codex `PreToolUse` file-edit events from the traditional file-write tools `apply_patch`, `edit`, `write`, `write_file`, and `file_write`. They SHALL evaluate the proposed result with the existing `ban-code-comments` language registry, scanner, and default comment categories. A hard-block plugin SHALL deny a proposed edit that introduces a finding before the tool mutates the file, while a warn plugin SHALL allow the edit and return concise model-visible guidance. Findings already present and unchanged in the target file SHALL not cause an unrelated edit to be denied.

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

The plugins SHALL not deny or warn on writes to Markdown, README files, unsupported file types, or paths that do not introduce a finding recognized by the existing scanner. They SHALL not register enforcement hooks for Bash, shell, exec, MCP, or other unsupported tools. Read-only tool calls and unrelated commands SHALL continue without policy output.

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

The project SHALL provide separately installable hard-block and warn-mode Codex plugins. Each plugin SHALL include valid Codex metadata, a `PreToolUse` hook configuration limited to the five traditional file-write tools, the `no-code-comments` guidance skill, and a launcher that obtains the exact CLI release coupled to the plugin. The launcher SHALL verify the published checksum before caching or executing the CLI.

#### Scenario: User installs the hard-block variant

- **WHEN** a user installs and trusts the hard-block plugin
- **THEN** Codex loads its traditional-file-write pre-tool denial behavior

#### Scenario: User installs the warn variant

- **WHEN** a user installs and trusts the warn plugin
- **THEN** Codex loads the same traditional-file-write policy coverage with warning behavior and no pre-tool denial

#### Scenario: CLI bootstrap integrity fails

- **WHEN** the launcher downloads an archive that does not match the published checksum
- **THEN** it refuses to execute the CLI and reports an actionable hook failure

## REMOVED Requirements

### Requirement: Audit opaque Bash writes after execution

**Reason**: Opaque Bash and other unsupported write paths are outside the plugin's high-value pre-write enforcement boundary; implementing reliable post-tool turn termination adds stateful complexity and did not provide dependable prevention.

**Migration**: Use the repository scanner or GitHub Action in CI to enforce the final repository state for comments introduced through Bash, MCP, generators, redirection, or other unsupported write paths.
