import assert from "node:assert/strict";
import { mkdtemp, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { Readable } from "node:stream";
import test from "node:test";
import { applyPatch, evaluateHook, parsePatch, reconstruct, resolve, runHook } from "../src/hook.js";

const event = (cwd, tool, toolInput) => ({ cwd, hook_event_name: "PreToolUse", tool_name: tool, tool_input: toolInput });

test("hard mode blocks a newly introduced comment for every traditional write tool", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-hook-"));
  const inputs = {
    apply_patch: { patch: "*** Begin Patch\n*** Add File: main.go\n+package main\n+// finding\n*** End Patch\n" },
    edit: { path: "main.go", content: "package main\n// finding\n" },
    write: { path: "main.go", content: "package main\n// finding\n" },
    write_file: { path: "main.go", content: "package main\n// finding\n" },
    file_write: { path: "main.go", content: "package main\n// finding\n" },
  };
  for (const [tool, input] of Object.entries(inputs)) {
    const response = await evaluateHook(event(cwd, tool, input), "hard");
    assert.equal(response.decision, "block", tool);
    assert.match(response.reason, /main\.go:2:1/);
  }
});

test("warn mode allows with model-visible guidance and legacy findings are subtracted", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-hook-"));
  await writeFile(path.join(cwd, "main.go"), "package main\n// legacy\nvar value = 1\n");
  const patch = "*** Begin Patch\n*** Update File: main.go\n@@\n package main\n // legacy\n-var value = 1\n+var value = 2\n*** End Patch\n";
  assert.deepEqual(await evaluateHook(event(cwd, "apply_patch", { patch }), "hard"), {});
  const response = await evaluateHook(event(cwd, "write_file", { path: "main.go", content: "package main\n// new\n" }), "warn");
  assert.equal(response.decision, undefined);
  assert.equal(response.hookSpecificOutput.additionalContext, response.systemMessage);
  assert.match(response.systemMessage, /README/);
});

test("literals, Markdown, unsupported tools, and non-pre events are no-ops", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-hook-"));
  for (const [tool, input] of [["write_file", { path: "main.go", content: "package main\nvar text = \"// literal\"\n" }], ["write_file", { path: "README.md", content: "# docs\n<!-- allowed -->\n" }], ["write_file", { path: "data.bin", content: "// unsupported\n" }], ["Bash", { path: "main.go", content: "package main\n// finding\n" }]]) assert.deepEqual(await evaluateHook(event(cwd, tool, input), "hard"), {});
  assert.deepEqual(await evaluateHook({ ...event(cwd, "write_file", { path: "main.go", content: "// finding\n" }), hook_event_name: "PostToolUse" }, "hard"), {});
});

test("hook protocol returns mode-specific operational responses", async () => {
  let output = "";
  const stream = { write: (value) => { output += value; } };
  assert.equal(await runHook("not json", "hard", stream), 0);
  let response = JSON.parse(output);
  assert.equal(response.decision, undefined);
  assert.match(response.systemMessage, /invalid_event/);
  output = "";
  await runHook(JSON.stringify({}), "warn", stream);
  assert.equal(output, "{}\n");
  output = "";
  await runHook({}, "hard", stream);
  assert.equal(output, "{}\n");
  output = "";
  await runHook(JSON.stringify({}), "invalid", stream);
  response = JSON.parse(output);
  assert.equal(response.decision, undefined);
  assert.match(response.systemMessage, /invalid_mode/);
});

test("Codex hook runner reads stdin streams before evaluating events", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-hook-"));
  const eventText = JSON.stringify(event(cwd, "write_file", { path: "main.go", content: "package main\n// finding\n" }));
  let output = "";
  await runHook(Readable.from([eventText]), "hard", { write: (value) => { output += value; } });
  const response = JSON.parse(output);
  assert.equal(response.decision, "block");
  assert.match(response.reason, /main\.go:2:1/);
});

test("hook reconstruction and path helpers reject malformed or unsafe proposals", () => {
  assert.equal(parsePatch("*** Begin Patch\n*** Delete File: main.go\n*** End Patch\n")[0].delete, true);
  assert.throws(() => parsePatch("*** Begin Patch\n*** Add File: main.go\n+// finding\n"), /end marker/);
  assert.equal(reconstruct({ filePath: "main.go", newContent: "package main\n" })[0].source, "package main\n");
  assert.deepEqual(reconstruct({ path: "main.go", oldString: "old", newString: "new" })[0], { path: "main.go", oldText: "old", newText: "new" });
  for (const input of [undefined, [], {}, { path: "main.go" }, { path: "main.go", content: 1 }, { path: "main.go", old_string: "", new_string: "// finding" }, { path: "main.go", old_string: 1, new_string: "new" }, { path: "main.go", old_string: "old", new_string: 1 }, { path: "main.go", old_string: 1, oldString: "old", new_string: "new" }, { path: "main.go", old_string: "old", new_string: "new", newString: "new" }, { content: "new", newContent: "new" }, { path: "main.go", file_path: "other.go", content: "new" }, { patch: "*** End Patch\n", content: "new" }]) assert.throws(() => reconstruct(input));
  const root = "/tmp/ban-code-comments-hook-root";
  assert.match(resolve(root, "a/main.go"), /main\.go$/);
  assert.throws(() => resolve(root, "../escape"), /escapes workspace/);
  assert.equal(applyPatch("one\ntwo\nthree\n", ["@@", " one", "-two", "+changed"]), "one\nchanged\nthree\n");
  assert.throws(() => applyPatch("one\n", ["@@", "-missing", "+new"]), /context/);
});

test("rename and delete proposals do not report removed legacy findings", async () => {
  const cwd = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-hook-"));
  await writeFile(path.join(cwd, "old.go"), "package main\n// legacy\n");
  const moved = await evaluateHook(event(cwd, "apply_patch", { patch: "*** Begin Patch\n*** Update File: old.go\n*** Move to: new.go\n@@\n package main\n // legacy\n*** End Patch\n" }), "hard");
  assert.deepEqual(moved, {});
  const deleted = await evaluateHook(event(cwd, "apply_patch", { patch: "*** Begin Patch\n*** Delete File: old.go\n*** End Patch\n" }), "hard");
  assert.deepEqual(deleted, {});
});
