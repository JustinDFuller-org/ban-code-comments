import assert from "node:assert/strict";
import { mkdtemp, mkdir, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { execFile } from "node:child_process";
import { promisify } from "node:util";
import test from "node:test";
import { discover } from "../src/discovery.js";

const exec = promisify(execFile);

test("default discovery resolves paths from the repository root when cwd is nested", async () => {
  const root = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-git-"));
  await mkdir(path.join(root, "nested"), { recursive: true });
  await writeFile(path.join(root, ".gitignore"), "ignored.js\n");
  await writeFile(path.join(root, "nested", "kept.js"), "const value = 1;\n");
  await writeFile(path.join(root, "ignored.js"), "const value = 2; // ignored\n");
  await exec("git", ["init", "-q", root]);
  const discovered = await discover([], { cwd: path.join(root, "nested") });
  assert.deepEqual(discovered.candidates.map((item) => item.relative), ["kept.js"]);
});
