import assert from "node:assert/strict";
import { mkdtemp, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { evaluateClaudeHook, runClaudeHook } from "../src/claude-hook.js";

const event = (cwd, tool, toolInput) => ({ cwd, hook_event_name: "PreToolUse", tool_name: tool, tool_input: toolInput });

test("Claude hard mode denies Write and Edit findings with native nested output", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-claude-"));
  const write = await evaluateClaudeHook(event(cwd, "Write", { file_path: path.join(cwd, "main.go"), content: "package main\n// finding\n" }));
  assert.equal(write.hookSpecificOutput.permissionDecision, "deny");
  assert.match(write.hookSpecificOutput.permissionDecisionReason, /main\.go:2:1/);
  await writeFile(path.join(cwd, "main.go"), "package main\nvar value = 1\n");
  const edit = await evaluateClaudeHook(event(cwd, "Edit", { file_path: path.join(cwd, "main.go"), old_string: "var value = 1", new_string: "var value = 1 // finding" }));
  assert.equal(edit.hookSpecificOutput.permissionDecision, "deny");
});

test("Claude Edit honors first and all replacement semantics", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-claude-"));
  const filePath = path.join(cwd, "main.go");
  await writeFile(filePath, "package main\nvar one = 1\nvar two = 1\n");
  const first = await evaluateClaudeHook(event(cwd, "Edit", { file_path: filePath, old_string: "= 1", new_string: "= 1 // finding" }));
  assert.equal(first.hookSpecificOutput.permissionDecision, "deny");
  const all = await evaluateClaudeHook(event(cwd, "Edit", { file_path: filePath, old_string: "= 1", new_string: "= 1 // finding", replace_all: true }));
  assert.equal(all.hookSpecificOutput.permissionDecision, "deny");
  assert.match(all.hookSpecificOutput.permissionDecisionReason, /main\.go:2:13/);
});

test("Claude warn mode allows findings and unchanged findings do not respond", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-claude-"));
  const filePath = path.join(cwd, "main.go");
  await writeFile(filePath, "package main\n// legacy\nvar value = 1\n");
  const warn = await evaluateClaudeHook(event(cwd, "Edit", { file_path: filePath, old_string: "var value = 1", new_string: "var value = 2 // new" }), "warn");
  assert.equal(warn.hookSpecificOutput.permissionDecision, undefined);
  assert.equal(warn.hookSpecificOutput.additionalContext, warn.systemMessage);
  const unchanged = await evaluateClaudeHook(event(cwd, "Edit", { file_path: filePath, old_string: "var value = 1", new_string: "var value = 2" }));
  assert.deepEqual(unchanged, {});
});

test("Claude unsupported events, literals, Markdown, and unsupported files are no-ops", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-claude-"));
  for (const item of [
    { tool_name: "Bash", tool_input: { command: "printf '// finding' > main.go" } },
    { tool_name: "Write", tool_input: { file_path: path.join(cwd, "main.go"), content: "package main\nvar text = \"// literal\"\n" } },
    { tool_name: "Write", tool_input: { file_path: path.join(cwd, "README.md"), content: "# docs\n<!-- allowed -->\n" } },
    { tool_name: "Write", tool_input: { file_path: path.join(cwd, "data.bin"), content: "// unsupported\n" } },
  ]) assert.deepEqual(await evaluateClaudeHook({ ...event(cwd, item.tool_name, item.tool_input), hook_event_name: item.tool_name === "Bash" ? "PreToolUse" : "PostToolUse" }), {});
});

test("Claude malformed proposals return mode-specific operational responses", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-claude-"));
  const hard = await evaluateClaudeHook(event(cwd, "Edit", { file_path: path.join(cwd, "main.go"), old_string: "missing", new_string: "new" }));
  assert.equal(hard.hookSpecificOutput.permissionDecision, "deny");
  const warn = await evaluateClaudeHook(event(cwd, "Edit", { file_path: path.join(cwd, "main.go"), old_string: "missing", new_string: "new" }), "warn");
  assert.match(warn.systemMessage, /proposal_unreadable/);
  let output = "";
  await runClaudeHook("not json", "hard", { write: (value) => { output += value; } });
  assert.match(output, /invalid_event/);
});
