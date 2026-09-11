import fs from "node:fs";

const releaseTag = process.env.RELEASE_TAG || process.argv[2];
if (!releaseTag) {
  console.error("RELEASE_TAG or a release tag argument is required");
  process.exit(2);
}

const source = fs.readFileSync(new URL("../src/version.js", import.meta.url), "utf8");
const match = source.match(/CLI_VERSION\s*=\s*["']([^"']+)["']/);
if (!match) {
  console.error("could not find CLI_VERSION in src/version.js");
  process.exit(2);
}

const expectedTag = `v${match[1]}`;
if (releaseTag !== expectedTag) {
  console.error(`release tag ${releaseTag} does not match embedded CLI version ${expectedTag}`);
  process.exit(1);
}

console.log(`release tag matches embedded CLI version: ${releaseTag}`);
