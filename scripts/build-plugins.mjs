import { execFile } from "node:child_process";
import * as fs from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);
const root = fileURLToPath(new URL("..", import.meta.url));

for (const [entry, destination] of [
  ["src/launcher-entry.js", "plugins/ban-code-comments-hard-block/bin/launcher.js"],
  ["src/launcher-entry.js", "plugins/ban-code-comments-warn/bin/launcher.js"],
  ["src/claude-launcher-entry.js", "plugins/ban-code-comments-claude-hard-block/bin/launcher.js"],
  ["src/claude-launcher-entry.js", "plugins/ban-code-comments-claude-warn/bin/launcher.js"],
]) {
  const output = path.join(root, destination);
  await fs.mkdir(path.dirname(output), { recursive: true });
  const bundleDirectory = path.join(path.dirname(output), `.bundle-${path.basename(output)}`);
  await fs.rm(output, { recursive: true, force: true });
  await fs.rm(bundleDirectory, { recursive: true, force: true });
  await execFileAsync("npx", ["--no", "ncc", "build", path.join(root, entry), "--minify", "--license", path.join(root, "LICENSE"), "-o", bundleDirectory], { cwd: root });
  await fs.rename(path.join(bundleDirectory, "index.js"), output);
  await fs.rm(bundleDirectory, { recursive: true, force: true });
}
