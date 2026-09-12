## Context

The repository already has a pure-Go scanner, a language registry covering supported source and configuration formats, and a checksum-verifying Node release downloader used by the GitHub Action. It has no Codex hook protocol or plugin package. The Codex Hooks API supplies JSON events to synchronous command hooks and supports pre-tool denial for `apply_patch`, Bash, and other local tools, but specialized tool paths are not a complete enforcement boundary.

## Goals / Non-Goals

**Goals:**

- Keep the Go scanner as the single source of truth for supported languages, lexical behavior, categories, and finding details.
- Evaluate reconstructible proposed edits before application and distinguish newly introduced findings from unchanged legacy findings.
- Provide fixed hard-block and warn plugin variants with a shared implementation and verified CLI bootstrap.
- Give agents short, actionable alternatives when a comment is rejected or reported.
- Test the hook process directly and through an isolated real Codex session.

**Non-Goals:**

- Reimplementing scanners in JavaScript, changing the existing CLI or GitHub Action policy, scanning Markdown, rewriting comments, or enforcing comments through a hosted service.
- Claiming prevention for opaque Bash commands or specialized Codex tools that do not pass through the documented hook path.
- Making the plugin an administrator-managed, undeletable policy; managed `requirements.toml` deployment remains an enterprise integration concern.

## Decisions

### Use a Go hook adapter over a second scanner

Add a private CLI hook entry point that reads Codex event JSON from stdin and emits Codex hook JSON on stdout. The adapter calls the existing registry and scanner, preserving the current supported-language set, literal handling, category defaults, and finding model. A JavaScript scanner is rejected because it would create policy drift; the Node layer remains a launcher and release downloader.

### Reconstruct proposed edits for pre-tool enforcement

For `apply_patch` and other reconstructible file-edit payloads, resolve target paths from the event, read their current contents, apply the proposed changes in memory or an isolated temporary workspace, and scan both states. Suppress only findings that are unchanged between the states; findings on added or modified content are policy violations. Emit a structured deny response for hard mode and an additional-context or system message response for warn mode.

### Use a stateful post-audit for opaque Bash

The Bash pre-tool hook records supported-file state before an opaque command, and its synchronous post-tool hook compares file contents afterward. It reports only newly introduced findings and can stop the current turn in hard mode, but it cannot undo or retroactively prevent the command. This makes the limitation explicit while retaining compatibility with legitimate formatter, generator, and shell workflows.

### Share a verified launcher across two plugin packages

Build a bundled Node launcher from the existing ESM downloader pattern. It selects the platform archive, downloads the exact CLI version coupled to the plugin, verifies `checksums.txt`, caches the executable under plugin data, and invokes the hook adapter with the selected fixed mode. The two plugin packages differ in metadata and hook mode wiring but share the policy and launcher source through the repository build process.

### Package hooks and guidance as Codex plugins

Create two valid `.codex-plugin/plugin.json` packages using the default `hooks/hooks.json` discovery layout. Each package contains the hook launcher and `skills/no-code-comments/SKILL.md`. Add `.agents/plugins/marketplace.json` with both variants so teams can discover and install them from the repository, and document trust review through `/hooks` after installation.

### Test at unit, package, and real Codex boundaries

Use Go tests for event decoding, patch reconstruction, state comparison, and policy responses; Node tests for launcher, mode, manifest, and checksum behavior; and checked-in fixture-driven cases spanning every registry language. Add an isolated Codex CLI test using a temporary workspace and plugin configuration to exercise actual pre-tool denial, warning, Markdown allowance, literal allowance, and opaque Bash post-audit behavior.

## Risks / Trade-offs

- [Opaque Bash can write before post-audit detection] -> State this limitation in the specification and documentation, stop the hard-mode turn afterward, and retain the existing CI/CLI scan as final enforcement.
- [Codex may change hook payloads or tool coverage] -> Keep the adapter contract tests against documented payloads, validate the installed Codex version locally, and report unsupported paths as unverified rather than green.
- [First-use CLI download adds network and release availability requirements] -> Reuse checksum verification and caching, fail safely for supported hook evaluation when the verified CLI is unavailable, and provide a local executable override for tests.
- [Non-managed hooks require review and trust] -> Make installation instructions explicit, keep hook output concise, and document that users can disable non-managed hooks unless an administrator deploys managed policy.
- [Two plugin packages can drift] -> Generate their repeated launcher/configuration assets from shared source and validate both packages in every change.

## Migration Plan

1. Add the hook adapter, plugin packages, skill, marketplace metadata, and tests without changing existing CLI or Action behavior.
2. Build and validate both plugin bundles, then test installation and trust in an isolated Codex home before documenting the repository marketplace flow.
3. Users can adopt warn mode first, then install the hard-block variant; existing repositories require no source migration because unchanged legacy comments remain editable.
4. Roll back by disabling or uninstalling the plugin; CLI, Action, and repository files remain unaffected.
