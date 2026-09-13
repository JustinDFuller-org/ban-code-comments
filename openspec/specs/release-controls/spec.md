# release-controls Specification

## Purpose

Define a manual, administrator-controlled release process that protects npm publication and GitHub Action references through immutable semantic tags, environment approval, and trusted publishing.

## Requirements

### Requirement: Publish npm releases only from manually selected semantic tags

The repository SHALL expose a manually dispatchable release workflow that runs against a protected `vMAJOR.MINOR.PATCH` tag. The workflow SHALL publish only when the selected tag identifies the current `main` commit and matches the package version being released.

#### Scenario: Valid tagged release is selected

- **WHEN** a repository administrator dispatches the release workflow using a protected semantic-version tag pointing to the current `main` commit
- **THEN** the workflow validates the tag and package version and proceeds to the npm publication gate

#### Scenario: Non-semantic or stale release ref is selected

- **WHEN** the release workflow is dispatched from a ref that is not `vMAJOR.MINOR.PATCH`, or whose commit differs from the current `main` commit
- **THEN** the workflow fails before publishing the npm package

### Requirement: Gate npm publication through the protected trusted-publisher environment

The npm release SHALL execute through the `npm-publish` environment, require approval from the `JustinDFuller` account, and use GitHub Actions OIDC trusted publishing for the repository's npm package. The release workflow SHALL not require a long-lived npm token or repository content write permission.

#### Scenario: Required environment approval is granted

- **WHEN** a valid tagged release reaches the `npm-publish` environment and `JustinDFuller` approves it
- **THEN** the workflow publishes the matching package version using the configured trusted publisher

#### Scenario: Required environment approval is absent

- **WHEN** a valid tagged release reaches the `npm-publish` environment without approval from `JustinDFuller`
- **THEN** the package is not published

### Requirement: Keep floating GitHub Action tags under manual administrator control

Floating Action references such as `v1` SHALL be protected by repository tag rulesets and SHALL be created or advanced manually by a repository administrator after the corresponding semantic-version release succeeds. The release workflow SHALL not create, update, force-update, or delete floating tags.

#### Scenario: Administrator promotes a published Action release

- **WHEN** the npm publication succeeds and a repository administrator manually advances the protected `v1` tag to the released commit
- **THEN** consumers using `@v1` resolve the administrator-selected released commit

#### Scenario: Release workflow completes

- **WHEN** the npm release workflow finishes successfully
- **THEN** no floating Action tag is changed automatically

### Requirement: Protect all release tag mutations

Semantic-version tags and floating Action tags SHALL be covered by active repository tag rulesets that prevent unauthorized creation, update, deletion, and non-fast-forward changes. Release publication and floating-tag promotion SHALL remain separate administrator actions.

#### Scenario: Unauthorized tag mutation is attempted

- **WHEN** a non-authorized actor attempts to create, update, delete, or rewrite a release tag
- **THEN** the repository tag ruleset rejects the mutation and no release reference changes
