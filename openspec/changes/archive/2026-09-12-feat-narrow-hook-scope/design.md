## Context

The existing hook processor handles reconstructible traditional file edits, but also registers broad `PreToolUse` and `PostToolUse` matchers and maintains workspace snapshots for Bash-like tools. The current Codex runtime has demonstrated reliable pre-tool behavior for `apply_patch`; the other supported names remain compatibility inputs covered by direct protocol tests. Warn responses currently use a top-level field that is not reliably surfaced to the model.

## Goals / Non-Goals

**Goals:**

- Make the plugin boundary explicit and inexpensive: only five reconstructible traditional file-write tools are hooked.
- Preserve pre-mutation hard blocking, warn-mode allowance, scanner semantics, release-pinned CLI bootstrap, and concurrent cache safety.
- Make warn guidance available in Codex's model-visible hook context with a compatibility fallback.
- Make the smoke harness validate the released installation path without a local executable override.

**Non-Goals:**

- Preventing or auditing comments written by Bash, shell, exec, MCP, generators, redirection, or other unsupported tools.
- Replacing the repository scanner or GitHub Action as the final-state CI enforcement mechanism.
- Expanding live runtime claims to tool names the installed Codex version does not expose.

## Decisions

- **Exact manifest matcher:** Configure `PreToolUse` with an anchored matcher for the five tool names and remove `PostToolUse`. This prevents launcher overhead and accidental policy output for unrelated tools; the Go allowlist remains a defense-in-depth check.
- **Delete opaque-write state:** Remove workspace capture/compare and per-call state persistence instead of repairing post-tool turn termination. The behavior is explicitly out of scope, and CI is better suited to final repository-state enforcement.
- **Dual-format warnings:** Add Codex's nested `hookSpecificOutput.additionalContext` for warn guidance and keep the top-level `systemMessage` for compatibility with direct callers and older integrations. Hard blocking continues to use the existing top-level block response.
- **Separate runtime and protocol coverage:** The live smoke harness tests tools actually available from the installed Codex runtime, while Go protocol tests exercise all five accepted tool names and their payload forms. This avoids inventing live coverage for unavailable tools without dropping compatibility validation.
- **Released-path smoke:** Remove the harness's requirement for `BAN_CODE_COMMENTS_EXECUTABLE`; it must install from the marketplace/plugin launcher and use the release-pinned downloader. `CODEX_SMOKE_REUSE_AUTH=1` remains an explicit opt-in only when authentication is needed.

## Risks / Trade-offs

- [Risk] Unsupported tools can introduce comments before CI runs. → Document the boundary prominently and retain CI scanner/Action enforcement as the authoritative final check.
- [Risk] A future Codex tool may use a different payload shape. → Keep the explicit allowlist, direct parser tests, and live smoke evidence limited to observed runtime tools.
- [Risk] Older hook consumers may ignore nested context. → Retain the top-level warning field until compatibility support is intentionally removed.
- [Risk] Removing post-tool state changes existing behavior. → Treat it as an intentional scope reduction, update the specification and skill/docs, and add tests proving unsupported events are no-ops.

## Migration Plan

Implement and test the narrowed hooks, update documentation and smoke coverage, publish a new CLI/plugin release, then run released-path local Codex QA and concurrent cache initialization. Rollback is reinstalling the prior plugin release; CI enforcement remains unchanged throughout.
