## 1. Package foundation and public contracts

- [x] 1.1 Rename the npm package to `ban-code-comments`, require Node.js 24+, add the `ban-code-comments` bin and package exports, and verify `npm pack --dry-run` contains the intended executable, API modules, and documentation.
- [x] 1.2 Define the shared plain-data model, category constants, supported-language registry, aliases, extensions, and special filenames, and verify registry tests cover every existing language and alias.
- [x] 1.3 Implement CLI argument parsing, defaults, validation, injected-stream execution, and exit-status mapping, and verify clean, finding, invalid-option, and scan-error cases return 0, 1, and 2 as specified.

## 2. JavaScript scanner and discovery

- [ ] 2.1 Port the lexical scanner families, comment classification, source ranges, and literal/state handling to JavaScript, and verify unit tests cover every existing scanner family and adversarial case.
- [x] 2.2 Move the checked-in language fixtures into the Node test layout and verify finding, clean, and false-positive fixtures cover every supported language and extension mapping.
- [ ] 2.3 Port repository-aware discovery, Git root handling, `.gitignore` checks, fixed directory exclusions, symlink behavior, language filters, include/exclude glob precedence, deduplication, and debug diagnostics, and verify discovery integration tests cover each branch.
- [x] 2.4 Implement JSON and text reporting from the shared result model, and verify output field names, ordering, ranges, text, summaries, newline behavior, and writer failures.

## 3. Codex hook core and differential parity

- [ ] 3.1 Port hook event decoding, supported tool filtering, path safety, content/edit/patch reconstruction, legacy-finding subtraction, hard-block responses, warn responses, and operational errors, and verify all existing hook fixtures and protocol cases.
- [ ] 3.2 Add the temporary Go-versus-JavaScript differential harness with deterministic temporary repositories and semantic normalization for JSON, text findings, diagnostics, and hook responses, and verify it compares exit statuses and outputs for identical inputs.
- [ ] 3.3 Run the differential matrix across every fixture class, language, category/filter combination, discovery edge case, invalid invocation, and supported hook tool/mode, and verify that no semantic or status mismatches remain before removing Go.

## 4. Integration replacement

- [ ] 4.1 Replace the GitHub Action downloader/runner path with direct calls to the shared JavaScript CLI while preserving Action input parsing, stream routing, workspace selection, and exit statuses, and verify Action unit tests cover defaults, configured inputs, findings, and operational failures.
- [ ] 4.2 Replace both Codex plugin launchers with self-contained bundles of the shared JavaScript hook entrypoint, preserve their metadata, matcher scope, and guidance skills, and verify plugin validation and offline hard/warn smoke cases.
- [ ] 4.3 Rebuild `dist` and plugin launcher artifacts from source and verify generated outputs are reproducible and contain no Go executable downloader, checksum, archive, or platform-cache dependency.

## 5. CI, release, and documentation

- [ ] 5.1 Replace Go setup, Go tests, Go coverage conversion, Go vet, and GoReleaser checks with Node installation, package tests, JS coverage, package validation, Action bundle checks, and plugin checks, and verify CI reports the required aggregate coverage threshold.
- [ ] 5.2 Replace the binary release workflow with tagged npm publication and package/version verification while retaining repository tag and Action release-reference handling, and verify a dry-run release validates metadata, bin execution, and bundled versions.
- [ ] 5.3 Update the README and package documentation for npm installation, direct CLI usage, programmatic API usage, Action integration, Codex plugins, supported Node runtime, options, reports, statuses, and offline/self-contained behavior, and verify every documented command matches the test suite.
- [ ] 5.4 Update Action and cross-platform smoke workflows to exercise the bundled JavaScript implementation on Ubuntu, macOS, and Windows without downloading a native CLI, and verify clean, finding, operational, and debug cases.

## 6. Remove Go and finalize

- [ ] 6.1 Delete the Go implementation, Go tests, Go module files, Go fixtures' old locations, GoReleaser configuration, binary-release references, and obsolete downloader/checksum/platform-cache modules, and verify targeted searches find no remaining Go implementation or archive dependency.
- [ ] 6.2 Delete the temporary differential harness after the final parity run and preserve the migrated JS fixtures and semantic regression tests, and verify the final test suite no longer requires Go.
- [ ] 6.3 Run the complete Node test, coverage, package, bundle, plugin, smoke, documentation, and strict OpenSpec validation checks, and verify the worktree contains only the approved JavaScript rewrite and planning artifacts.
