## 1. Hook protocol and policy evaluation

- [ ] 1.1 Add the private CLI hook entry point that decodes documented Codex event JSON and emits structured allow, deny, warning, and operational-error responses; verify Go protocol tests cover PreToolUse and PostToolUse payloads for hard and warn modes.
- [ ] 1.2 Implement reconstructible proposed-edit evaluation using the existing language registry and scanner, comparing pre-write and proposed findings so only newly introduced or modified findings are reported; verify tests cover new files, legacy findings, deletions, renames, all supported-language fixtures, literals, directives, and Markdown/non-code paths.
- [ ] 1.3 Implement per-tool workspace state capture and synchronous opaque-Bash post-audit behavior; verify tests cover clean writes, newly introduced findings, hard-mode turn-stop responses, warn-mode continuation, missing files, and pre-existing dirty worktrees.

## 2. Plugin packaging and runtime

- [ ] 2.1 Extract or share the existing ESM release downloader and runner for plugin use, including platform selection, exact CLI-version coupling, checksum verification, plugin-data caching, cleanup, and a local executable override for tests; verify Node tests cover supported targets, cache hits, checksum failures, and launch failures.
- [ ] 2.2 Create valid `ban-code-comments-hard-block` and `ban-code-comments-warn` plugin packages with manifests, synchronous hook configuration, generated launcher bundles, and the shared `no-code-comments` skill; verify both packages pass plugin validation and their hook definitions select the intended mode.
- [ ] 2.3 Add repository marketplace metadata and README installation guidance covering marketplace setup, plugin trust review, variant selection, supported-language behavior, approved documentation alternatives, opaque-Bash limitations, and rollback; verify metadata parses and documented install commands match the package names and paths.

## 3. Fixture and Codex integration coverage

- [ ] 3.1 Add checked-in hook event and proposed-content fixtures that exercise every language in the existing scanner registry plus clean, finding, false-positive, legacy-comment, Markdown, and unsupported cases; verify the fixture suite resolves paths through the real registry rather than inline-only language tables.
- [ ] 3.2 Add an isolated real-Codex smoke harness using a temporary Codex home and repository, exercising hard-mode pre-tool denial, warn-mode allowance, Markdown and literal allowance, and opaque-Bash post-audit behavior; verify the harness records observed hook decisions and distinguishes unsupported tool paths from passing coverage.

## 4. Repository verification and release readiness

- [ ] 4.1 Run Go tests, vet, formatting, Node tests, bundled-plugin validation, and the existing CLI dogfood scan; verify all pass without changing the existing CLI or GitHub Action contract.
- [ ] 4.2 Validate strict OpenSpec artifacts, generated distributions, marketplace JSON, release-version coupling, and diff cleanliness; verify `openspec validate --changes --strict`, package checks, and `git diff --check` pass before review.
