## Why

Repositories should be able to enforce the existing `ban-code-comments` policy without installing Go or manually downloading a release archive. A first-class GitHub Action will make the tool easy to adopt while keeping the tested Go CLI and its exit-code contract as the single enforcement engine.

## What Changes

- Add a root GitHub Action that downloads and runs the exact released CLI binary associated with the action release.
- Support the CLI's path, language, category, include, exclude, format, and debug options through action inputs with matching defaults and semantics.
- Select the runner's supported operating-system and architecture archive, verify it with the release checksum file, cache it, and invoke it without shell interpolation.
- Preserve CLI stdout, stderr, and exit behavior so clean scans pass, findings fail, and operational errors remain distinguishable.
- Document checkout requirements, action usage, input mapping, supported runners, release references, and secure pinning guidance.
- Add action packaging, release-coupling, unit, integration, and hosted workflow validation without changing the Go scanner implementation.

## Capabilities

### New Capabilities

- `github-action`: Provides a versioned, cross-platform GitHub Action wrapper for the released `ban-code-comments` CLI.

### Modified Capabilities

None.

## Impact

This adds root action metadata and bundled JavaScript action dependencies, a small downloader/runner implementation, action and release validation workflows, and README usage documentation. The existing Go CLI, GoReleaser archives, checksum artifacts, supported language behavior, and exit-code contract remain canonical. The action has no required GitHub token or write permission and requires callers to check out their repository before invocation.
