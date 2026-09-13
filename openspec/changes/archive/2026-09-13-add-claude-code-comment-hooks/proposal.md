## Why

Claude Code users currently have no installable hook integration for the repository's no-code-comments policy, even though Claude Code exposes a pre-tool boundary with enough information to inspect file edits before they are applied. This change brings the existing hard-block and warn experiences to Claude Code while preserving the shared scanner and CI enforcement boundary.

## What Changes

- Add separately installable Claude Code hard-block and warn plugins for reconstructible `Edit` and `Write` tool calls.
- Reconstruct Claude Code's full writes and string replacements, compare proposed findings with existing findings, and preserve literal, Markdown, unsupported-file, and clean-edit behavior.
- Return Claude-native `PreToolUse` decisions for blocking and nested model-visible context for warnings.
- Bundle self-contained JavaScript launchers and guidance skills without runtime downloads.
- Add Claude plugin manifests, a Git-hosted Claude marketplace catalog, structural validation, protocol tests, and an optional authenticated Claude CLI smoke harness.
- Document that Bash, PowerShell, MCP, `NotebookEdit`, and other opaque or unsupported write paths remain final-state CI/Action coverage rather than pre-write plugin coverage.

## Capabilities

### New Capabilities

- `claude-code-comment-hooks`: Installable Claude Code plugins that enforce the no-code-comments policy before supported file edits.

### Modified Capabilities

None.

## Impact

- Adds Claude-specific hook adapters, launchers, generated plugin bundles, manifests, skills, marketplace metadata, tests, validation, smoke coverage, and documentation.
- Reuses the existing JavaScript scanner and language registry; no CLI flags, GitHub Action behavior, or Codex plugin behavior changes.
- Requires Claude Code plugin support and the existing Node.js 24 runtime contract for bundled hook execution.
