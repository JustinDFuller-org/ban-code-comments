## Why

Teams that want to ban code comments need a lightweight, reliable way to detect them across the source and configuration languages in their repositories. A standalone, cross-platform CLI makes that policy enforceable locally and in automation without requiring a language runtime or a custom parser setup.

## What Changes

- Add a Go CLI that recursively detects supported code comments and returns a non-zero exit status when it finds policy violations.
- Add repository-aware file discovery, language selection, comment-category selection, and include/exclude glob flags without a checked-in configuration file.
- Add JSON-first and human-readable reporting with optional debug diagnostics for skipped files.
- Add tagged, checksummed GitHub Release distribution for macOS, Linux, and Windows binaries.
- Document standalone, GitHub Actions, and Claude hook invocation of the released CLI.

## Capabilities

### New Capabilities

- `comment-detection-cli`: Provides configurable, repository-aware detection and reporting of code comments through a portable command-line interface.

### Modified Capabilities

None.

## Impact

This adds a Go module, pure-Go lexical scanners and fixtures, CLI documentation, GitHub Actions release automation, and GoReleaser configuration. It introduces no hosted service, repository configuration format, GitHub Action package, or Claude hook package in the initial release.
