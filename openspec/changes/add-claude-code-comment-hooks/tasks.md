## 1. Claude hook evaluation

- [x] 1.1 Add a Claude-specific hook adapter that accepts only `PreToolUse` `Edit` and `Write` events and verify unsupported event and tool inputs return no policy response.
- [x] 1.2 Reconstruct `Write` full-content proposals and `Edit` first/all replacement proposals, including absolute paths and `replace_all`, and verify malformed, missing-file, and replacement-error cases are handled by mode.
- [x] 1.3 Reuse differential scanner evaluation and Claude-native response formatting for hard denial, warn guidance, unchanged findings, literals, Markdown, unsupported files, and clean edits; verify the focused Claude hook tests pass.

## 2. Claude plugin packaging

- [x] 2.1 Add hard-block and warn Claude plugin manifests, hook configurations, bundled guidance skills, and launcher entry points using `${CLAUDE_PLUGIN_ROOT}`; verify each plugin has the required Claude directory structure.
- [x] 2.2 Extend the plugin build to generate self-contained hard-block and warn Claude launchers without runtime downloads; verify generated bundles execute offline against representative hook JSON.
- [x] 2.3 Add `.claude-plugin/marketplace.json` entries for both plugins and document marketplace installation, enablement, rollback, Node.js requirements, and the explicit unsupported-write boundary; verify documentation commands and names match the manifests.

## 3. Validation and smoke coverage

- [x] 3.1 Add structural validation for Claude manifests, marketplace metadata, anchored `Edit|Write` matchers, launcher commands, skills, and absence of unsupported hook events; verify the validator passes for both plugins.
- [x] 3.2 Add deterministic protocol tests for hard and warn behavior across Write, Edit, `replace_all`, existing findings, literals, Markdown, malformed input, and unsupported tools; verify `npm test` and coverage remain passing.
- [x] 3.3 Add a Claude smoke harness that runs `claude plugin validate`, exercises the bundled launcher, and optionally runs authenticated `claude -p --plugin-dir` scenarios; verify unauthenticated environments report live smoke as unrun rather than passed.

## 4. Integration and release readiness

- [x] 4.1 Integrate Claude validation and bundle checks into the repository’s standard check workflow without changing Codex, CLI, or GitHub Action behavior; verify the complete local check command passes.
- [ ] 4.2 Run the installed Claude CLI validation and local plugin smoke on Claude Code 2.1.216 when available, recording plugin validation, hard-block, warn, and unsupported-path results; verify runtime evidence is distinguished from protocol-only evidence.
- [ ] 4.3 Verify version coupling, generated artifacts, OpenSpec validation, and release documentation for the new Claude plugins; verify the change is ready for implementation and release review without implementing application code in this proposal.
