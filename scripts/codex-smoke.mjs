import { spawnSync } from "node:child_process";
import * as fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("..", import.meta.url));
const codex = process.env.CODEX_BIN || "codex";
const cli = process.env.BAN_CODE_COMMENTS_EXECUTABLE;

function command(args, environment) {
  const result = spawnSync(args[0], args.slice(1), { env: environment, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"] });
  return { status: typeof result.status === "number" ? result.status : 2, output: `${result.stdout || ""}${result.stderr || ""}` };
}

function unsupported(reason) {
  console.log(JSON.stringify({ status: "unsupported", reason }));
  process.exit(0);
}

if (command([codex, "--version"], process.env).status !== 0) {
  unsupported("codex executable is not available");
}
if (!cli) unsupported("set BAN_CODE_COMMENTS_EXECUTABLE to a built hook CLI for live smoke coverage");

const authPath = process.env.CODEX_SMOKE_AUTH_FILE || (process.env.CODEX_SMOKE_REUSE_AUTH === "1" ? path.join(process.env.HOME || "", ".codex", "auth.json") : "");
if (!authPath) unsupported("set CODEX_SMOKE_AUTH_FILE or CODEX_SMOKE_REUSE_AUTH=1 to opt into authenticated Codex coverage");
try {
  await fs.access(authPath);
} catch {
  unsupported(`Codex auth file is unavailable at ${authPath}`);
}

const home = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-codex-home-"));
const repository = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-codex-repo-"));
await fs.symlink(authPath, path.join(home, "auth.json"));
await fs.writeFile(path.join(repository, "main.go"), "package main\n\nfunc main() {}\n");
await fs.writeFile(path.join(repository, "README.md"), "# Smoke\n");
const environment = { ...process.env, CODEX_HOME: home, BAN_CODE_COMMENTS_EXECUTABLE: cli };

const marketplace = command([codex, "plugin", "marketplace", "add", root, "--json"], environment);
if (marketplace.status !== 0) throw new Error(`marketplace setup failed: ${marketplace.output}`);

function plugin(name, action) {
  const result = command([codex, "plugin", action, `${name}@ban-code-comments`, "--json"], environment);
  if (result.status !== 0) throw new Error(`${action} ${name} failed: ${result.output}`);
}

function runPrompt(prompt) {
  return command([codex, "exec", "--ephemeral", "--dangerously-bypass-hook-trust", "--dangerously-bypass-approvals-and-sandbox", "--skip-git-repo-check", "--cd", repository, prompt], environment);
}

function observation(caseName, passed, result) {
  return { case: caseName, status: passed ? "passed" : "failed", exit: result.status, evidence: result.output.replaceAll(/\s+/g, " ").slice(-240) };
}

const observations = [];
plugin("ban-code-comments-hard-block", "add");
let result = runPrompt("Use apply_patch once to add an ordinary Go comment containing the exact text // codex hard pre smoke to main.go. Do not use any other tool.");
observations.push(observation("hard-pre-denial", result.status === 0 && result.output.includes("PreToolUse Blocked") && !(await fs.readFile(path.join(repository, "main.go"), "utf8")).includes("codex hard pre smoke"), result));

result = runPrompt("Use apply_patch once to add a Markdown comment to README.md and a Go string literal containing // literal smoke to main.go. Do not use any other tool.");
const markdownAndLiteral = await Promise.all([
  fs.readFile(path.join(repository, "README.md"), "utf8"),
  fs.readFile(path.join(repository, "main.go"), "utf8"),
]);
observations.push(observation("markdown-and-literal-allowance", result.status === 0 && markdownAndLiteral[0].includes("<!--") && markdownAndLiteral[1].includes("// literal smoke"), result));

result = runPrompt("Use Bash once to append the exact line // codex opaque post smoke to main.go. Do not use apply_patch or any other tool.");
observations.push(observation("hard-opaque-post-audit", result.status === 0 && result.output.includes("PostToolUse Stopped") && (await fs.readFile(path.join(repository, "main.go"), "utf8")).includes("codex opaque post smoke"), result));

plugin("ban-code-comments-hard-block", "remove");
plugin("ban-code-comments-warn", "add");
await fs.writeFile(path.join(repository, "main.go"), "package main\n\nfunc main() {}\n");
result = runPrompt("Use apply_patch once to add an ordinary Go comment containing the exact text // codex warn smoke to main.go. Do not use any other tool.");
observations.push(observation("warn-pre-allowance", result.status === 0 && result.output.includes("codex warn smoke") && (await fs.readFile(path.join(repository, "main.go"), "utf8")).includes("codex warn smoke"), result));

const failed = observations.filter((observation) => observation.status !== "passed");
console.log(JSON.stringify({ status: failed.length === 0 ? "passed" : "failed", observations }));
if (failed.length > 0) process.exit(1);
