## Why

The repository's canonical comment scanner is implemented in Go, while the GitHub Action and Codex plugins are already JavaScript wrappers that must download and execute that separate implementation. Maintaining two integration layers around a new, unused package adds release, caching, checksum, CI, and platform complexity without providing user value.

## What Changes

- **BREAKING** Replace the Go CLI and scanner with a Node.js 24+ `ban-code-comments` npm package.
- Preserve the existing CLI options, supported-language behavior, discovery rules, finding model, JSON/text reports, exit statuses, and Codex hard-block/warn semantics.
- Add a documented programmatic JavaScript API used by the CLI, GitHub Action, and Codex plugins.
- Add a temporary Go-versus-JavaScript differential harness that compares semantic results and exit statuses over the existing fixtures, options, errors, and hook events.
- Replace binary downloading, checksums, GoReleaser, and Go-specific pipelines with npm packaging, bundled JavaScript Action/plugin distributions, and Node-based CI and release validation.
- Delete the Go implementation, Go tests and artifacts, Go release configuration, and the temporary differential harness after parity is established.

## Capabilities

### New Capabilities

- `comment-detection-cli`: A publishable JavaScript CLI and programmatic comment-detection API with the existing scan, report, and hook contracts.

### Modified Capabilities

- `github-action`: Run the shared bundled JavaScript implementation directly instead of downloading a matching standalone Go binary, while preserving Action inputs, reports, statuses, permissions, and runner support.
- `codex-comment-hooks`: Use the shared bundled JavaScript hook evaluator directly instead of bootstrapping a checksum-verified Go executable, while preserving plugin variants and enforcement boundaries.

## Impact

The package metadata, JavaScript source, generated Action and plugin bundles, tests, fixtures, README, CI workflows, release workflow, and OpenSpec capability contracts will change. Go tooling, GoReleaser, binary archives, checksum downloads, platform-specific executable caches, and the current Go coverage pipeline will be removed. The public package will expose a CLI bin plus programmatic scan, discovery, result, and hook evaluation functions; no compatibility layer for the unpublished Go package is required.
