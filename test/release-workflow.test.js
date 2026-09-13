import assert from "node:assert/strict";
import fs from "node:fs";
import test from "node:test";

const workflow = fs.readFileSync(new URL("../.github/workflows/release.yml", import.meta.url), "utf8");
const readme = fs.readFileSync(new URL("../README.md", import.meta.url), "utf8");
const validateJob = workflow.slice(workflow.indexOf("  validate:"), workflow.indexOf("  stage:"));
const stageJob = workflow.slice(workflow.indexOf("  stage:"));

test("release workflow starts from a protected semantic tag", () => {
  assert.match(workflow, /on:\n  push:\n    tags:\n      - 'v\*\.\*\.\*'/);
  assert.doesNotMatch(workflow, /workflow_dispatch/);
  assert.match(workflow, /RELEASE_REF: \$\{\{ github\.ref \}\}/);
  assert.match(workflow, /refs\/tags\/v\(0\|\[1-9\]\[0-9\]\*\)\\\.\(0\|\[1-9\]\[0-9\]\*\)/);
  assert.match(workflow, /test "\$GITHUB_SHA" = "\$\(git rev-parse origin\/main\)"/);
  assert.match(workflow, /RELEASE_TAG: \$\{\{ github\.ref_name \}\}/);
  assert.match(workflow, /persist-credentials: false/);
  assert.ok(workflow.indexOf("unsupported release tag") < workflow.indexOf("npm stage publish"));
  assert.ok(workflow.indexOf("test \"$GITHUB_SHA\"") < workflow.indexOf("npm stage publish"));
  assert.ok(workflow.indexOf("node scripts/check-release-version.mjs") < workflow.indexOf("npm stage publish"));
});

test("release workflow separates validation from trusted staged publishing", () => {
  assert.match(workflow, /validate:\n/);
  assert.match(workflow, /stage:\n    needs: validate/);
  assert.match(validateJob, /permissions:\n      contents: read/);
  assert.doesNotMatch(validateJob, /id-token: write|environment: npm-publish/);
  assert.match(stageJob, /permissions:\n      contents: read\n      id-token: write/);
  assert.match(workflow, /environment: npm-publish/);
  assert.match(workflow, /id-token: write/);
  assert.match(workflow, /npm install --global npm@11\.15\.0/);
  assert.match(workflow, /npm stage publish .*--access public/);
  assert.match(workflow, /actions\/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a/);
  assert.match(workflow, /actions\/download-artifact@3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c/);
  assert.match(workflow, /name: release-package/);
  assert.match(workflow, /path: release-artifact/);
  assert.match(workflow, /cd release-artifact\n          sha256sum \*\.tgz > SHA256SUMS/);
  assert.match(workflow, /sha256sum -c SHA256SUMS/);
  assert.match(workflow, /find release-artifact -name '\*\.tgz'/);
  assert.match(workflow, /npm stage publish "\.\/\$\(find release-artifact/);
  assert.doesNotMatch(workflow, /contents: write/);
  assert.doesNotMatch(workflow, /NPM_TOKEN|NODE_AUTH_TOKEN/);
  assert.doesNotMatch(workflow, /npm publish --access public/);
  assert.doesNotMatch(workflow, /git push|update-major-tag|major_tag/);
});

test("release documentation separates npm publication from floating tag promotion", () => {
  assert.match(readme, /three npm publication actions/);
  assert.match(readme, /npm-publish/);
  assert.match(readme, /JustinDFuller/);
  assert.match(readme, /Staged Packages/);
  assert.match(readme, /npm 2FA/);
  assert.match(readme, /floating Action tag/);
  assert.match(readme, /never creates or updates floating tags/);
});
