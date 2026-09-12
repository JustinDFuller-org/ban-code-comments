import assert from "node:assert/strict";
import * as fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import crypto from "node:crypto";
import { expectedChecksum, verifyChecksum } from "../src/checksum.js";
import { downloadCLI, findExecutable } from "../src/downloader.js";
import { createPluginCache, downloadPluginCLI, pluginDataRoot } from "../src/plugin-cache.js";
import { cliArguments, parseBoolean } from "../src/inputs.js";
import { assetName, releaseURLs, targetFor } from "../src/platform.js";
import { runCLI } from "../src/runner.js";
import { main as hookMain, runHook } from "../src/hook-launcher.js";
import { main as actionMain } from "../src/index.js";

test("maps supported runner targets to release coordinates", () => {
  assert.deepEqual(targetFor("linux", "x64"), { archiveFormat: "tar.gz", releaseOS: "linux", releaseArch: "amd64" });
  assert.deepEqual(targetFor("linux", "arm64"), { archiveFormat: "tar.gz", releaseOS: "linux", releaseArch: "arm64" });
  assert.deepEqual(targetFor("darwin", "x64"), { archiveFormat: "tar.gz", releaseOS: "darwin", releaseArch: "amd64" });
  assert.deepEqual(targetFor("darwin", "arm64"), { archiveFormat: "tar.gz", releaseOS: "darwin", releaseArch: "arm64" });
  assert.deepEqual(targetFor("win32", "x64"), { archiveFormat: "zip", releaseOS: "windows", releaseArch: "amd64" });
  assert.throws(() => targetFor("win32", "arm64"), /unsupported runner/);
  assert.equal(assetName("1.0.0", targetFor("linux", "x64")), "ban-code-comments_1.0.0_linux_amd64.tar.gz");
  assert.deepEqual(releaseURLs("1.0.0", targetFor("linux", "x64")), {
    archive: "https://github.com/JustinDFuller-org/ban-code-comments/releases/download/v1.0.0/ban-code-comments_1.0.0_linux_amd64.tar.gz",
    checksums: "https://github.com/JustinDFuller-org/ban-code-comments/releases/download/v1.0.0/checksums.txt",
  });
});

test("maps action inputs to safe CLI arguments", () => {
  assert.deepEqual(cliArguments({
    paths: "src folder\nfile.go",
    languages: "go,python\nrust",
    categories: "ordinary\ndirective",
    include: "src/**/*.go\nfolder with spaces/**",
    exclude: "vendor/**",
    format: "text",
    debug: "true",
  }), [
    "--languages", "go,python", "--languages", "rust",
    "--categories", "ordinary,directive",
    "--include", "src/**/*.go", "--include", "folder with spaces/**",
    "--exclude", "vendor/**", "--format", "text", "--debug",
    "src folder", "file.go",
  ]);
  assert.deepEqual(cliArguments({ format: "json", debug: "false" }), ["--format", "json", "."]);
  assert.throws(() => parseBoolean("sometimes"), /invalid boolean/);
});

test("parses and verifies release checksums", async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-test-"));
  const archive = path.join(directory, "archive.tar.gz");
  await fs.writeFile(archive, "release archive");
  const expected = crypto.createHash("sha256").update("release archive").digest("hex");
  const checksums = `${expected}  archive.tar.gz\n`;
  assert.equal(expectedChecksum(checksums, "archive.tar.gz"), expected);
  assert.equal(await verifyChecksum(archive, checksums, "archive.tar.gz"), expected);
  await assert.rejects(verifyChecksum(archive, `${"0".repeat(64)}  archive.tar.gz`, "archive.tar.gz"), /checksum mismatch/);
  assert.throws(() => expectedChecksum("not a checksum", "archive.tar.gz"), /checksum entry not found/);
  assert.throws(() => expectedChecksum(`${"z".repeat(64)}  archive.tar.gz`, "archive.tar.gz"), /checksum entry not found/);
});

test("finds executables inside wrapped release archives and cache directories", async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-test-"));
  const unixExecutable = path.join(directory, "ban-code-comments_1.0.0_linux_amd64", "ban-code-comments");
  const windowsExecutable = path.join(directory, "ban-code-comments_1.0.0_windows_amd64", "ban-code-comments.exe");
  await fs.mkdir(path.dirname(unixExecutable), { recursive: true });
  await fs.mkdir(path.dirname(windowsExecutable), { recursive: true });
  await fs.writeFile(unixExecutable, "binary");
  await fs.writeFile(windowsExecutable, "binary");
  assert.equal(await findExecutable(directory), unixExecutable);
  assert.equal(await findExecutable(directory, "ban-code-comments.exe"), windowsExecutable);
});

test("uses a cached verified executable without downloading the archive again", async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-cache-test-"));
  const executable = path.join(directory, "ban-code-comments");
  const archive = path.join(directory, "ban-code-comments_1.0.0_linux_amd64.tar.gz");
  const checksums = path.join(directory, "checksums.txt");
  await fs.writeFile(executable, "binary");
  await fs.writeFile(archive, "binary");
  const digest = crypto.createHash("sha256").update("binary").digest("hex");
  await fs.writeFile(checksums, `${digest}  ${path.basename(archive)}\n`);
  let archiveDownloads = 0;
  const result = await downloadCLI("1.0.0", {
    target: targetFor("linux", "x64"),
    cache: {
      find: () => directory,
      downloadTool: async () => {
        archiveDownloads += 1;
        return checksums;
      },
      extractTar: async (_archivePath, verificationDirectory) => {
        await fs.writeFile(path.join(verificationDirectory, "ban-code-comments"), "binary");
        return verificationDirectory;
      },
    },
  });
  assert.equal(result, executable);
  assert.equal(archiveDownloads, 1);
});

test("verifies a cache miss before extraction and caching", async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-download-test-"));
  const archive = path.join(directory, "archive.tar.gz");
  const checksums = path.join(directory, "checksums.txt");
  const archiveContents = "verified release archive";
  const digest = crypto.createHash("sha256").update(archiveContents).digest("hex");
  await fs.writeFile(archive, archiveContents);
  await fs.writeFile(checksums, `${digest}  ban-code-comments_1.0.0_linux_amd64.tar.gz\n`);
  const calls = [];
  const cachedExecutable = path.join(directory, "cached", "ban-code-comments");
  const result = await downloadCLI("1.0.0", {
    target: targetFor("linux", "x64"),
    cache: {
      find: () => undefined,
      downloadTool: async (url) => {
        calls.push(url.endsWith("checksums.txt") ? "download-checksums" : "download-archive");
        return url.endsWith("checksums.txt") ? checksums : archive;
      },
      extractTar: async () => {
        calls.push("extract");
        await fs.writeFile(path.join(directory, "ban-code-comments"), "binary");
        return directory;
      },
      cacheDir: async () => {
        calls.push("cache");
        await fs.mkdir(path.dirname(cachedExecutable), { recursive: true });
        await fs.writeFile(cachedExecutable, "binary");
        return path.dirname(cachedExecutable);
      },
    },
  });
  assert.equal(result, cachedExecutable);
  assert.deepEqual(calls, ["download-archive", "download-checksums", "extract", "cache"]);
});

test("downloads and discovers the Windows executable", async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-windows-download-test-"));
  const archive = path.join(directory, "archive.zip");
  const checksums = path.join(directory, "checksums.txt");
  const archiveContents = "verified Windows release archive";
  const filename = "ban-code-comments_1.0.0_windows_amd64.zip";
  const digest = crypto.createHash("sha256").update(archiveContents).digest("hex");
  await fs.writeFile(archive, archiveContents);
  await fs.writeFile(checksums, `${digest}  ${filename}\n`);
  const cachedDirectory = path.join(directory, "cached");
  const result = await downloadCLI("1.0.0", {
    target: targetFor("win32", "x64"),
    cache: {
      find: () => undefined,
      downloadTool: async (url) => url.endsWith("checksums.txt") ? checksums : archive,
      extractZip: async (_archivePath, extractionDirectory) => {
        const executable = path.join(extractionDirectory, "ban-code-comments_1.0.0_windows_amd64", "ban-code-comments.exe");
        await fs.mkdir(path.dirname(executable), { recursive: true });
        await fs.writeFile(executable, "binary");
        return extractionDirectory;
      },
      cacheDir: async () => {
        const executable = path.join(cachedDirectory, "ban-code-comments_1.0.0_windows_amd64", "ban-code-comments.exe");
        await fs.mkdir(path.dirname(executable), { recursive: true });
        await fs.writeFile(executable, "binary");
        return cachedDirectory;
      },
    },
  });
  assert.equal(result, path.join(cachedDirectory, "ban-code-comments_1.0.0_windows_amd64", "ban-code-comments.exe"));
});

test("redownloads when a cached executable differs from its verified archive", async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-cache-tamper-test-"));
  const cachedExecutable = path.join(directory, "ban-code-comments");
  const cachedArchive = path.join(directory, "ban-code-comments_1.0.0_linux_amd64.tar.gz");
  const checksums = path.join(directory, "checksums.txt");
  const freshDirectory = path.join(directory, "fresh");
  await fs.writeFile(cachedExecutable, "tampered");
  await fs.writeFile(cachedArchive, "binary");
  const digest = crypto.createHash("sha256").update("binary").digest("hex");
  await fs.writeFile(checksums, `${digest}  ${path.basename(cachedArchive)}\n`);
  const downloads = [];
  const result = await downloadCLI("1.0.0", {
    target: targetFor("linux", "x64"),
    cache: {
      find: () => directory,
      downloadTool: async (url) => {
        downloads.push(url.endsWith("checksums.txt") ? "checksums" : "archive");
        return url.endsWith("checksums.txt") ? checksums : cachedArchive;
      },
      extractTar: async (_archivePath, extractionDirectory) => {
        await fs.writeFile(path.join(extractionDirectory, "ban-code-comments"), "binary");
        return extractionDirectory;
      },
      cacheDir: async () => {
        await fs.mkdir(freshDirectory, { recursive: true });
        await fs.writeFile(path.join(freshDirectory, "ban-code-comments"), "binary");
        return freshDirectory;
      },
    },
  });
  assert.equal(result, path.join(freshDirectory, "ban-code-comments"));
  assert.deepEqual(downloads, ["checksums", "archive", "checksums"]);
});

test("rejects a mismatched cache miss before extraction", async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-download-failure-test-"));
  const archive = path.join(directory, "archive.tar.gz");
  const checksums = path.join(directory, "checksums.txt");
  await fs.writeFile(archive, "tampered release archive");
  await fs.writeFile(checksums, `${"0".repeat(64)}  ban-code-comments_1.0.0_linux_amd64.tar.gz\n`);
  let extracted = false;
  await assert.rejects(downloadCLI("1.0.0", {
    target: targetFor("linux", "x64"),
    cache: {
      find: () => undefined,
      downloadTool: async (url) => url.endsWith("checksums.txt") ? checksums : archive,
      extractTar: async () => {
        extracted = true;
        return directory;
      },
    },
  }), /checksum mismatch/);
  assert.equal(extracted, false);
});

test("preserves CLI exit statuses", async () => {
  for (const code of [0, 1, 2]) {
    const actual = await runCLI(process.execPath, ["-e", `process.exit(${code})`]);
    assert.equal(actual, code);
  }
});

test("reports subprocess launch failures while returning status 2", async () => {
  const errors = [];
  const originalError = console.error;
  console.error = (...args) => errors.push(args.join(" "));
  try {
    assert.equal(await runCLI(path.join(os.tmpdir(), "ban-code-comments-missing-executable"), []), 2);
  } finally {
    console.error = originalError;
  }
  assert.match(errors.join("\n"), /ENOENT|no such file|not found/);
});

test("keeps plugin data isolated and supports a local executable override", async () => {
  assert.equal(pluginDataRoot({ BAN_CODE_COMMENTS_PLUGIN_DATA: "/tmp/plugin-data" }), "/tmp/plugin-data");
  assert.equal(pluginDataRoot({ CODEX_PLUGIN_DATA: "/tmp/codex-plugin-data" }), "/tmp/codex-plugin-data");
  assert.equal(pluginDataRoot({ CODEX_HOME: "/tmp/codex-home" }), path.join("/tmp/codex-home", "plugins", "data", "ban-code-comments"));
  assert.equal(await downloadPluginCLI({ executable: process.execPath }), process.execPath);
});

test("serializes concurrent plugin cache writes", async () => {
  const root = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-plugin-cache-test-"));
  const source = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-plugin-source-"));
  await fs.writeFile(path.join(source, "ban-code-comments"), "binary");
  const cache = createPluginCache(root, { target: targetFor("darwin", "arm64"), fetch: async () => ({ ok: true }) });
  const destinations = await Promise.all([
    cache.cacheDir(source, "ban-code-comments", "1.0.0", "darwin-arm64"),
    cache.cacheDir(source, "ban-code-comments", "1.0.0", "darwin-arm64"),
  ]);
  assert.deepEqual(destinations, [destinations[0], destinations[0]]);
  assert.equal(await fs.readFile(path.join(destinations[0], "ban-code-comments"), "utf8"), "binary");
});

test("reports invalid hook launcher modes and runs the configured executable", async () => {
  assert.equal(await hookMain("invalid"), 0);
  assert.equal(await runHook("warn", { executable: process.execPath }), 1);
});

test("maps action inputs through the executable runner", async () => {
  const calls = [];
  const code = await actionMain({
    core: { getInput: (name) => ({ paths: "src", format: "text" }[name] || "") },
    download: async (version) => { calls.push(["download", version]); return "/tmp/ban-code-comments"; },
    run: async (...args) => { calls.push(["run", ...args]); return 0; },
  });
  assert.equal(code, 0);
  assert.deepEqual(calls, [["download", "1.0.0"], ["run", "/tmp/ban-code-comments", ["--format", "text", "src"]]]);
});

test("runs the hook launcher entrypoint without downloading for invalid input", async () => {
  const originalArguments = process.argv;
  process.argv = [process.execPath, path.resolve("src/launcher-entry.js"), "invalid"];
  try {
    await import("../src/launcher-entry.js");
  } finally {
    process.argv = originalArguments;
  }
  await new Promise((resolve) => setImmediate(resolve));
  assert.equal(process.exitCode, 0);
  process.exitCode = undefined;
});
