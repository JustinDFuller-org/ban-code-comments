## Local verification

- `node --test test/release-workflow.test.js` passed all 3 focused workflow and documentation tests.
- `npm test` passed all 59 tests.
- `npm run coverage` passed with 93.74% JavaScript line coverage.
- `git diff --check` passed.
- `npm run check` reached the generated-distribution comparison but local Node `v22.22.0` regenerated `dist/index.js` differently; hosted CI on the required Node 24 runtime passed the release commit.

## Live governance verification

- The `npm-publish` environment is active with a required reviewer of `JustinDFuller` and a deployment branch policy of tag pattern `v*.*.*`.
- The active `Releases` tag ruleset targets all tags and includes creation, update, deletion, and non-fast-forward protections.
- `.github/workflows/release.yml` supplies `id-token: write`, uses the `npm-publish` environment, invokes `npm publish --access public`, and contains no long-lived npm token or repository write permission.
- Public npm metadata for `@justindfuller/ban-code-comments@1.0.2` confirms the published package and registry integrity signature. npm does not expose the account-level trusted-publisher registration through public package metadata, so that registration remains an external configuration prerequisite rather than independently observable evidence.
- Historical release run `34757757995` used the removed `NPM_TOKEN` path and failed with npm `EOTP` after signing provenance; this confirms that a future release must prove the configured trusted publisher instead of relying on token fallback.
- No successful `workflow_dispatch` release from a protected semantic tag has been run; the manual OIDC publication path remains unproven until the npm publisher mapping is confirmed.
- The package Trusted Publisher configuration provided for verification identifies `JustinDFuller-org/ban-code-comments`, workflow `release.yml`, environment `npm-publish`, and permits `npm publish` plus `npm stage publish`.
