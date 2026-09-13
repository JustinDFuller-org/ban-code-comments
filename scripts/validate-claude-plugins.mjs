import * as fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { CLI_VERSION } from "../src/version.js";

const root = fileURLToPath(new URL("..", import.meta.url));
const marketplace = JSON.parse(await fs.readFile(path.join(root, ".claude-plugin", "marketplace.json"), "utf8"));
const expected = new Map([
  ["ban-code-comments-claude-hard-block", "hard"],
  ["ban-code-comments-claude-warn", "warn"],
]);

if (marketplace.name !== "ban-code-comments") throw new Error("marketplace name is invalid");
if (!marketplace.description?.includes("no-code-comments")) throw new Error("marketplace description is invalid");
if (marketplace.plugins?.length !== expected.size) throw new Error("marketplace plugin count is invalid");
if (new Set(marketplace.plugins.map((entry) => entry.name)).size !== expected.size || [...expected.keys()].some((name) => !marketplace.plugins.some((entry) => entry.name === name))) throw new Error("marketplace plugin variants are incomplete");

for (const entry of marketplace.plugins) {
  const mode = expected.get(entry.name);
  if (!mode) throw new Error(`unexpected Claude marketplace plugin ${entry.name}`);
  if (entry.source !== `./plugins/${entry.name}`) throw new Error(`${entry.name} source path is invalid`);
  if (entry.version !== CLI_VERSION) throw new Error(`${entry.name} marketplace version is invalid`);
  const pluginRoot = path.join(root, entry.source.slice(2));
  const manifest = JSON.parse(await fs.readFile(path.join(pluginRoot, ".claude-plugin", "plugin.json"), "utf8"));
  if (manifest.name !== entry.name || manifest.version !== CLI_VERSION) throw new Error(`${entry.name} manifest/version is invalid`);
  const hooks = JSON.parse(await fs.readFile(path.join(pluginRoot, "hooks", "hooks.json"), "utf8"));
  const definitions = hooks.hooks?.PreToolUse;
  if (!Array.isArray(definitions) || definitions.length !== 1 || definitions[0].matcher !== "^(Edit|Write)$") throw new Error(`${entry.name} PreToolUse matcher is invalid`);
  const hook = definitions[0].hooks?.[0];
  if (hook?.type !== "command" || hook.command !== `node "\${CLAUDE_PLUGIN_ROOT}/bin/launcher.js" --mode ${mode}`) throw new Error(`${entry.name} launcher command is invalid`);
  for (const event of Object.keys(hooks.hooks || {})) if (event !== "PreToolUse") throw new Error(`${entry.name} unsupported hook event ${event}`);
  const bundle = await fs.stat(path.join(pluginRoot, "bin", "launcher.js"));
  if (!bundle.isFile() || bundle.size < 1000) throw new Error(`${entry.name} launcher bundle is missing or empty`);
  const skill = await fs.readFile(path.join(pluginRoot, "skills", "no-code-comments", "SKILL.md"), "utf8");
  if (!skill.startsWith("---\n") || !skill.includes("Git history") || !skill.includes("Markdown") || !skill.includes("CI")) throw new Error(`${entry.name} guidance skill is invalid`);
}

console.log(`validated ${expected.size} Claude plugins at CLI version ${CLI_VERSION}`);
