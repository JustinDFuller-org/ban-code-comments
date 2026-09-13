import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtemp, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";

const root = process.cwd();
const launchers = [
  ["Codex source", "src/launcher-entry.js", "codex", "hard"],
  ["Codex hard bundle", "plugins/ban-code-comments-hard-block/bin/launcher.js", "codex", "hard"],
  ["Codex warn bundle", "plugins/ban-code-comments-warn/bin/launcher.js", "codex", "warn"],
  ["Claude source", "src/claude-launcher-entry.js", "claude", "hard"],
  ["Claude hard bundle", "plugins/ban-code-comments-claude-hard-block/bin/launcher.js", "claude", "hard"],
  ["Claude warn bundle", "plugins/ban-code-comments-claude-warn/bin/launcher.js", "claude", "warn"],
];

function runLauncher(relativePath, mode, event) {
  const result = spawnSync(process.execPath, [path.join(root, relativePath), "--mode", mode], { cwd: root, input: `${JSON.stringify(event)}\n`, encoding: "utf8" });
  assert.equal(result.status, 0, result.stderr);
  return JSON.parse(result.stdout);
}

function assertFindingResponse(kind, mode, response) {
  if (kind === "codex") {
    if (mode === "hard") assert.equal(response.decision, "block");
    else assert.equal(response.decision, undefined);
    assert.match(mode === "hard" ? response.reason : response.systemMessage, /finding/);
    return;
  }
  if (mode === "hard") assert.equal(response.hookSpecificOutput.permissionDecision, "deny");
  else assert.equal(response.hookSpecificOutput.permissionDecision, undefined);
  assert.match(mode === "hard" ? response.hookSpecificOutput.permissionDecisionReason : response.hookSpecificOutput.additionalContext, /finding/);
}

test("source and bundled launchers decode stdin and preserve enforcement modes", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-launcher-"));
  await writeFile(path.join(cwd, "main.go"), "package main\n");
  const finding = { cwd, hook_event_name: "PreToolUse", tool_name: "write_file", tool_input: { path: "main.go", content: "package main\n// finding\n" } };
  const clean = { ...finding, tool_input: { path: "main.go", content: "package main\n" } };
  for (const [, relativePath, kind, mode] of launchers) {
    const event = kind === "codex" ? finding : { ...finding, tool_name: "Write", tool_input: { file_path: path.join(cwd, "main.go"), content: "package main\n// finding\n" } };
    assertFindingResponse(kind, mode, runLauncher(relativePath, mode, event));
    const cleanEvent = kind === "codex" ? clean : { ...event, tool_input: { file_path: path.join(cwd, "main.go"), content: "package main\n" } };
    assert.deepEqual(runLauncher(relativePath, mode, cleanEvent), {});
  }
});

test("all bundled launchers fail open with diagnostics for malformed input", () => {
  for (const [, relativePath, , mode] of launchers) {
    const result = spawnSync(process.execPath, [path.join(root, relativePath), "--mode", mode], { cwd: root, input: "not json\n", encoding: "utf8" });
    assert.equal(result.status, 0, result.stderr);
    const response = JSON.parse(result.stdout);
    assert.equal(response.decision, undefined);
    assert.equal(response.hookSpecificOutput?.permissionDecision, undefined);
    assert.match(response.systemMessage || response.hookSpecificOutput?.additionalContext, /invalid_event/);
  }
});
