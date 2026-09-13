## Why

The current release workflow publishes automatically when a version tag is pushed and moves the floating Action tag from inside the workflow. That makes tag creation an implicit publication trigger and gives automation write authority over a protected release reference, which conflicts with the repository's manual administrator-only release policy.

## What Changes

- Require an administrator to create a protected semantic-version tag and manually dispatch `release.yml` using that tag as the workflow ref.
- Publish npm releases only through the `npm-publish` environment after the required `JustinDFuller` approval, using GitHub Actions trusted publishing with OIDC.
- Validate that the selected tag is `vMAJOR.MINOR.PATCH`, points at the current `main` commit, and matches the package's embedded version before publishing.
- Remove automatic promotion of floating Action tags; an administrator manually advances protected tags such as `v1` after a successful npm release.
- Document the separate semantic-version publication and floating-tag promotion steps.

## Capabilities

### New Capabilities

- `release-controls`: Defines manual, protected, administrator-controlled npm and GitHub Action release behavior.

### Modified Capabilities

None.

## Impact

- Changes `.github/workflows/release.yml` and release documentation.
- Adds repository-level release governance requirements without changing the npm package API, Action inputs, or runtime behavior.
- Relies on the existing `npm-publish` environment, semantic-tag deployment policy, required reviewer, tag ruleset, and npm trusted-publisher configuration.
