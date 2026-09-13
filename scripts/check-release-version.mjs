import fs from "node:fs";

const releaseTag = process.env.RELEASE_TAG || process.argv[2];
if (!releaseTag) {
  console.error("RELEASE_TAG or a release tag argument is required");
  process.exit(2);
}

const packageData = JSON.parse(fs.readFileSync(new URL("../package.json", import.meta.url), "utf8"));
const source = fs.readFileSync(new URL("../src/version.js", import.meta.url), "utf8");
const match = source.match(/CLI_VERSION\s*=\s*["']([^"']+)["']/);
if (!match) {
  console.error("could not find CLI_VERSION in src/version.js");
  process.exit(2);
}

const versions = [
  match[1],
  packageData.version,
  JSON.parse(fs.readFileSync(new URL("../plugins/ban-code-comments-hard-block/.codex-plugin/plugin.json", import.meta.url), "utf8")).version,
  JSON.parse(fs.readFileSync(new URL("../plugins/ban-code-comments-warn/.codex-plugin/plugin.json", import.meta.url), "utf8")).version,
];
if (versions.some((version) => version !== versions[0])) {
  console.error(`release versions do not match: ${versions.join(", ")}`);
  process.exit(1);
}

const expectedTag = `v${packageData.version}`;
if (releaseTag !== expectedTag) {
  console.error(`release tag ${releaseTag} does not match embedded CLI version ${expectedTag}`);
  process.exit(1);
}

console.log(`release tag matches embedded CLI version: ${releaseTag}`);
