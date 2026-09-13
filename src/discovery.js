import { promises as fs } from "node:fs";
import { execFile } from "node:child_process";
import path from "node:path";
import { promisify } from "node:util";
import { lookup } from "./languages.js";

const execFileAsync = promisify(execFile);
const skippedDirectories = new Set([".git", "node_modules", "vendor", "dist", "build", "coverage"]);

export function normalizePath(value) { return String(value).replaceAll("\\", "/"); }

export function globToRegExp(pattern) {
  let expression = "^";
  const normalized = normalizePath(pattern);
  for (let index = 0; index < normalized.length; index += 1) {
    const character = normalized[index];
    if (character === "*" && normalized[index + 1] === "*") {
      if (normalized[index + 2] === "/") { expression += "(?:.*/)?"; index += 2; }
      else { expression += ".*"; index += 1; }
    } else if (character === "*") expression += "[^/]*";
    else if (character === "?") expression += "[^/]";
    else expression += /[.+^${}()|[\]\\]/.test(character) ? `\\${character}` : character;
  }
  return new RegExp(`${expression}$`);
}

function matches(relativePath, patterns) { return patterns.length === 0 || patterns.some((pattern) => globToRegExp(pattern).test(relativePath)); }
function excluded(relativePath, patterns) { return patterns.some((pattern) => globToRegExp(pattern).test(relativePath)); }
function display(filePath, cwd) { return normalizePath(path.relative(cwd, filePath) || path.basename(filePath)); }

async function walk(directory, cwd, options, diagnostics) {
  const entries = await fs.readdir(directory, { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    const fullPath = path.join(directory, entry.name);
    const relative = display(fullPath, cwd);
    if (entry.isSymbolicLink()) { diagnostics.push({ path: relative, reason: "symbolic link" }); continue; }
    if (entry.isDirectory()) {
      if (skippedDirectories.has(entry.name)) diagnostics.push({ path: relative, reason: "fixed directory exclusion" });
      else files.push(...await walk(fullPath, cwd, options, diagnostics));
      continue;
    }
    const language = lookup(fullPath);
    if (!entry.isFile() || !language) { if (entry.isFile()) diagnostics.push({ path: relative, reason: "unsupported file" }); continue; }
    if (!matches(relative, options.includes || [])) { diagnostics.push({ path: relative, reason: "not included" }); continue; }
    if (excluded(relative, options.excludes || [])) { diagnostics.push({ path: relative, reason: "excluded" }); continue; }
    if (options.languages?.size && !options.languages.has(language)) { diagnostics.push({ path: relative, reason: "language filter" }); continue; }
    files.push({ path: fullPath, relative, language });
  }
  return files;
}

async function tracked(cwd) {
  const { stdout } = await execFileAsync("git", ["-C", cwd, "ls-files", "--cached", "--others", "--exclude-standard", "-z"], { encoding: "utf8" });
  return stdout.split("\0").filter(Boolean).map((file) => path.resolve(cwd, file));
}

export async function discover(inputPaths = [], options = {}) {
  const cwd = path.resolve(options.cwd || process.cwd());
  const diagnostics = [];
  const candidates = [];
  const paths = inputPaths.length ? inputPaths : null;
  if (!paths) {
    try {
      for (const fullPath of await tracked(cwd)) {
        const language = lookup(fullPath); const relative = display(fullPath, cwd);
        if (!language || relative.split("/").some((part) => skippedDirectories.has(part))) continue;
        if (matches(relative, options.includes || []) && !excluded(relative, options.excludes || []) && (!options.languages?.size || options.languages.has(language))) candidates.push({ path: fullPath, relative, language });
        else diagnostics.push({ path: relative, reason: "selection filter" });
      }
    } catch { candidates.push(...await walk(cwd, cwd, options, diagnostics)); }
  } else {
    for (const input of paths) {
      const fullPath = path.resolve(cwd, input);
      try {
        const stats = await fs.lstat(fullPath);
        if (stats.isDirectory()) candidates.push(...await walk(fullPath, cwd, options, diagnostics));
        else if (stats.isFile()) {
          const language = lookup(fullPath); const relative = display(fullPath, cwd);
          if (!language) diagnostics.push({ path: relative, reason: "unsupported file" });
          else if (!matches(relative, options.includes || [])) diagnostics.push({ path: relative, reason: "not included" });
          else if (excluded(relative, options.excludes || [])) diagnostics.push({ path: relative, reason: "excluded" });
          else if (options.languages?.size && !options.languages.has(language)) diagnostics.push({ path: relative, reason: "language filter" });
          else candidates.push({ path: fullPath, relative, language });
        }
      } catch (error) { throw new Error(`${input}: ${error.message}`); }
    }
  }
  const unique = [...new Map(candidates.map((candidate) => [path.resolve(candidate.path), candidate])).values()];
  unique.sort((left, right) => left.relative.localeCompare(right.relative));
  return { candidates: unique, diagnostics, skipped: diagnostics.length };
}

export { skippedDirectories };
