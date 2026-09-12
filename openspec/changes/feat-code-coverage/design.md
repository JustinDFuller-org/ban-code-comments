## Context

See proposal.md - Why. The repository currently has one CI workflow that runs the Go suite, the JavaScript Action tests, static checks, and release validation. There is no coverage report or threshold. A local Go profile reports 80.6 percent statement coverage, while the existing Node test runner reports 95.73 percent line coverage for the JavaScript source and test process.

## Goals / Non-Goals

**Goals:**

- Publish comparable Go and JavaScript Cobertura reports to GitHub Code Quality.
- Make aggregate line coverage visible on pull requests and enforce a minimum of at least 90 percent after the baseline is raised.
- Keep the privileged upload permission isolated from candidate test execution where practical.
- Establish hosted evidence for evaluation, activation, fork behavior, and rollback.

**Non-Goals:**

- Changing the CLI, Action inputs, scanner policy, hook behavior, or release artifacts.
- Enforcing separate per-language thresholds; the selected policy is one repository aggregate.
- Adding Codecov, Coveralls, or another external coverage service.
- Treating branch, function, or statement coverage as GitHub's enforcement metric; GitHub evaluates line coverage.

## Decisions

### Extend the existing CI workflow with a separate upload job

The test job will generate both reports and upload them as one artifact. A dependent upload job will download the artifact and call GitHub's coverage uploader twice, once for Go and once for JavaScript. This avoids running the test suites twice and keeps `code-quality: write` off the test job. A separate workflow was considered, but would duplicate the current test setup or require additional coordination.

### Use GitHub Code Quality as the reporting and enforcement system

The repository will use `actions/upload-code-coverage` pinned to the verified `v1` commit `1c15be36fc3733ba839b1dd643bd9556e4426dc1`, with `actions/upload-artifact` pinned to the verified `v4` commit `ea165f8d65b6e75b540449e92b4886f43607fa02`. This matches the requested GitHub-native workflow and avoids maintaining a third-party service. The upload job will have only `contents: read` and `code-quality: write` permissions.

### Convert each language's native output to Cobertura

Go will use `go test ./... -coverprofile` and the pinned `github.com/boumenot/gocover-cobertura@v1.5.0` converter with strict path resolution. JavaScript will add the pinned `c8` development dependency and generate Cobertura output from the existing `node:test` suite. The JavaScript report will include `src/**/*.js` and exclude tests, `dist`, plugin bundles, and coverage output.

### Establish the baseline before activating the threshold

Implementation will first add tests and narrowly scoped testability refactoring until complete Go production coverage reaches at least 90 percent. The ruleset will then be configured in Evaluate mode. After real pull requests confirm that reports, aggregate calculations, fork handling, and expected violations behave correctly, the minimum will be set to the greater of 90 percent and the hosted default-branch baseline and the ruleset will become Active. The maximum drop will be one percentage point; GitHub treats zero as disabled.

### Use pull requests without privileged candidate execution

Coverage will run from the normal `pull_request` event and will use the pull-request head commit for report mapping. The workflow will not use `pull_request_target` to execute candidate code. Fork uploads will be skipped by condition and reported as unavailable rather than green coverage publication.

## Risks / Trade-offs

- [GitHub Code Quality or coverage restriction is unavailable on the repository plan] -> Verify availability before implementation; do not silently substitute an external service.
- [Go and JavaScript reports are aggregated in a way that hides a weak individual product] -> Keep the aggregate policy explicit, publish separate labeled reports, and retain the per-file breakdown for review.
- [Cobertura path mapping is incorrect] -> Validate report filenames against repository-relative production paths locally and in a hosted run before enabling the ruleset.
- [Fork pull requests cannot upload with write permission] -> Keep fork-safe conditional uploads and verify that skipped publication does not leave a misleading required check or unblock claim.
- [The 90 percent threshold blocks existing work] -> Raise the Go baseline before activation and pilot the ruleset in Evaluate mode.

## Migration Plan

1. Add coverage generation, report validation, and targeted Go tests without enabling merge protection.
2. Run the workflow on `main` and a same-repository pull request to establish comparable GitHub baselines.
3. Enable or confirm GitHub Code Quality and create or update the default-branch ruleset in Evaluate mode.
4. Exercise a passing pull request, a below-threshold pull request, and a fork-style permission-limited path.
5. Set the final minimum from the higher of 90 percent and the hosted baseline, then change enforcement to Active.

To roll back, change the ruleset to Evaluate or disable the Restrict code coverage rule; report generation and pull-request visibility can remain enabled independently.
