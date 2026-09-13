import { execFile } from "node:child_process";
import { mkdtemp, cp, mkdir, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import { evaluateHook } from "../src/hook.js";

const root = path.resolve(new URL("..", import.meta.url).pathname);

async function command(file, args, options = {}) {
  return new Promise((resolve) => {
    const child = execFile(file, args, { cwd: root, encoding: "utf8", ...options, input: undefined }, (error, stdout, stderr) => resolve({ code: error ? (Number.isInteger(error.code) ? error.code : 2) : 0, stdout, stderr }));
    if (options.input !== undefined) child.stdin.end(options.input);
  });
}

function normalizeResult(text) {
  const value = JSON.parse(text);
  const marker = "fixtures/";
  const findings = value.findings.map((item) => {
    const index = item.path.indexOf(marker);
    return { ...item, path: index >= 0 ? item.path.slice(index) : item.path };
  });
  return { findings, summary: value.summary };
}

function normalizeHook(value) {
  return JSON.parse(JSON.stringify(value));
}

async function run() {
  const workspace = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-differential-"));
  await mkdir(path.join(workspace, "fixtures"));
  await cp(path.join(root, "test", "fixtures"), path.join(workspace, "fixtures"), { recursive: true });
  await writeFile(path.join(workspace, "sample.go"), "package main\nvar value = 1 // finding\n");
  const goBinary = path.join(workspace, "ban-code-comments-go");
  const build = await command("go", ["build", "-o", goBinary, "./cmd/ban-code-comments"]);
  if (build.code !== 0) throw new Error(`could not build Go oracle: ${build.stderr}`);
  const fixturePath = path.join(workspace, "fixtures");
  const js = await command(process.execPath, [path.join(root, "bin", "ban-code-comments.js"), "--format", "json", fixturePath]);
  const go = await command(goBinary, ["--format", "json", fixturePath]);
  if (js.code !== go.code || JSON.stringify(normalizeResult(js.stdout)) !== JSON.stringify(normalizeResult(go.stdout))) throw new Error(`fixture parity mismatch: JavaScript ${js.code}, Go ${go.code}`);
  for (const args of [["--categories", "ordinary", fixturePath], ["--categories", "documentation", fixturePath], ["--categories", "header", fixturePath], ["--categories", "directive", fixturePath], ["--languages", "javascript,python", fixturePath], ["--include", "**/finding/**", "--exclude", "**/python/**", fixturePath]]) {
    const left = await command(process.execPath, [path.join(root, "bin", "ban-code-comments.js"), "--format", "json", ...args]);
    const right = await command(goBinary, ["--format", "json", ...args]);
    if (left.code !== right.code || JSON.stringify(normalizeResult(left.stdout)) !== JSON.stringify(normalizeResult(right.stdout))) throw new Error(`option parity mismatch for ${args.join(" ")}`);
  }
  for (const args of [["--format", "invalid", fixturePath], ["--languages", "unsupported", fixturePath], ["/missing/ban-code-comments-path"]]) {
    const left = await command(process.execPath, [path.join(root, "bin", "ban-code-comments.js"), ...args]);
    const right = await command(goBinary, args);
    if (left.code !== right.code) throw new Error(`status parity mismatch for ${args.join(" ")}: JavaScript ${left.code}, Go ${right.code}`);
  }
  for (const mode of ["hard", "warn"]) for (const tool of ["apply_patch", "edit", "write", "write_file", "file_write"]) {
    const toolInput = tool === "apply_patch" ? { patch: "*** Begin Patch\n*** Add File: mode.go\n+package main\n+// finding\n*** End Patch\n" } : { path: "sample.go", content: "package main\n// finding\n" };
    const event = { cwd: workspace, hook_event_name: "PreToolUse", tool_name: tool, tool_input: toolInput };
    const jsHook = normalizeHook(await evaluateHook(event, mode));
    const goHook = await command(goBinary, ["hook", "--mode", mode], { input: `${JSON.stringify(event)}\n` });
    const goHookValue = normalizeHook(JSON.parse(goHook.stdout));
    if (JSON.stringify(jsHook) !== JSON.stringify(goHookValue)) throw new Error(`hook parity mismatch for ${tool}/${mode}`);
  }
  process.stdout.write(`differential parity passed: fixture status ${js.code}, hook matrix complete\n`);
  return 0;
}

if (import.meta.url === `file://${process.argv[1]}`) run().catch((error) => { process.stderr.write(`${error.message}\n`); process.exitCode = 1; });

export { normalizeHook, normalizeResult, run };
