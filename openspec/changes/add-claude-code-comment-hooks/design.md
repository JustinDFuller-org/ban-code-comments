## Context

See proposal.md for motivation. The repository already has a JavaScript scanner, differential finding logic, generated Codex plugin launchers, and two Codex plugin variants. Claude Code command hooks receive JSON on stdin and use Claude-specific nested `hookSpecificOutput` fields for `PreToolUse` decisions. Claude plugin files are cached inside the plugin directory, so launchers must be self-contained and reference `${CLAUDE_PLUGIN_ROOT}`.

## Goals / Non-Goals

**Goals:**

- Reuse the scanner and finding model as the sole policy source of truth.
- Translate Claude `Write` and `Edit` payloads into the existing proposed-content evaluation model.
- Preserve hard-block, warn, unchanged-finding, literal, and documentation behavior across platforms supported by the Node runtime.
- Distribute both variants from this repository through a Claude marketplace catalog and validate the installed plugin shape.

**Non-Goals:**

- Auditing or preventing Bash, PowerShell, MCP, NotebookEdit, generators, redirection, or other opaque writes at the hook boundary.
- Changing the CLI, GitHub Action, Codex plugins, scanner language rules, or comment categories.
- Adding a runtime service, network download, or second scanner implementation.

## Decisions

- **Use a Claude-specific protocol adapter.** Keep Codex-compatible behavior stable and add an adapter that accepts only `PreToolUse` `Edit` and `Write` events. `Write` maps directly from `file_path` and `content`; `Edit` maps from `file_path`, `old_string`, `new_string`, and `replace_all`.
- **Honor Claude replacement semantics.** Reconstruct the first occurrence when `replace_all` is false and every occurrence when it is true. Missing or otherwise unreconstructible proposals produce an operational response rather than silently skipping evaluation.
- **Use Claude-native structured output.** Hard mode returns nested `permissionDecision: "deny"` and `permissionDecisionReason`. Warn mode omits a permission decision, returns nested `additionalContext` for Claude, and retains a top-level `systemMessage` for the user. Hard operational failures deny; warn operational failures report without denying.
- **Use separate plugin directories and a separate marketplace catalog.** Add two `.claude-plugin/plugin.json` manifests, root-level `hooks/hooks.json` files, generated launchers, and guidance skills. The catalog uses relative sources because it is hosted in the same Git repository.
- **Keep the file boundary explicit.** Accept Claude's absolute `file_path` values, display paths relative to `cwd` when possible, and rely on Claude's own permission system for authorization. Do not add `PostToolUse` or `FileChanged` state tracking because those events cannot provide equivalent pre-mutation prevention.
- **Validate at two levels.** Structural repository tests verify manifests and hook contracts; `claude plugin validate` and an optional authenticated `claude -p --plugin-dir` smoke provide local runtime evidence without making live model access a CI prerequisite.

## Risks / Trade-offs

- [Risk] Claude or future tools can write through paths outside `Edit` and `Write`. -> Document the limitation and retain the scanner and GitHub Action as final-state enforcement.
- [Risk] Claude hook output schemas can evolve. -> Keep the adapter isolated, test the exact installed CLI shape, and use the documented nested `hookSpecificOutput` contract.
- [Risk] Hook commands depend on `node` being available in the Claude process environment. -> Preserve the package's Node.js 24 requirement and report launcher/runtime failures clearly.
- [Risk] Live Claude smoke is nondeterministic and may require credentials. -> Make protocol and manifest validation deterministic, and report authenticated runtime smoke separately when it is run.

## Migration Plan

Build and validate the new bundles, publish them with the matching package version, and add the Claude marketplace entry. Users can install either variant independently; rollback is disabling or uninstalling the selected Claude plugin. Existing Codex, Action, and CLI workflows remain unchanged.
