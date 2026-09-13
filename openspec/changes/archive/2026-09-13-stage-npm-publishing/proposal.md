## Why

The current release workflow requires a manual dispatch and can publish directly to npm after one GitHub environment approval. Triggering the workflow from a protected semantic-version tag and staging the package for a separate npm maintainer approval creates a second publication boundary without introducing long-lived npm credentials.

## What Changes

- Trigger the release workflow automatically when an authorized administrator creates a protected `vMAJOR.MINOR.PATCH` tag.
- Run release validation before the npm environment gate, with no OIDC publication permission in the validation job.
- Pass the exact validated package tarball and checksum to an `npm-publish` environment job.
- Replace direct `npm publish` with `npm stage publish` using the configured OIDC trusted publisher and npm 11.15.0 or newer.
- Require manual approval of the staged package in npm with 2FA before it becomes publicly available.
- Narrow the npm trusted-publisher relationship to staged publishing and retain least-privilege repository permissions.
- Preserve separate manual promotion of protected floating Action tags such as `v1`; the release workflow will not mutate them.
- Update release documentation and verification tests for the automatic trigger and three npm publication steps.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `release-controls`: Change npm release initiation from manual dispatch to protected tag creation and change publication from direct npm release to OIDC-backed staged publishing with npm-side approval.

## Impact

- Changes `.github/workflows/release.yml`, release documentation, and release workflow tests.
- Adds an immutable package-artifact handoff between validation and the approved staging job.
- Requires npm trusted-publisher configuration to allow staged publishing without direct publishing.
- Does not change package APIs, Action inputs, scanner behavior, or floating Action tag governance.
