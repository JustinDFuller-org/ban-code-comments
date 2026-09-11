import * as fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import * as toolCache from "@actions/tool-cache";
import { verifyChecksum } from "./checksum.js";
import { assetName, releaseURLs, targetFor } from "./platform.js";

const TOOL_NAME = "ban-code-comments";

async function findExecutable(root, executableName = process.platform === "win32" ? "ban-code-comments.exe" : "ban-code-comments") {
  const entries = await fs.readdir(root, { withFileTypes: true });
  for (const entry of entries) {
    const entryPath = path.join(root, entry.name);
    if (entry.isDirectory()) {
      const found = await findExecutable(entryPath, executableName);
      if (found) return found;
    } else if (entry.isFile() && entry.name === executableName) {
      return entryPath;
    }
  }
  return null;
}

async function downloadCLI(version, options = {}) {
  const cache = options.cache || toolCache;
  const target = options.target || targetFor();
  const urls = releaseURLs(version, target);
  const filename = assetName(version, target);
  const cached = cache.find(TOOL_NAME, version, target.releaseArch);
  if (cached) {
    const executable = await findExecutable(cached);
    if (executable) return executable;
  }

  const tempDirectory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-"));
  try {
    const archivePath = await cache.downloadTool(urls.archive);
    const checksumPath = await cache.downloadTool(urls.checksums);
    const checksums = await fs.readFile(checksumPath, "utf8");
    await verifyChecksum(archivePath, checksums, filename);

    const extractedPath = target.archiveFormat === "zip"
      ? await cache.extractZip(archivePath, tempDirectory)
      : await cache.extractTar(archivePath, tempDirectory);
    const executable = await findExecutable(extractedPath);
    if (!executable) throw new Error(`released archive does not contain ${path.basename("ban-code-comments")}`);
    if (target.releaseOS !== "windows") await fs.chmod(executable, 0o755);
    const cachePath = await cache.cacheDir(extractedPath, TOOL_NAME, version, target.releaseArch);
    const cachedExecutable = await findExecutable(cachePath);
    if (!cachedExecutable) throw new Error(`cached release does not contain ${path.basename(executable)}`);
    return cachedExecutable;
  } finally {
    await fs.rm(tempDirectory, { recursive: true, force: true });
  }
}

export { TOOL_NAME, downloadCLI, findExecutable };
