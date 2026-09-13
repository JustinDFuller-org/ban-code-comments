import { spawnSync } from "node:child_process";
import { mkdtemp, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("..", import.meta.url));
const plugins = [
  { name: "ban-code-comments-claude-hard-block", mode: "hard" },
  { name: "ban-code-comments-claude-warn", mode: "warn" },
];
const observations = [];

function command(args, cwd = root) {
  const result = spawnSync(args[0], args.slice(1), { cwd, encoding: "utf8", timeout: 120000 });
  return { status: result.status, output: `${result.stdout || ""}${result.stderr || ""}`.trim() };
}

function observation(name, status, details = "") {
  observations.push({ name, status, ...(details ? { details } : {}) });
}

const smokeRoot = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-claude-smoke-"));
const filePath = path.join(smokeRoot, "main.go");
await writeFile(filePath, "package main\nvar value = 1\n");

for (const plugin of plugins) {
  const pluginRoot = path.join(root, "plugins", plugin.name);
  const validation = command(["claude", "plugin", "validate", pluginRoot]);
  observation(`${plugin.name}-validate`, validation.status === 0 ? "passed" : "failed", validation.output);
  const event = JSON.stringify({ cwd: smokeRoot, hook_event_name: "PreToolUse", tool_name: "Write", tool_input: { file_path: filePath, content: "package main\n// smoke finding\n" } });
  const result = spawnSync("node", [path.join(pluginRoot, "bin", "launcher.js"), "--mode", plugin.mode], { cwd: root, input: `${event}\n`, encoding: "utf8", timeout: 120000 });
  let response;
  try {
    response = JSON.parse(result.stdout);
  } catch (error) {
    observation(`${plugin.name}-launcher`, "failed", `${result.stderr || result.stdout || error.message}`.trim());
    continue;
  }
  const passed = plugin.mode === "hard" ? response.hookSpecificOutput?.permissionDecision === "deny" : response.hookSpecificOutput?.permissionDecision === undefined && typeof response.hookSpecificOutput?.additionalContext === "string";
  observation(`${plugin.name}-launcher`, passed ? "passed" : "failed", JSON.stringify(response));
}

if (process.env.CLAUDE_SMOKE_LIVE !== "1") {
  observation("authenticated-live-smoke", "unrun", "set CLAUDE_SMOKE_LIVE=1 to opt into claude -p --plugin-dir coverage");
} else {
  const repository = process.env.CLAUDE_SMOKE_REPOSITORY || smokeRoot;
  for (const plugin of plugins) {
    const result = command(["claude", "-p", "--plugin-dir", path.join(root, "plugins", plugin.name), "--output-format", "json", "Create a source comment in main.go so the pre-edit hook can evaluate it."], repository);
    observation(`${plugin.name}-authenticated-live`, result.status === 0 ? "passed" : "failed", result.output);
  }
}

for (const item of observations) console.log(JSON.stringify(item));
if (observations.some((item) => item.status === "failed")) process.exitCode = 1;
