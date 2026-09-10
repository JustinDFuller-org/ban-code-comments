## Context

The repository is currently only a source-control project with no OpenSpec root or shared agent instruction directory. The generated OpenSpec files must remain tool-compatible and easy for all supported agents to discover.

## Goals / Non-Goals

**Goals:**

- Use the standard OpenSpec `spec-driven` schema.
- Install the shared `.agents` skills supported by the OpenSpec CLI.
- Keep the initialization self-contained and suitable for future changes.

**Non-Goals:**

- Changing application behavior, CLI behavior, or CI workflows.
- Adding agent-specific adapters or custom workflow logic.

## Decisions

- Use the CLI-generated shared `.agents/skills` output so the repository receives the canonical instructions and future OpenSpec updates remain recognizable.
- Keep the generated `openspec/config.yaml` at the repository root and use the `spec-driven` schema so changes have the standard proposal, design, specs, and tasks lifecycle.
- Record initialization as a capability rather than opting out of specifications, because the repository gains an observable contributor-facing workflow.

## Risks / Trade-offs

- [Generated instructions can evolve with the CLI] → Validate the checked-in skills and OpenSpec artifacts before opening the PR.
- [The initial capability is documentation-facing] → Keep its requirements limited to discoverability and schema behavior; runtime changes remain out of scope.

## Migration Plan

No runtime migration is required. Commit the generated structure and archive this bootstrap change so the capability becomes the repository's baseline specification.
