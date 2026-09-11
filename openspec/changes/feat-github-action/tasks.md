## 1. Action package and metadata

- [x] 1.1 Add the root `action.yml`, Node runtime entrypoint, package scripts, and committed bundled distribution; verify metadata validation and a clean bundle check pass.
- [x] 1.2 Add the exact CLI-version source of truth and release-tag validation; verify an exact semver Action release resolves to the same CLI version and no runtime version override is accepted.

## 2. Downloader and CLI invocation

- [x] 2.1 Implement runner OS/architecture mapping, explicit GoReleaser asset naming, release URL construction, archive extraction, and actionable unsupported-platform errors; verify all supported targets and Windows executable handling with unit tests.
- [x] 2.2 Implement checksum download and SHA-256 verification before extraction or execution; verify valid, mismatched, missing, and malformed checksum cases fail safely.
- [x] 2.3 Add version/platform-aware tool caching and verified executable discovery; verify cache hits avoid redundant downloads while cache misses verify before caching.
- [x] 2.4 Translate Action inputs for paths, languages, categories, includes, excludes, format, and debug into a shell-free CLI argument array; verify defaults, newline/comma list handling, spaces, glob characters, and invalid values.
- [x] 2.5 Stream CLI stdout/stderr and return exit statuses `0`, `1`, and `2` unchanged; verify clean, finding, and operational-failure subprocess scenarios.

## 3. Release and workflow integration

- [x] 3.1 Make GoReleaser archive naming explicit and keep checksum artifacts compatible with the downloader; verify `goreleaser check` and a release snapshot produce every expected supported archive name.
- [x] 3.2 Update release automation to build and publish the bundled Action alongside the matching CLI release and maintain the documented semver/major-tag policy; verify tag/version consistency before publishing.
- [x] 3.3 Add CI checks for package installation, Action bundling, generated-distribution drift, existing Go tests, static analysis, and OpenSpec validation; verify the complete local CI command set passes.

## 4. Tests and documentation

- [x] 4.1 Add unit and fixture-backed integration coverage for Action input mapping, download verification, caching, process status propagation, and all existing supported CLI behavior; verify the real checked-in fixture matrix remains green.
- [x] 4.2 Add a hosted GitHub Actions smoke matrix for Linux, macOS, and Windows covering checkout, clean scans, findings, filters, text/JSON output, debug diagnostics, and failure statuses; preserve run evidence before claiming hosted behavior works.
- [x] 4.3 Update the README with checkout-first Action examples, all inputs and defaults, supported runner targets, exact and major release references, permissions, exit semantics, checksum/release behavior, and full-SHA pinning guidance; verify every documented example matches the implemented metadata.
