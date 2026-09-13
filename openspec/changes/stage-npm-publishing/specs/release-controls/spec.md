## MODIFIED Requirements

### Requirement: Publish npm releases only from manually selected semantic tags

The repository SHALL automatically start the release workflow when an authorized administrator creates a protected `vMAJOR.MINOR.PATCH` tag. The workflow SHALL run against that tag, validate that it identifies the current `main` commit and matches the package version being released, and SHALL NOT require a separate manual workflow dispatch.

#### Scenario: Valid tagged release is selected

- **WHEN** a repository administrator creates a protected semantic-version tag pointing to the current `main` commit
- **THEN** the release workflow starts automatically, validates the tag and package version, and proceeds to the npm publication gate

#### Scenario: Non-semantic or stale release ref is selected

- **WHEN** a tag that does not match `vMAJOR.MINOR.PATCH`, or whose commit differs from the current `main` commit, triggers the release workflow
- **THEN** the workflow fails before staging or publishing the npm package

### Requirement: Gate npm publication through the protected trusted-publisher environment

The npm release SHALL execute through the `npm-publish` environment, require approval from the `JustinDFuller` account before the OIDC-capable staging job runs, and use GitHub Actions OIDC trusted publishing to invoke `npm stage publish`. The trusted-publisher relationship SHALL authorize staged publishing without authorizing direct `npm publish`. The release workflow SHALL not require a long-lived npm token or repository content write permission.

#### Scenario: Required environment approval is granted

- **WHEN** a valid tagged release passes validation and `JustinDFuller` approves the `npm-publish` environment
- **THEN** the workflow stages the matching validated package version for npm review using the configured trusted publisher, and the version is not publicly available until npm approval

#### Scenario: Required environment approval is absent

- **WHEN** a valid tagged release passes validation but the `npm-publish` environment is not approved
- **THEN** the package is not staged or published

#### Scenario: Staged package is approved in npm

- **WHEN** a maintainer reviews the staged package and approves it with npm 2FA
- **THEN** the staged package becomes publicly available at the matching package version
