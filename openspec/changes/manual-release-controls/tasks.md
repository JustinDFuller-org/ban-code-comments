## 1. Release workflow

- [x] 1.1 Replace push-tag publication with manual workflow dispatch against the selected semantic-version tag and derive the release version from `github.ref_name`; verify the workflow trigger and ref expression with a YAML/structure check.
- [x] 1.2 Validate the protected tag format, tagged commit against `origin/main`, and package version before publication; verify invalid, stale, and mismatched refs fail before `npm publish`.
- [x] 1.3 Retain the `npm-publish` environment, OIDC permission, pinned actions, package validation, and trusted `npm publish` command without long-lived npm credentials; verify the workflow has `id-token: write`, no npm token secret, and the expected environment.
- [x] 1.4 Remove the automatic floating-major-tag update job and all workflow repository write permission; verify the workflow contains no tag-push mutation or `contents: write` permission.

## 2. Release documentation

- [x] 2.1 Document administrator creation of the protected semantic-version tag and manual dispatch of `release.yml` using that tag; verify the README release sequence matches the workflow ref contract.
- [x] 2.2 Document the required `JustinDFuller` environment approval and trusted-publisher path; verify the documented account, environment, and OIDC flow match repository configuration.
- [x] 2.3 Document manual promotion of protected floating Action tags such as `v1` after npm publication; verify no documentation claims that the workflow moves floating tags automatically.

## 3. Governance and verification

- [ ] 3.1 Verify the `npm-publish` environment remains restricted to `v*.*.*` tags and requires `JustinDFuller` approval; record the live environment policy and reviewer result.
- [ ] 3.2 Verify npm trusted-publisher configuration identifies this repository, `release.yml`, and the `npm-publish` environment; record the live publisher configuration result without exposing credentials.
- [ ] 3.3 Verify the active tag ruleset protects semantic and floating tags against unauthorized creation, update, deletion, and non-fast-forward changes; record the live ruleset result.
- [ ] 3.4 Add or update workflow/documentation validation for valid tags, stale tags, mismatched versions, denied approvals, and absence of automatic floating-tag mutation; verify the validation passes on the proposal diff.

## 4. Final verification

- [ ] 4.1 Run repository checks, strict OpenSpec validation, YAML/workflow validation, and `git diff --check`; verify all required checks pass.
- [ ] 4.2 Confirm the proposal layer contains only planning artifacts and no implementation changes before handing it to review; verify the diff contains only `openspec/changes/manual-release-controls/**`.
