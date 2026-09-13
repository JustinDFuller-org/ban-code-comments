## Context

See proposal.md for motivation. The current repository has a complete Go implementation in `cmd/` and `internal/`, a JavaScript Action wrapper that downloads released Go archives, and generated Codex plugin launchers that do the same. The Go implementation is covered by unit, adversarial, fixture, discovery, report, CLI, and hook tests; the JavaScript suite currently covers wrapper, cache, input, and integration behavior.

## Goals / Non-Goals

**Goals:**

- Make one shared Node.js implementation authoritative for the CLI, Action, and Codex plugins.
- Preserve the existing observable scan, discovery, report, exit-status, and hook behavior.
- Provide a publishable npm package with a CLI executable and documented programmatic API.
- Establish parity against the working Go implementation before deleting it.
- Leave the final repository free of Go tooling, native release archives, and runtime download dependencies.

**Non-Goals:**

- Supporting the unpublished Go module or preserving its internal package API.
- Expanding language coverage, changing policy categories, or redesigning hook coverage.
- Adding a parser framework or a separate configuration file.

## Decisions

### Use a mechanical JavaScript port as the behavior source

Port the registry, discovery, lexical state machines, comment classification, result model, report formatting, CLI parsing, and hook reconstruction directly from the Go behavior. Use Node buffers and strings, explicit path normalization, and deterministic object serialization. Do not replace the lexical scanners with AST or third-party parser packages because that would create unvalidated behavior differences across the existing language and literal edge cases.

The public API SHALL provide a high-level check operation, source scanning, discovery, result/report helpers, and hook evaluation. The CLI entrypoint, Action entrypoint, and plugin launcher SHALL call those shared operations rather than maintaining separate implementations.

### Keep the public contract centered on plain data

Programmatic results SHALL use plain JavaScript objects matching the existing JSON shape: findings contain path, language, category, range, and text; summaries contain files scanned, files skipped, and findings. Hook evaluation SHALL return the existing response fields for no-op, hard-block, warn, and operational-error cases. The CLI runner SHALL accept argv and injectable streams so the same engine can be tested without a subprocess.

### Make integrations self-contained

The GitHub Action SHALL import the shared implementation and preserve its current input mapping and stream/status behavior. Its bundled `dist` output SHALL contain the scanner and have no downloader, checksum, archive, or Go executable dependency.

Each Codex plugin launcher SHALL be bundled from the shared hook entrypoint and retain its current mode-specific hook metadata and guidance skill. Plugin execution SHALL be offline-capable after installation and shall not use the current release cache or checksum bootstrap.

### Use npm metadata as the release source of truth

Rename the package to `ban-code-comments`, add the executable bin and package exports, and require Node.js 24 or newer. The package version SHALL drive npm publication, Action/plugin bundle validation, and the repository release tag. Git tags remain necessary for Action and plugin references, but GoReleaser, binary archives, checksums, and platform executable matrices are removed.

### Validate parity with a temporary semantic oracle

Before deleting Go, add a migration-only harness that invokes both implementations with the same working directory, files, argv, and hook-event stdin. Compare exit statuses and normalized semantic values: parsed JSON results, text finding tuples, debug path/reason diagnostics, and parsed hook response fields. Normalize only runtime-specific path separators and serialization details. Run the matrix across every existing fixture category, supported language, option/filter combination, invalid invocation, discovery edge case, and hook payload/tool mode. Delete the harness together with Go after it reports no mismatches.

## Risks / Trade-offs

- [A mechanical port can miss subtle scanner state behavior] -> Keep all existing real fixture files and adversarial cases, port unit tests before deletion, and require the differential matrix to pass.
- [Node filesystem and path behavior differs across operating systems] -> Run Action and package smoke tests on Linux, macOS, and Windows, normalize only in the differential comparator, and keep user-facing paths slash-normalized as before.
- [Generated bundles can drift from source] -> Rebuild Action and plugin distributions in CI and fail when generated output differs from the checked-in artifacts.
- [Removing runtime downloads changes integration failure modes] -> Test offline Action/plugin execution and make package/bundle validation part of release checks.
- [npm publication adds a supply-chain boundary] -> Validate package contents, lock the dependency graph, publish only tagged versions, and verify the installed package's bin and version before release completion.

## Migration Plan

1. Create the shared JavaScript package structure and port the Go behavior while Go remains available as the parity oracle.
2. Move the existing scanner and hook fixtures into the Node test layout and port the Go test coverage.
3. Add and run the temporary differential harness until CLI and hook semantic outputs and statuses match.
4. Switch the Action and both plugins to the shared JavaScript implementation and validate bundled artifacts plus cross-platform smoke cases.
5. Replace CI and release workflows with Node/npm checks, publish validation, and tag-driven npm release steps.
6. Remove the Go source, tests, fixtures' old locations, module files, GoReleaser configuration, binary downloader modules, and differential harness.
7. Run strict OpenSpec validation and final repository scans proving the deletion is complete.

No rollback compatibility layer is required because the package has not been distributed to users. A faulty release is corrected by publishing a new npm patch version and moving the Action major tag according to the repository's normal release process.
