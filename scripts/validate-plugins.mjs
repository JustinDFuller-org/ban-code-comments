import * as fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { CLI_VERSION } from "../src/version.js";

const root = fileURLToPath(new URL("..", import.meta.url));
const marketplacePath = path.join(root, ".agents", "plugins", "marketplace.json");
const marketplace = JSON.parse(await fs.readFile(marketplacePath, "utf8"));
const expected = new Map([
  ["ban-code-comments-hard-block", "hard"],
  ["ban-code-comments-warn", "warn"],
]);

if (marketplace.name !== "ban-code-comments") throw new Error("marketplace name is invalid");
if (marketplace.plugins?.length !== expected.size) throw new Error("marketplace plugin count is invalid");

for (const entry of marketplace.plugins) {
  const mode = expected.get(entry.name);
  if (!mode) throw new Error(`unexpected marketplace plugin ${entry.name}`);
  if (entry.source?.source !== "local" || entry.source.path !== `./plugins/${entry.name}`) throw new Error(`${entry.name} source path is invalid`);
  const pluginRoot = path.join(root, entry.source.path.slice(2));
  const manifest = JSON.parse(await fs.readFile(path.join(pluginRoot, ".codex-plugin", "plugin.json"), "utf8"));
  if (manifest.name !== entry.name || manifest.version !== CLI_VERSION) throw new Error(`${entry.name} manifest/version is invalid`);
  const hooks = JSON.parse(await fs.readFile(path.join(pluginRoot, "hooks", "hooks.json"), "utf8"));
  const matcher = "^(apply_patch|edit|write|write_file|file_write)$";
  for (const event of ["PreToolUse"]) {
    const definitions = hooks.hooks?.[event];
    if (!Array.isArray(definitions) || definitions.length !== 1 || definitions[0].matcher !== matcher) throw new Error(`${entry.name} ${event} matcher is invalid`);
    const command = definitions[0].hooks?.[0]?.command;
    if (definitions[0].hooks?.[0]?.type !== "command" || command !== `node \${PLUGIN_ROOT}/bin/launcher.js --mode ${mode}`) throw new Error(`${entry.name} ${event} command is invalid`);
  }
  if (hooks.hooks?.PostToolUse) throw new Error(`${entry.name} PostToolUse must be absent`);
  const bundle = await fs.stat(path.join(pluginRoot, "bin", "launcher.js"));
  if (!bundle.isFile() || bundle.size < 1000) throw new Error(`${entry.name} launcher bundle is missing or empty`);
  const skill = await fs.readFile(path.join(pluginRoot, "skills", "no-code-comments", "SKILL.md"), "utf8");
  if (!skill.startsWith("---\n") || !skill.includes("Git history") || !skill.includes("Markdown")) throw new Error(`${entry.name} guidance skill is invalid`);
}

console.log(`validated ${expected.size} Codex plugins at CLI version ${CLI_VERSION}`);
