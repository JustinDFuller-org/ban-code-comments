## Purpose

Provide a trustworthy GitHub-native signal and merge-protection policy for how thoroughly the repository's Go CLI and JavaScript Action source are exercised by tests.

## ADDED Requirements

### Requirement: CI SHALL produce complete coverage reports

CI SHALL run the existing Go and JavaScript test suites with coverage enabled and SHALL produce Cobertura XML reports covering all non-generated production source in both products. Test files, generated bundles, and intentional test fixtures SHALL not be counted as production source.

#### Scenario: Coverage reports are generated for the default branch

- **WHEN** CI runs for a commit on the default branch
- **THEN** the workflow produces valid Go and JavaScript Cobertura reports using the same test scope used for pull requests

#### Scenario: Uncovered production source is represented

- **WHEN** a production source file is not exercised by the test suite
- **THEN** its uncovered lines remain in the report rather than being silently omitted

### Requirement: Coverage SHALL be published for pull-request review

The workflow SHALL upload the Go and JavaScript coverage reports to GitHub Code Quality for default-branch commits and eligible pull requests. Published results SHALL identify the pull-request head commit and SHALL expose aggregate line coverage and changed-file coverage deltas.

#### Scenario: Same-repository pull request publishes coverage

- **WHEN** a pull request from a branch in the repository completes its coverage tests successfully
- **THEN** GitHub Code Quality receives both reports for the pull-request head commit and posts the aggregate comparison against the default branch

#### Scenario: Coverage methodology remains comparable

- **WHEN** coverage is uploaded for a default-branch commit and a pull request
- **THEN** both uploads use the same source inclusion, exclusion, test, and report-generation rules

### Requirement: Aggregate coverage SHALL be enforceable at the repository threshold

The default-branch ruleset SHALL support a minimum aggregate line-coverage threshold of at least 90 percent. The configured minimum SHALL be the greater of 90 percent and the measured default-branch baseline at activation, and the ruleset SHALL also limit subsequent aggregate coverage decline according to the configured maximum-drop policy.

#### Scenario: Pull request meets the coverage bar

- **WHEN** the uploaded pull-request aggregate line coverage is at or above the configured minimum and its decline is within the configured maximum
- **THEN** the coverage rule does not block the pull request from merging

#### Scenario: Pull request falls below the coverage bar

- **WHEN** the uploaded pull-request aggregate line coverage is below the configured minimum or declines beyond the configured maximum
- **THEN** an active coverage ruleset reports a merge-blocking violation

### Requirement: Rollout SHALL distinguish evaluation from active enforcement

Repository administrators SHALL be able to evaluate the coverage rule against real pull requests before activating merge protection. Coverage reporting SHALL remain available while the rule is being calibrated.

#### Scenario: Rule is in evaluation mode

- **WHEN** the ruleset enforcement status is Evaluate and a pull request violates the configured coverage policy
- **THEN** GitHub records the would-block result without preventing the pull request from merging

#### Scenario: Rule is active

- **WHEN** the ruleset enforcement status is Active and a pull request violates the configured coverage policy
- **THEN** GitHub prevents the pull request from merging until coverage is restored or an authorized bypass is used

### Requirement: External contributions SHALL fail safely

Coverage collection SHALL continue to run for pull requests from forks, but the workflow SHALL not attempt to grant or use repository write permission for an ineligible fork upload. A fork limitation SHALL not be reported as successful coverage publication, and it SHALL not fail the workflow solely because the upload is unavailable.

#### Scenario: Fork pull request cannot upload coverage

- **WHEN** a fork pull request completes its tests but lacks permission to upload Code Quality data
- **THEN** the upload is skipped with an explicit notice, the test result remains truthful, and no privileged candidate workflow is required
