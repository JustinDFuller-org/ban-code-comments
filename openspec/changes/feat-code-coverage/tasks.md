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
- [x] 4.2 Upload the reports as one artifact and add a dependent upload job with only `contents: read` and `code-quality: write`; verify both labeled reports use immutable action pins and the upload job receives the expected files.
- [x] 4.3 Restrict Code Quality uploads to eligible default-branch and same-repository pull-request events, preserving normal `pull_request` execution and fork-safe skip behavior; verify fork handling does not claim successful publication or fail tests solely because upload permission is unavailable.

## 5. Validate the repository integration

- [x] 5.1 Run the complete local validation suite, including Go tests, race tests, vet, formatting, JavaScript tests, package/build checks, report validation, strict OpenSpec validation, and `git diff --check`; verify all required checks pass.
- [ ] 5.2 Run the workflow on `main` and a same-repository pull request; verify GitHub Code Quality shows aggregate line coverage, default-branch comparison, and changed-file deltas for both reports.

## 6. Pilot and activate enforcement

- [ ] 6.1 Enable GitHub Code Quality and configure the default-branch ruleset with `Restrict code coverage` in Evaluate mode; verify the minimum is the greater of 90 percent and the hosted default-branch aggregate baseline, with a one-percentage-point maximum drop.
- [ ] 6.2 Exercise a passing pull request, a deliberately below-threshold pull request, and a permission-limited fork-style path; verify Evaluate mode records would-block behavior, fork publication remains explicitly unavailable, and no misleading green coverage check is produced.
- [ ] 6.3 Change the ruleset to Active after the hosted pilot is accepted; verify the below-threshold pull request is merge-blocked and the restored-coverage pull request is mergeable.
- [ ] 6.4 Document the coverage commands, GitHub Code Quality prerequisite, aggregate line-coverage policy, fork limitation, and rollback procedure; verify the documentation matches the implemented workflow and ruleset.
