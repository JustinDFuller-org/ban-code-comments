import crypto from "node:crypto";
import { execFile } from "node:child_process";
import * as fs from "node:fs/promises";
import fsSync from "node:fs";
import os from "node:os";
import path from "node:path";
import { promisify } from "node:util";
import { downloadCLI, TOOL_NAME } from "./downloader.js";
import { CLI_VERSION } from "./version.js";
import { targetFor } from "./platform.js";

const execFileAsync = promisify(execFile);

function pluginDataRoot(environment = process.env) {
  if (environment.BAN_CODE_COMMENTS_PLUGIN_DATA) return environment.BAN_CODE_COMMENTS_PLUGIN_DATA;
  if (environment.CODEX_PLUGIN_DATA) return environment.CODEX_PLUGIN_DATA;
  if (environment.CODEX_HOME) return path.join(environment.CODEX_HOME, "plugins", "data", "ban-code-comments");
  return path.join(os.tmpdir(), "ban-code-comments-plugin-data");
}

function cacheRoot(root, version, target) {
  return path.join(root, TOOL_NAME, version, `${target.releaseOS}-${target.releaseArch}`);
}

function createPluginCache(root, options = {}) {
  const fetchImpl = options.fetch || globalThis.fetch;
  const target = options.target || targetFor();
  if (typeof fetchImpl !== "function") throw new Error("Node fetch is unavailable; use Node 18 or newer");
  return {
    find(_name, version, _architecture) {
      const directory = cacheRoot(root, version, target);
      return fsSync.existsSync(directory) ? directory : undefined;
    },
    async downloadTool(url) {
      const response = await fetchImpl(url);
      if (!response.ok) throw new Error(`download failed with HTTP ${response.status}: ${url}`);
      const contents = Buffer.from(await response.arrayBuffer());
      const digest = crypto.createHash("sha256").update(contents).digest("hex");
      const temporaryDirectory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-plugin-download-"));
      const filename = path.basename(new URL(url).pathname) || `download-${digest}`;
      const destination = path.join(temporaryDirectory, filename);
      await fs.writeFile(destination, contents, { mode: 0o600 });
      return destination;
    },
    async extractTar(archivePath, destination) {
      await execFileAsync("tar", ["-xzf", archivePath, "-C", destination]);
      return destination;
    },
    async extractZip(archivePath, destination) {
      await execFileAsync(process.platform === "win32" ? "powershell.exe" : "unzip", process.platform === "win32"
        ? ["-NoProfile", "-NonInteractive", "-Command", `Expand-Archive -LiteralPath '${archivePath.replaceAll("'", "''")}' -DestinationPath '${destination.replaceAll("'", "''")}' -Force`]
        : ["-q", archivePath, "-d", destination]);
      return destination;
    },
    async cacheDir(source, name, version, architecture) {
      const destination = path.join(root, name, version, architecture);
      const lock = `${destination}.lock`;
      await fs.mkdir(path.dirname(destination), { recursive: true, mode: 0o700 });
      const deadline = Date.now() + 30000;
      let acquired = false;
      while (!acquired) {
        try {
          await fs.mkdir(lock, { recursive: false, mode: 0o700 });
          acquired = true;
        } catch (error) {
          if (error?.code !== "EEXIST") throw error;
          try {
            const details = await fs.stat(lock);
            if (Date.now() - details.mtimeMs > 60000) await fs.rm(lock, { recursive: true, force: true });
          } catch (statError) {
            if (statError?.code !== "ENOENT") throw statError;
          }
          if (Date.now() >= deadline) throw new Error(`timed out waiting for plugin cache lock ${destination}`);
          await new Promise((resolve) => setTimeout(resolve, 50));
        }
      }
      try {
        await fs.rm(destination, { recursive: true, force: true });
        await fs.cp(source, destination, { recursive: true, force: true, verbatimSymlinks: true });
        return destination;
      } finally {
        await fs.rm(lock, { recursive: true, force: true });
      }
    },
  };
}

async function downloadPluginCLI(options = {}) {
  const environment = options.environment || process.env;
  const override = options.executable || environment.BAN_CODE_COMMENTS_EXECUTABLE || environment.BAN_CODE_COMMENTS_CLI_PATH;
  if (override) return override;
  const target = options.target || targetFor();
  const root = options.root || pluginDataRoot(environment);
  return downloadCLI(CLI_VERSION, {
    target,
    cache: options.cache || createPluginCache(root, { target, fetch: options.fetch }),
  });
}

export { createPluginCache, downloadPluginCLI, pluginDataRoot };
