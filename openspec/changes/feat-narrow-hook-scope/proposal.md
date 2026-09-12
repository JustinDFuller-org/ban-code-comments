## Why

The Codex plugins currently register broad pre- and post-tool hooks, including an opaque Bash audit that is difficult to enforce reliably and adds stateful complexity. The highest-value behavior is preventing forbidden comments before traditional file-write tools mutate files, while CI can provide comprehensive final-state enforcement for every other write path.

## What Changes

- Narrow both plugin manifests to `PreToolUse` for `apply_patch`, `edit`, `write`, `write_file`, and `file_write` only.
- Remove Bash, MCP, opaque-write, `PostToolUse`, workspace snapshot, and hook state-management behavior from the plugin.
- Preserve hard-mode pre-write blocking and warn-mode allowance for reconstructible traditional file writes.
- Make warn guidance visible through Codex hook context while retaining the existing compatibility response field.
- Update unit, integration, plugin-validation, and released-path Codex smoke coverage around the narrowed boundary.
- Update the main specification, README, bundled skill guidance, and CI guidance to explain that CI covers unsupported write paths.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `codex-comment-hooks`: limit hook enforcement to traditional reconstructible file-write tools and remove opaque Bash post-auditing.

## Impact

The Go hook processor and response model, both plugin `hooks.json` files, the JavaScript launcher/cache interface, the Codex smoke harness, hook tests and fixtures, OpenSpec documentation, README/plugin guidance, and release validation expectations are affected. The public CLI scanner and GitHub Action remain the comprehensive CI enforcement path.
