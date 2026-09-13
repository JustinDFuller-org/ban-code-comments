import assert from "node:assert/strict";
import { mkdtemp, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { runCLI } from "../src/cli-runner.js";

function streams() {
  const value = { stdout: "", stderr: "" };
  return { value, stdout: { write: (chunk) => { value.stdout += chunk; } }, stderr: { write: (chunk) => { value.stderr += chunk; } } };
}

test("CLI runner returns policy and clean statuses with injected streams", async () => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-cli-"));
  const file = path.join(directory, "fixture.js");
  await writeFile(file, "const value = 1; // finding\n");
  const finding = streams();
  assert.equal(await runCLI([file], finding), 1);
  assert.match(finding.value.stdout, /\"findings\":\[/);
  const clean = streams();
  assert.equal(await runCLI(["--categories", "directive", file], clean), 0);
  assert.match(clean.value.stdout, /\"findings\":\[\]/);
});

test("CLI runner maps invalid options and unreadable paths to status 2", async () => {
  const invalid = streams();
  assert.equal(await runCLI(["--format", "yaml"], invalid), 2);
  assert.match(invalid.value.stderr, /unsupported format/);
  const missing = streams();
  assert.equal(await runCLI(["/tmp/ban-code-comments-no-such-path"], missing), 2);
  assert.match(missing.value.stderr, /no such file|ENOENT/i);
});
