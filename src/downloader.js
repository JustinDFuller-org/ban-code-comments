import crypto from "node:crypto";
import * as fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import * as toolCache from "@actions/tool-cache";
import { verifyChecksum } from "./checksum.js";
import { assetName, releaseURLs, targetFor } from "./platform.js";

const TOOL_NAME = "ban-code-comments";

function cacheArchitecture(target) {
  return `${target.releaseOS}-${target.releaseArch}`;
}

function executableNameFor(target) {
  return target.releaseOS === "windows" ? "ban-code-comments.exe" : "ban-code-comments";
}

async function findNamedFile(root, filename) {
  const entries = await fs.readdir(root, { withFileTypes: true });
  for (const entry of entries) {
    const entryPath = path.join(root, entry.name);
    if (entry.isDirectory()) {
      const found = await findNamedFile(entryPath, filename);
      if (found) return found;
    } else if (entry.isFile() && entry.name === filename) {
      return entryPath;
    }
  }
  return null;
}

async function findExecutable(root, executableName = process.platform === "win32" ? "ban-code-comments.exe" : "ban-code-comments") {
  return findNamedFile(root, executableName);
}

async function fileDigest(filePath) {
  const contents = await fs.readFile(filePath);
  return crypto.createHash("sha256").update(contents).digest("hex");
}

async function verifyCachedExecutable(cachedRoot, archivePath, filename, target, cache, urls) {
  const checksumPath = await cache.downloadTool(urls.checksums);
  const checksums = await fs.readFile(checksumPath, "utf8");
  await verifyChecksum(archivePath, checksums, filename);

  const verificationDirectory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-cache-"));
  try {
    const extractedPath = target.archiveFormat === "zip"
      ? await cache.extractZip(archivePath, verificationDirectory)
      : await cache.extractTar(archivePath, verificationDirectory);
    const expectedExecutable = await findExecutable(extractedPath, executableNameFor(target));
    const cachedExecutable = await findExecutable(cachedRoot, executableNameFor(target));
    if (!expectedExecutable || !cachedExecutable) return null;
    const [expectedDigest, cachedDigest] = await Promise.all([
      fileDigest(expectedExecutable),
      fileDigest(cachedExecutable),
    ]);
    if (expectedDigest !== cachedDigest) {
      throw new Error(`cached executable does not match verified archive ${filename}`);
    }
    return cachedExecutable;
  } finally {
    await fs.rm(verificationDirectory, { recursive: true, force: true });
  }
}

async function downloadCLI(version, options = {}) {
  const cache = options.cache || toolCache;
  const target = options.target || targetFor();
  const urls = releaseURLs(version, target);
  const filename = assetName(version, target);
  const cacheArch = cacheArchitecture(target);
  const cached = cache.find(TOOL_NAME, version, cacheArch);
  if (cached) {
    const cachedArchive = await findNamedFile(cached, filename);
    if (cachedArchive) {
      try {
        const executable = await verifyCachedExecutable(cached, cachedArchive, filename, target, cache, urls);
        if (executable) return executable;
      } catch {
      }
    }
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
    const executable = await findExecutable(extractedPath, executableNameFor(target));
    if (!executable) throw new Error(`released archive does not contain ${executableNameFor(target)}`);
    if (target.releaseOS !== "windows") await fs.chmod(executable, 0o755);
    await fs.copyFile(archivePath, path.join(extractedPath, filename));
    const cachePath = await cache.cacheDir(extractedPath, TOOL_NAME, version, cacheArch);
    const cachedExecutable = await findExecutable(cachePath, executableNameFor(target));
    if (!cachedExecutable) throw new Error(`cached release does not contain ${path.basename(executable)}`);
    return cachedExecutable;
  } finally {
    await fs.rm(tempDirectory, { recursive: true, force: true });
  }
}

export { TOOL_NAME, downloadCLI, findExecutable };
