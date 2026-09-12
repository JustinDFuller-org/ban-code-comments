## Context

See proposal.md - Why. The repository currently has one CI workflow that runs the Go suite, the JavaScript Action tests, static checks, and release validation. There is no coverage report or threshold. A local Go profile reports 80.6 percent statement coverage, while the existing Node test runner reports 95.73 percent line coverage for the JavaScript source and test process.

## Goals / Non-Goals

**Goals:**

- Publish comparable Go and JavaScript Cobertura reports as GitHub Actions artifacts.
- Make weighted aggregate line coverage visible in the Actions job summary and enforce a minimum of at least 90 percent.
- Keep the coverage gate read-only and safe for fork pull requests.
- Establish hosted evidence for the native gate, organization status-check configuration, and fork behavior.
- Migrate repository-owned URLs and the Go module namespace to `JustinDFuller-org`.

**Non-Goals:**

- Changing the CLI, Action inputs, scanner policy, or hook behavior.
- Enforcing separate per-language thresholds; the selected policy is one weighted repository aggregate.
- Adding Codecov, Coveralls, GitHub Code Quality, or another external coverage service.
- Treating branch, function, or statement coverage as the enforcement metric; the gate evaluates line coverage.

## Decisions

### Extend the existing CI workflow with a native coverage job

The test job will generate both reports and upload them as one short-retention artifact. A dependent `coverage` job will download the artifact and run a checked-in Node script that parses both Cobertura roots, sums covered and valid lines, writes a job summary, and fails below 90 percent. This avoids running the test suites twice and needs only `contents: read`.

### Use repository-owned Actions for reporting and enforcement

The repository will use the pinned `actions/upload-artifact` v4 commit already used by CI. The native script will be the source of truth for the threshold, and the organization repository may require the resulting `coverage` check through its normal status-check ruleset. No paid Code Quality integration or privileged upload is required.

### Convert each language's native output to Cobertura

Go will use `go test ./... -coverprofile` and the pinned `github.com/boumenot/gocover-cobertura@v1.5.0` converter with strict path resolution. JavaScript will add the pinned `c8` development dependency and generate Cobertura output from the existing `node:test` suite. The JavaScript report will include `src/**/*.js` and exclude tests, `dist`, plugin bundles, and coverage output.

### Establish the baseline before requiring the threshold

Implementation will first add tests and narrowly scoped testability refactoring until complete Go production coverage reaches at least 90 percent. After real pull requests confirm report generation, weighted aggregation, threshold failures, and fork handling, the organization repository ruleset may require the `coverage` job. The fixed minimum is 90 percent.

### Use pull requests without privileged candidate execution

Coverage will run from the normal `pull_request` event. The workflow will not use `pull_request_target` to execute candidate code or require secrets. Fork pull requests run the same tests, artifact upload, and native gate with read-only permissions.

## Risks / Trade-offs

- [Actions minutes or artifact storage are constrained] -> Keep artifact retention short and make the gate depend only on the existing test job.
- [Go and JavaScript reports are aggregated in a way that hides a weak individual product] -> Keep the aggregate policy explicit, publish the per-language reports, and retain the per-file breakdown in the artifact.
- [Cobertura path mapping is incorrect] -> Validate report filenames against repository-relative production paths locally and in a hosted run before requiring the check.
- [The 90 percent threshold blocks existing work] -> Raise the Go baseline before requiring the status check.
- [Old-owner references keep downloading from the former repository] -> Migrate the Go module, imports, release URLs, Action examples, plugin metadata, tests, and generated bundles together.

## Migration Plan

1. Add coverage generation, report validation, and targeted Go tests without enabling merge protection.
2. Migrate all repository-owned URLs and module/import paths to `JustinDFuller-org`.
3. Run the workflow on `main` and a same-repository pull request to establish hosted native-gate evidence.
4. Exercise a passing pull request, a deliberately below-threshold pull request, and a fork-style read-only path.
5. Optionally require the `coverage` status check in the organization repository ruleset after the pilot is accepted; this requires repository ruleset administration permission.

To roll back, remove the `coverage` status check from the repository ruleset; report generation and artifact retention can remain enabled independently.
