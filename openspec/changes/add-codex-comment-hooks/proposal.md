## Why

The current CLI and GitHub Action detect comments after files exist, so an agent can write prohibited comments before the policy reports them. Codex lifecycle hooks provide an earlier enforcement point that can guide agents toward better documentation and block supported pre-write paths.

## What Changes

- Add hard-block and warn-mode Codex plugins that share the existing Go scanner as their policy engine.
- Add a Codex hook adapter that evaluates proposed file edits, blocks or warns on newly introduced findings, and preserves unrelated writes.
- Add post-write auditing for opaque Bash commands whose resulting file contents cannot be reconstructed before execution.
- Add a concise guidance skill covering Git history, PR rationale, simplified code, directory README documentation, Markdown documentation, and avoiding comments that restate code.
- Add checksum-verified CLI bootstrapping, plugin manifests, hook configuration, repository marketplace metadata, installation guidance, and comprehensive fixture and Codex integration tests.

## Capabilities

### New Capabilities

- `codex-comment-hooks`: Provides installable Codex plugins that enforce or warn on newly introduced comments in supported source and configuration languages.

### Modified Capabilities

None.

## Impact

This adds a private hook-event protocol to the Go CLI, Node-based plugin launchers and generated bundles, two plugin packages, a repository marketplace catalog, user-facing installation guidance, and hook-specific tests. The existing scanner language registry, comment categories, CLI behavior, GitHub Action contract, and Markdown handling remain unchanged.
