## Why

The Codex launcher currently passes the `process.stdin` stream directly to `JSON.parse`, which coerces it to `"[object Object]"` and turns every normal invocation into an operational error. In hard mode, both Codex and Claude adapters fail closed for malformed or otherwise unevaluable hook input, blocking agent work because the policy tool itself could not evaluate the proposal.

## What Changes

- Correct Codex hook input handling so JSON received on stdin is decoded before evaluation.
- Make operational hook failures fail open in both Codex and Claude integrations while preserving a visible, non-blocking diagnostic.
- Preserve hard-mode denial for actual newly introduced scanner findings and existing warn-mode guidance.
- Add regression coverage for launcher stdin handling, malformed input, invalid modes, malformed proposals, and generated plugin bundles.
- Document the fail-open boundary and regenerate the four self-contained Codex and Claude plugin launchers.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `codex-comment-hooks`: Normal hook input must be decoded correctly, and operational failures must allow the tool call with diagnostic context rather than deny it.
- `claude-code-comment-hooks`: Operational failures must allow the tool call with diagnostic context rather than deny it.

## Impact

- Changes the shared Codex and Claude hook adapters and their launcher entrypoints.
- Changes the observable response for malformed or unevaluable hook events from fail-closed to fail-open; genuine comment findings remain enforced.
- Updates both hook specifications, user-facing documentation, tests, and generated plugin bundles. No scanner, CLI, Action, or unsupported-write policy changes are intended.
