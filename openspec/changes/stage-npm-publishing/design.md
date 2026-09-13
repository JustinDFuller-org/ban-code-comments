## Context

The existing `release.yml` is manually dispatched and places validation and direct `npm publish` in one `npm-publish` environment job. The repository already validates semantic tags against `origin/main`, checks generated package content, uses Node 24, and protects floating Action tags separately. The package already exists on npm, and the trusted publisher is the external authorization boundary for CI publication.

## Goals / Non-Goals

**Goals:**

- Start release processing from protected semantic tag creation.
- Run release validation before the environment approval while keeping OIDC permission confined to the approved staging job.
- Stage exactly the package artifact that passed validation.
- Require a separate npm maintainer approval before public availability.
- Preserve manual floating Action tag promotion and least-privilege repository permissions.

**Non-Goals:**

- Changing package APIs, Action inputs, scanner behavior, or package versioning conventions.
- Automatically creating or advancing GitHub releases or floating Action tags.
- Replacing npm trusted publishing with a long-lived token.

## Decisions

- **Use a protected tag-push trigger.** Configure the workflow for semantic-version tag pushes and remove manual dispatch. The tag remains the sole release-version source, while strict validation rejects malformed or stale refs that reach the workflow filter.

- **Split validation from staging.** An unprotected validation job checks out the tag, verifies the tag commit and package version, installs dependencies, runs tests and coverage, validates plugins, rebuilds generated files, and creates the package tarball. It receives no OIDC permission and does not reference the npm environment.

- **Pass an immutable tarball between jobs.** The validation job records a checksum and uploads the exact `npm pack` output. The approved staging job downloads and verifies that artifact rather than rebuilding after approval. This prevents a validated package and staged package from diverging.

- **Gate only the staging job.** The staging job depends on validation, targets `npm-publish`, retains `contents: read` and `id-token: write`, installs npm 11.15.0 or newer, and runs `npm stage publish` against the verified tarball. npm staged publishing defers 2FA and public availability to the maintainer approval step.

- **Narrow external trust permissions.** Update the npm trusted-publisher relationship so this workflow can use staged publishing but cannot use direct `npm publish`. The repository workflow itself will contain no direct publish command or long-lived npm credential.

- **Keep Action tag promotion separate.** The workflow will not mutate `v1` or other floating tags. Documentation will distinguish the three npm publication actions from the separate manual Action promotion required for floating consumers.

## Risks / Trade-offs

- [Validation runs code from the newly created tag before administrator approval] -> Keep the validation job free of OIDC, environment secrets, and write permissions; rely on protected tag rules and read-only repository access.
- [Artifact upload/download adds workflow complexity] -> Verify a cryptographic checksum before staging and test the handoff statically and in hosted CI.
- [A staged version cannot be staged twice without resolving the first stage] -> Document that maintainers must approve or reject an existing staged version before retrying a staging attempt for the same package version.
- [External npm trust configuration is not repository-tracked] -> Add it to the implementation acceptance checklist and verify that direct publishing is disallowed.

## Migration Plan

1. Merge the planning-approved implementation that changes the workflow and documentation.
2. Update the npm trusted-publisher permissions to allow `npm stage publish` only.
3. Create a new protected semantic-version tag on the current `main` commit.
4. Confirm the workflow validates automatically, pauses at `npm-publish`, and stages the expected package after approval.
5. Review and approve the staged package in npm with 2FA.
6. If the package must not ship, reject the staged package; if staging fails, resolve or reject the existing stage before retrying.

Rollback is a repository workflow revert plus disabling the staged-only trusted-publisher mapping if necessary. Already staged packages must be rejected in npm; an already approved package cannot be rolled back by rerunning the workflow.
