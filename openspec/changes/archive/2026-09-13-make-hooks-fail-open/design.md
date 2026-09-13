## Context

Codex currently passes the `process.stdin` stream directly to `JSON.parse`, which coerces it to `"[object Object]"`.
Both adapters also map operational failures to hard-mode denial, even when no policy finding was successfully evaluated.

## Goals / Non-Goals

**Goals:**
- Normalize launcher input before event evaluation, including the Codex stdin path.
- Make decoding, reconstruction, filesystem, and evaluator failures visible but non-blocking in both integrations.
- Preserve hard denial for successfully evaluated newly introduced findings.

**Non-Goals:**
- Change scanner policy, supported tools, finding comparison, or CI enforcement.
- Expand plugin pre-enforcement to opaque write paths.

## Decisions

- **Input:** Read async stdin streams as UTF-8 text before JSON decoding, and keep the shared runner tolerant of already-decoded event values used by programmatic callers.
- **Diagnostics:** Codex operational responses retain `systemMessage` and hook context without `decision: block`; Claude responses retain `systemMessage` and `additionalContext` without `permissionDecision: deny`.
- **Boundary:** Catch invalid modes, input decoding, event evaluation, and launcher-level failures with the same fail-open diagnostic contract. Successful finding responses remain independent.
- **Artifacts:** Regenerate all four plugin launchers from corrected source entrypoints so installed artifacts match source behavior.

## Risks / Trade-offs

- [Reduced enforcement] A malformed event, unavailable file, or runtime error can bypass pre-tool policy. -> Preserve diagnostics and repository scanner/CI final-state authority.
- [Protocol compatibility] Diagnostic fields must not accidentally be interpreted as denial. -> Test native Codex and Claude response shapes for every operational-error mode.
- [Artifact drift] Source fixes without rebuilt bundles would leave installed plugins broken. -> Validate and smoke-test every generated launcher.

## Migration Plan

- Ship source, specifications, tests, documentation, and regenerated bundles together.
- Validate source and bundled launchers with valid, malformed, clean, finding, and unevaluable events.
- Roll back the plugin release if runtime smoke tests show a protocol or enforcement regression.

## Open Questions

None.
