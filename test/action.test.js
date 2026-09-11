import assert from "node:assert/strict";
import * as fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import crypto from "node:crypto";
import { expectedChecksum, verifyChecksum } from "../src/checksum.js";
import { downloadCLI, findExecutable } from "../src/downloader.js";
import { cliArguments, parseBoolean } from "../src/inputs.js";
import { assetName, releaseURLs, targetFor } from "../src/platform.js";
import { runCLI } from "../src/runner.js";

test("maps supported runner targets to release coordinates", () => {
  assert.deepEqual(targetFor("linux", "x64"), { archiveFormat: "tar.gz", releaseOS: "linux", releaseArch: "amd64" });
  assert.deepEqual(targetFor("linux", "arm64"), { archiveFormat: "tar.gz", releaseOS: "linux", releaseArch: "arm64" });
  assert.deepEqual(targetFor("darwin", "x64"), { archiveFormat: "tar.gz", releaseOS: "darwin", releaseArch: "amd64" });
  assert.deepEqual(targetFor("darwin", "arm64"), { archiveFormat: "tar.gz", releaseOS: "darwin", releaseArch: "arm64" });
  assert.deepEqual(targetFor("win32", "x64"), { archiveFormat: "zip", releaseOS: "windows", releaseArch: "amd64" });
  assert.throws(() => targetFor("win32", "arm64"), /unsupported runner/);
  assert.equal(assetName("1.0.0", targetFor("linux", "x64")), "ban-code-comments_1.0.0_linux_amd64.tar.gz");
  assert.deepEqual(releaseURLs("1.0.0", targetFor("linux", "x64")), {
    archive: "https://github.com/JustinDFuller/ban-code-comments/releases/download/v1.0.0/ban-code-comments_1.0.0_linux_amd64.tar.gz",
    checksums: "https://github.com/JustinDFuller/ban-code-comments/releases/download/v1.0.0/checksums.txt",
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

test("uses a cached verified executable without downloading again", async () => {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ban-code-comments-cache-test-"));
  const executable = path.join(directory, "ban-code-comments");
  await fs.writeFile(executable, "binary");
  let downloads = 0;
  const result = await downloadCLI("1.0.0", {
    target: targetFor("linux", "x64"),
    cache: {
      find: () => directory,
      downloadTool: async () => {
        downloads += 1;
        throw new Error("download should not be called on a cache hit");
      },
    },
  });
  assert.equal(result, executable);
  assert.equal(downloads, 0);
});

test("preserves CLI exit statuses", async () => {
  for (const code of [0, 1, 2]) {
    const actual = await runCLI(process.execPath, ["-e", `process.exit(${code})`]);
    assert.equal(actual, code);
  }
});
