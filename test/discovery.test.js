import assert from "node:assert/strict";
import { mkdtemp, mkdir, symlink, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { discover } from "../src/discovery.js";

test("discovery applies language, include, exclude, fixed-directory, and symlink filters", async () => {
  const root = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-discovery-"));
  await mkdir(path.join(root, "src"), { recursive: true });
  await mkdir(path.join(root, "node_modules"), { recursive: true });
  await writeFile(path.join(root, "src", "kept.js"), "const value = 1;\n");
  await writeFile(path.join(root, "src", "excluded.js"), "const value = 2;\n");
  await writeFile(path.join(root, "README.md"), "# docs\n");
  await writeFile(path.join(root, "node_modules", "ignored.js"), "const value = 3;\n");
  await symlink(path.join(root, "src", "kept.js"), path.join(root, "linked.js"));
  const selected = await discover(["."], { cwd: root, languages: new Set(["javascript"]), includes: ["src/**"], excludes: ["**/excluded.js"] });
  assert.deepEqual(selected.candidates.map((item) => item.relative), ["src/kept.js"]);
  assert.ok(selected.diagnostics.some((item) => item.reason === "excluded"));
  assert.ok(selected.diagnostics.some((item) => item.reason === "symbolic link"));
  assert.ok(selected.diagnostics.some((item) => item.reason === "unsupported file"));
});

test("discovery reports missing explicit paths and deduplicates files", async () => {
  const root = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-discovery-"));
  await writeFile(path.join(root, "fixture.py"), "value = 1\n");
  const selected = await discover(["fixture.py", "./fixture.py"], { cwd: root });
  assert.equal(selected.candidates.length, 1);
  await assert.rejects(discover(["missing.py"], { cwd: root }), /no such file|ENOENT/i);
});
