# Proposal

## Why

The repository tests its Go CLI and JavaScript Action wrapper, but CI does not currently measure or enforce how much production code those tests exercise. This change establishes a repository-owned aggregate quality bar with native GitHub Actions checks and migrates all repository-owned links and module references to the `JustinDFuller-org` organization.

## What Changes

- Add reproducible Go and JavaScript coverage generation to GitHub Actions.
- Convert both reports to Cobertura XML and retain them as short-lived GitHub Actions artifacts.
- Add a native Actions coverage gate that computes weighted aggregate line coverage and fails below 90 percent.
- Raise the current Go baseline to at least 90 percent before enforcement; preserve the JavaScript coverage signal.
- Require the native coverage job through the organization repository's normal status-check ruleset after hosted validation.
- Keep fork pull requests safe because coverage collection and the gate use only read permissions.
- Replace every old-owner repository URL and Go module/import path with `JustinDFuller-org`.

## Capabilities

### New Capabilities

- `code-coverage-enforcement`: Generate, review, and enforce aggregate test line coverage with a native GitHub Actions check.

### Modified Capabilities

- Repository ownership references and Go module namespace migrate to `JustinDFuller-org`.

## Impact

- GitHub Actions CI gains coverage report generation, artifact transfer, and a native aggregate coverage gate.
- JavaScript development dependencies gain a pinned coverage reporter; the Go coverage converter remains a CI tool rather than a runtime dependency.
- Existing production APIs and Action inputs remain unchanged.
- Repository administrators may require the `coverage` status check in the organization repository ruleset after hosted validation.
