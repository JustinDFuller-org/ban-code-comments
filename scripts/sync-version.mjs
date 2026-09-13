import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.dirname(fileURLToPath(import.meta.url));
const packagePath = path.join(root, "..", "package.json");
const packageData = JSON.parse(fs.readFileSync(packagePath, "utf8"));
const version = packageData.version;

const sourcePath = path.join(root, "..", "src", "version.js");
const source = fs.readFileSync(sourcePath, "utf8");
const updatedSource = source.replace(/CLI_VERSION\s*=\s*["'][^"']+["']/, `CLI_VERSION = "${version}"`);
if (updatedSource !== source) fs.writeFileSync(sourcePath, updatedSource);

for (const manifest of [
  "plugins/ban-code-comments-hard-block/.codex-plugin/plugin.json",
  "plugins/ban-code-comments-warn/.codex-plugin/plugin.json",
  "plugins/ban-code-comments-claude-hard-block/.claude-plugin/plugin.json",
  "plugins/ban-code-comments-claude-warn/.claude-plugin/plugin.json",
]) {
  const manifestPath = path.join(root, "..", manifest);
  const content = fs.readFileSync(manifestPath, "utf8");
  const updated = content.replace(/("version"\s*:\s*)"[^"]+"/, `$1"${version}"`);
  if (updated !== content) fs.writeFileSync(manifestPath, updated);
}

const marketplacePath = path.join(root, "..", ".claude-plugin", "marketplace.json");
const marketplace = fs.readFileSync(marketplacePath, "utf8");
const updatedMarketplace = marketplace.replace(/("version"\s*:\s*)"[^"]+"/g, `$1"${version}"`);
if (updatedMarketplace !== marketplace) fs.writeFileSync(marketplacePath, updatedMarketplace);
