## Context

The current `release.yml` starts on any `v*` tag push, publishes dotted tags through `npm-publish`, and automatically moves the major Action tag with repository write permission. The live repository already has a `v*.*.*` deployment policy and `JustinDFuller` required reviewer on `npm-publish`, plus an active tag ruleset covering all tags.

## Goals / Non-Goals

**Goals:**

- Make semantic-version npm publication an explicit workflow-dispatch action against the protected tag being released.
- Preserve OIDC trusted publishing and the existing validation, test, coverage, build, and package-integrity checks.
- Remove workflow-owned tag mutation and document manual promotion of floating Action tags.

**Non-Goals:**

- Changing package APIs, Action inputs, scanner behavior, or generated distributions.
- Changing the existing GitHub environment, npm trusted-publisher registration, or tag ruleset configuration in this change.
- Automating release creation, semantic tag creation, or floating-tag promotion.

## Decisions

- **Dispatch on the semantic tag, not `main` with a free-form input.** The workflow ref will be the protected `vMAJOR.MINOR.PATCH` tag and `github.ref_name` will be the sole release-version source. This aligns the workflow run with the `npm-publish` tag policy and prevents a version input from being paired with unrelated source.
- **Validate tag-to-main coupling before entering publication.** The workflow will require a tag ref, validate strict semantic-version syntax, fetch `origin/main`, and require the tagged commit to equal the current `main` commit. A separate package-version check will continue to require the embedded CLI version to match the tag.
- **Use environment approval as the Justin gate.** Any repository administrator who can dispatch the protected tag may start the run; the `npm-publish` environment remains the authoritative required approval from `JustinDFuller`. The workflow will not hard-code a single dispatcher identity.
- **Remove release workflow write authority.** The publication job retains `contents: read` and `id-token: write` for trusted publishing. The automatic major-tag job is deleted so the workflow cannot mutate protected tags.
- **Keep floating-tag promotion operationally separate.** After successful publication, an administrator manually advances `v1` or another compatible major tag under the tag ruleset. Documentation will describe this as a second action and will distinguish exact version references from floating Action references.

Alternatives considered: retaining push-tag publication was rejected because tag creation would remain an implicit release trigger; dispatching from `main` with a version input was rejected because it conflicts with the existing tag-only environment policy; allowing the workflow to update `v1` was rejected because it violates manual floating-tag governance.

## Risks / Trade-offs

- [A release requires two manual tag operations] -> Keep the semantic release tag and floating Action tag promotion steps explicit in the README and validate the published version before promotion.
- [A protected tag may be created before publication and then fail validation] -> Fail before `npm publish` for stale commits, malformed versions, or package mismatches; correct future releases use a new patch version.
- [Repository settings are external to the code change] -> Add acceptance checks that inspect the environment reviewer, tag deployment pattern, trusted publisher, and active tag ruleset before implementation is considered complete.
