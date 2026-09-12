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

### Requirement: Coverage SHALL be evaluated by a native Actions check

The workflow SHALL upload both reports as one GitHub Actions artifact and SHALL run a dependent `coverage` job that computes weighted aggregate line coverage. The job SHALL write the result to the job summary and SHALL fail below 90 percent.

#### Scenario: Pull request evaluates coverage

- **WHEN** a pull request completes its coverage tests
- **THEN** the coverage job downloads both reports, computes the weighted aggregate, and reports the result on the pull request check

#### Scenario: Coverage methodology remains comparable

- **WHEN** coverage is evaluated for a default-branch commit or a pull request
- **THEN** both runs use the same source inclusion, exclusion, test, and report-generation rules

### Requirement: Aggregate coverage SHALL be enforceable at 90 percent

The native coverage job SHALL fail when weighted aggregate line coverage is below 90 percent. The organization repository MAY require the `coverage` check through its normal status-check ruleset after hosted validation.

#### Scenario: Pull request meets the coverage bar

- **WHEN** the pull-request aggregate line coverage is at or above 90 percent
- **THEN** the coverage job succeeds and does not block the pull request

#### Scenario: Pull request falls below the coverage bar

- **WHEN** the pull-request aggregate line coverage is below 90 percent
- **THEN** the coverage job fails and a repository ruleset requiring that check blocks the pull request

### Requirement: Rollout SHALL distinguish reporting from enforcement

Repository administrators SHALL be able to run and review the native coverage check against real pull requests before requiring it in the organization repository ruleset. Coverage artifacts and summaries SHALL remain available while the threshold is being calibrated.

#### Scenario: Check is not yet required

- **WHEN** the `coverage` check is not yet required and a pull request violates the coverage policy
- **THEN** the failed check is visible without changing merge protection

#### Scenario: Check is required

- **WHEN** the organization ruleset requires the `coverage` check and a pull request violates the coverage policy
- **THEN** GitHub prevents the pull request from merging until coverage is restored or an authorized bypass is used

### Requirement: External contributions SHALL fail safely

Coverage collection SHALL continue to run for pull requests from forks using read-only permissions. A fork limitation SHALL not be reported, because the native gate and artifact upload do not require repository write permission.

#### Scenario: Fork pull request runs the native gate

- **WHEN** a fork pull request completes its tests
- **THEN** the same artifact and native coverage gate run without a privileged candidate workflow

### Requirement: Repository ownership SHALL use the organization namespace

Repository-owned Go module paths, imports, release URLs, Action examples, plugin metadata, tests, and generated bundles SHALL use `github.com/JustinDFuller-org/ban-code-comments`.

#### Scenario: Organization references are migrated

- **WHEN** a user installs the CLI, invokes the Action, installs a plugin, or the Action downloads a release
- **THEN** the documented and runtime URL resolves under `JustinDFuller-org`
