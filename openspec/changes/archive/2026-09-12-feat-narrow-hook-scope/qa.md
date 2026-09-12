# Codex Released-Path Validation

Result: PASS

Release tag: `v1.0.1`

CLI version: `1.0.1`

Installed plugin versions: `ban-code-comments-hard-block@1.0.1`, `ban-code-comments-warn@1.0.1`

Exact smoke command: `CODEX_SMOKE_REUSE_AUTH=1 npm run smoke:codex`

Installation path: temporary Codex marketplace installation using the plugin launcher and release-pinned checksum-verified cache; no `BAN_CODE_COMMENTS_EXECUTABLE` or local binary override.

Passed cases: hard pre-write denial before mutation, Markdown and string-literal allowance, unsupported Bash write outside plugin scope, and warn-mode model-visible guidance.

Concurrent cache result: Node test coverage passed for parallel startup, producing one shared valid checksum-verified released CLI cache entry.

Target archive verification: the published `ban-code-comments_1.0.1_darwin_arm64.tar.gz` checksum matched `checksums.txt`, and its executable was present and runnable.

Hosted or cross-platform validation: not exercised by this local harness; the release contains Linux amd64/arm64, macOS amd64/arm64, and Windows amd64 assets.
