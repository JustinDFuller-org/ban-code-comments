## Why

The repository tests its Go CLI and JavaScript Action wrapper, but CI does not currently measure or enforce how much production code those tests exercise. GitHub Code Quality can display native pull-request coverage and block regressions, so this change establishes a measurable aggregate quality bar before future changes reduce test coverage.

## What Changes

- Add reproducible Go and JavaScript coverage generation to GitHub Actions.
- Convert both reports to Cobertura XML and upload them to GitHub Code Quality for pull-request reporting.
- Raise the current Go baseline to at least 90% before enforcement; preserve the JavaScript coverage signal.
- Add a repository-level aggregate line-coverage rule with a minimum of at least 90% and a controlled maximum drop relative to the default branch.
- Roll the rule out in evaluation mode first, verify hosted behavior, then activate enforcement.
- Keep fork pull requests safe when the coverage uploader cannot receive write permission.

## Capabilities

### New Capabilities

- `code-coverage-enforcement`: Generate, publish, review, and enforce aggregate test line coverage for the Go CLI and JavaScript Action sources in GitHub pull requests.

### Modified Capabilities

- None.

## Impact

- GitHub Actions CI gains coverage report generation, artifact transfer, and Code Quality upload jobs.
- JavaScript development dependencies gain a pinned coverage reporter; the Go coverage converter remains a CI tool rather than a runtime dependency.
- Existing production APIs and Action inputs remain unchanged.
- Repository administrators must enable GitHub Code Quality and configure the default-branch ruleset after coverage is available on `main`.
