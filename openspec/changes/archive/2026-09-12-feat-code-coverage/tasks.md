## 1. Establish the coverage baseline

- [x] 1.1 Measure Go production coverage and identify the largest uncovered CLI, hook, scanner, and discovery paths; verify the baseline report and uncovered-function inventory are reproducible with `go test ./... -coverprofile` and `go tool cover`.
- [x] 1.2 Define the production-source inclusion and exclusion set for both languages; verify tests, generated bundles, plugin bundles, fixtures, and coverage output are absent from the measured source set.

## 2. Raise Go test coverage

- [x] 2.1 Add focused tests for uncovered CLI execution and error paths, using testable seams where necessary; verify `go test ./cmd/ban-code-comments ./...` passes and the Go profile includes the exercised production paths.
- [x] 2.2 Add focused tests for uncovered hook, scanner, discovery, registry, and reporting branches without changing their existing behavior; verify targeted tests and `go test ./...` pass.
- [x] 2.3 Convert the Go profile to strict Cobertura XML with the pinned converter and verify repository-relative filenames, line counts, and a total production coverage of at least 90 percent.

## 3. Add JavaScript coverage reporting

- [x] 3.1 Add the pinned `c8` development dependency and a coverage script for the existing `node:test` suite; verify `npm ci` and the coverage script produce text results and Cobertura XML for all `src/**/*.js` production files.
- [x] 3.2 Add regression coverage for any newly exposed JavaScript paths needed to keep the existing line-coverage baseline from declining; verify `npm test`, the coverage threshold/report command, and the existing Action build remain green.

## 4. Integrate coverage into GitHub Actions

- [x] 4.1 Update the CI test job to generate Go and JavaScript reports with identical source scope on default-branch pushes and pull requests; verify workflow YAML structure and local report commands.
- [x] 4.2 Upload both reports as one short-retention artifact and add a dependent native coverage job with only `contents: read`; verify the pinned artifact action receives the expected files.
- [x] 4.3 Add the repository-owned weighted aggregate coverage script and native `coverage` check with a 90 percent minimum; verify fork pull requests use the same read-only path.

## 5. Validate the repository integration

- [x] 5.1 Run the complete local validation suite, including Go tests, race tests, vet, formatting, JavaScript tests, package/build checks, report validation, strict OpenSpec validation, and `git diff --check`; verify all required checks pass.
- [x] 5.2 Run the workflow on `main` and a same-repository pull request; verify the native `coverage` check passes at the measured aggregate and publishes the report summary and artifacts.

## 6. Migrate organization ownership and activate enforcement

- [x] 6.1 Migrate the Go module/import path, release URLs, Action examples, plugin metadata, tests, and generated bundles to `JustinDFuller-org`; verify no repository-owned old-owner references remain.
- [x] 6.2 Exercise a passing pull request, a deliberately below-threshold pull request, and a fork-style read-only path; verify the native check succeeds or fails truthfully at 90 percent.
- [x] 6.3 Verify the organization ruleset can require the native `coverage` status check as an administrator follow-up; document that the current GitHub App lacks ruleset-write permission and leave the existing ruleset unchanged.
- [x] 6.4 Document the coverage commands, native Actions gate, artifact retention, organization namespace, fork behavior, and rollback procedure; verify documentation matches the workflow.
