## 1. Narrow the hook boundary

- [ ] 1.1 Change both plugin manifests to register only anchored `PreToolUse` hooks for `apply_patch`, `edit`, `write`, `write_file`, and `file_write`, and verify plugin validation accepts the manifests and excludes `PostToolUse`.
- [ ] 1.2 Remove Bash/shell/exec post-audit dispatch, workspace snapshot comparison, persisted hook state, and obsolete state-directory plumbing, and verify the Go package builds and unsupported events return no policy response.
- [ ] 1.3 Preserve the traditional-tool allowlist and scanner-based reconstruction behavior, and verify direct hook tests cover all five tool names plus malformed and unsupported inputs.

## 2. Align Codex responses and documentation

- [ ] 2.1 Add warn guidance to Codex's nested `hookSpecificOutput.additionalContext` while retaining the top-level compatibility warning field, and verify raw protocol tests assert both fields and hard mode still blocks before mutation.
- [ ] 2.2 Update the main Codex hook specification, README, bundled guidance skill, and CI guidance to describe traditional file-write enforcement and CI-only coverage for Bash, MCP, and other unsupported write paths; verify documentation references are consistent.

## 3. Update automated coverage

- [ ] 3.1 Replace Bash/PostToolUse unit and fixture integration scenarios with hard blocking, warn allowance, Markdown/string-literal allowance, legacy findings, clean edits, rename/delete behavior, and unsupported-tool no-op cases; verify `go test ./...` passes.
- [ ] 3.2 Update plugin structural validation and any package/launcher assertions for the PreToolUse-only configuration, and verify `npm test` plus plugin validation pass without requiring a local executable override.
- [ ] 3.3 Keep or adapt concurrent plugin-cache initialization coverage and verify parallel startup produces one valid checksum-verified released CLI cache entry.

## 4. Validate the released installation path

- [ ] 4.1 Update the real Codex smoke harness to install the plugin through the normal marketplace/plugin launcher with no `BAN_CODE_COMMENTS_EXECUTABLE` or local binary override, using `CODEX_SMOKE_REUSE_AUTH=1` only when required; verify it reports released artifact/version evidence.
- [ ] 4.2 Run local released-path Codex QA for hard blocking before file mutation, Markdown and literal allowance, warn-mode model-visible guidance, clean edits, and unsupported-tool non-enforcement; verify the old missing-hook failure does not occur.
- [ ] 4.3 Record release tag, CLI version, installed plugin versions, exact smoke command, cache/concurrency result, and any unrun hosted or cross-platform validation; verify the final report gives an explicit PASS/FAIL result.

## 5. Release and handoff

- [ ] 5.1 Update release/version coupling and publish the new plugin/CLI artifact required by the launcher, then verify the target-machine archive checksum and executable are the new release rather than v1.0.0.
- [ ] 5.2 Run the complete local validation suite against the released artifact and verify the worktree, generated OpenSpec status, and release evidence are ready for review.
