import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";

const workflow = fs.readFileSync(new URL("../.github/workflows/release.yml", import.meta.url), "utf8");
const readme = fs.readFileSync(new URL("../README.md", import.meta.url), "utf8");

test("release workflow is manually dispatched from a protected semantic tag", () => {
  assert.match(workflow, /on:\n  workflow_dispatch:\n/);
  assert.doesNotMatch(workflow, /\n  push:\n/);
  assert.match(workflow, /RELEASE_REF: \$\{\{ github\.ref \}\}/);
  assert.match(workflow, /refs\/tags\/v\(0\|\[1-9\]\[0-9\]\*\)\\\.\(0\|\[1-9\]\[0-9\]\*\)/);
  assert.match(workflow, /test "\$GITHUB_SHA" = "\$\(git rev-parse origin\/main\)"/);
  assert.match(workflow, /RELEASE_TAG: \$\{\{ github\.ref_name \}\}/);
  assert.ok(workflow.indexOf("unsupported release tag") < workflow.indexOf("npm publish --access public"));
  assert.ok(workflow.indexOf("test \"$GITHUB_SHA\"") < workflow.indexOf("npm publish --access public"));
  assert.ok(workflow.indexOf("node scripts/check-release-version.mjs") < workflow.indexOf("npm publish --access public"));
});

test("release workflow uses trusted npm publishing without repository writes", () => {
  assert.match(workflow, /environment: npm-publish/);
  assert.match(workflow, /id-token: write/);
  assert.match(workflow, /npm publish --access public/);
  assert.doesNotMatch(workflow, /contents: write/);
  assert.doesNotMatch(workflow, /NPM_TOKEN|NODE_AUTH_TOKEN/);
  assert.doesNotMatch(workflow, /git push|update-major-tag|major_tag/);
});

test("release documentation separates npm publication from floating tag promotion", () => {
  assert.match(readme, /manual administrator actions/);
  assert.match(readme, /npm-publish/);
  assert.match(readme, /JustinDFuller/);
  assert.match(readme, /floating Action tag/);
  assert.match(readme, /never creates or updates floating tags/);
});
