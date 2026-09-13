## 1. Release workflow trigger and validation

- [x] 1.1 Change `release.yml` from manual dispatch to protected semantic-version tag pushes and verify the workflow trigger matches the release-control scenarios.
- [x] 1.2 Split release validation from npm staging, keeping validation outside `npm-publish` with read-only permissions and verifying tag format, tagged commit, package versions, tests, coverage, plugins, generated files, and CLI version.
- [x] 1.3 Create and checksum the validated `npm pack` artifact, upload it from the validation job, download it in the staging job, and verify the checksum before staging.

## 2. Protected staged publication

- [x] 2.1 Configure the approved staging job to depend on validation, use the `npm-publish` environment, and retain only the required `contents: read` and `id-token: write` permissions.
- [x] 2.2 Install npm 11.15.0 or newer and replace direct `npm publish` with `npm stage publish` against the verified package artifact; verify no long-lived npm credential or direct-publish command remains.
- [ ] 2.3 Narrow the external npm trusted-publisher relationship to allow staged publishing without direct publishing and record the live configuration result without exposing credentials.

## 3. Documentation and verification

- [x] 3.1 Document the tag, GitHub environment approval, and npm staged-package approval sequence, including npm 2FA, rejection, retry, and the separate manual floating Action tag promotion.
- [x] 3.2 Update release workflow tests for automatic tag triggering, validation-before-approval, immutable artifact handoff, staged publishing, least-privilege permissions, invalid or stale tags, and absence of floating-tag mutation; verify the focused tests pass.
- [x] 3.3 Run repository checks, strict OpenSpec validation, workflow/documentation validation, and `git diff --check`; verify the implementation diff is limited to the release workflow, release documentation, focused tests, and task tracking.
- [ ] 3.4 Execute one hosted release verification with a new protected semantic tag, confirming automatic start, environment pause, staged package creation, npm-side review availability, and no public publication before npm approval.
