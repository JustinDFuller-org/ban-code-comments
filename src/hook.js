import { promises as fs } from "node:fs";
import path from "node:path";
import { lookup } from "./languages.js";
import { CATEGORIES } from "./model.js";
import { scanSource } from "./scanner.js";

const tools = new Set(["apply_patch", "edit", "write", "write_file", "file_write"]);
const selected = new Set([CATEGORIES.ORDINARY, CATEGORIES.DOCUMENTATION]);

function errorResponse(mode, code, message) {
  const text = `ban-code-comments hook operational error (${code}): ${message}`;
  return { systemMessage: text, hookSpecificOutput: { hookEventName: "PreToolUse", additionalContext: text } };
}

function resolve(root, name) {
  const normalized = String(name || "").trim().replaceAll("\\", "/").replace(new RegExp("^(?:a|b)/"), "");
  if (!normalized) throw new Error(`invalid proposed path ${JSON.stringify(name || "")}`);
  const target = path.isAbsolute(normalized) ? path.normalize(normalized) : path.resolve(root, normalized);
  const relative = path.relative(root, target);
  if (relative === ".." || relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative)) throw new Error(`proposed path escapes workspace: ${JSON.stringify(name)}`);
  return target;
}

function display(root, name) { return path.relative(root, resolve(root, name)).replaceAll(path.sep, "/"); }

function stringValue(input, ...keys) { for (const key of keys) if (typeof input?.[key] === "string") return input[key]; return ""; }

function parsePatch(patch) {
  const lines = String(patch).replaceAll("\r\n", "\n").split("\n");
  const wrapped = lines[0]?.trim() === "*** Begin Patch";
  if (wrapped && !lines.some((line) => line.trim() === "*** End Patch")) throw new Error("patch is missing end marker");
  let index = wrapped ? 1 : 0;
  const changes = [];
  while (index < lines.length) {
    const line = lines[index];
    if (!line || line.trim() === "*** End Patch") { index += 1; continue; }
    if (line.startsWith("*** Add File: ")) {
      const body = collectBody(lines, index + 1); if (body.lines.some((item) => item && !item.startsWith("+"))) throw new Error(`unsupported add-file line ${JSON.stringify(body.lines.find((item) => item && !item.startsWith("+")))}`); changes.push({ path: line.slice(14), source: body.lines.filter((item) => item.startsWith("+")).map((item) => item.slice(1)).join("\n") }); index = body.next; continue;
    }
    if (line.startsWith("*** Delete File: ")) { changes.push({ path: line.slice(17), delete: true }); index += 1; continue; }
    if (line.startsWith("*** Update File: ")) {
      const body = collectBody(lines, index + 1); let newPath = ""; let patchLines = body.lines;
      if (patchLines[0]?.startsWith("*** Move to: ")) { newPath = patchLines[0].slice(13); patchLines = patchLines.slice(1); }
      changes.push({ path: line.slice(17), newPath, patchLines }); index = body.next; continue;
    }
    throw new Error(`unsupported patch line ${JSON.stringify(line)}`);
  }
  return changes;
}

function collectBody(lines, start) {
  let next = start;
  while (next < lines.length && !/^(?:\*\*\* (?:Add|Delete|Update) File:|\*\*\* End Patch)/.test(lines[next])) next += 1;
  return { lines: lines.slice(start, next), next };
}

function reconstruct(raw) {
  if (raw === undefined || raw === null) throw new Error("hook event has no tool input");
  if (typeof raw === "string") return parsePatch(raw);
  if (typeof raw !== "object" || Array.isArray(raw)) throw new Error("tool input is not an object");
  const patchKeys = ["patch", "input", "command"].filter((key) => Object.hasOwn(raw, key));
  const contentKeys = ["content", "new_content", "newContent"].filter((key) => Object.hasOwn(raw, key));
  const pathKeys = ["path", "file_path", "filePath", "filename"].filter((key) => Object.hasOwn(raw, key));
  const editKeys = ["old_string", "oldString", "new_string", "newString"].filter((key) => Object.hasOwn(raw, key));
  if (patchKeys.length > 1 || contentKeys.length > 1 || pathKeys.length > 1) throw new Error("tool input contains duplicate field aliases");
  if (patchKeys.length > 0 && (contentKeys.length > 0 || pathKeys.length > 0 || editKeys.length > 0)) throw new Error("tool input contains conflicting proposal fields");
  if (contentKeys.length > 0 && editKeys.length > 0) throw new Error("tool input contains conflicting proposal fields");
  const patch = stringValue(raw, "patch", "input", "command");
  if (patch.trimStart().startsWith("*** Begin Patch")) return parsePatch(patch);
  const filePath = stringValue(raw, "path", "file_path", "filePath", "filename");
  if (!filePath) throw new Error("tool input does not identify a file");
  for (const key of ["content", "new_content", "newContent"]) if (typeof raw[key] === "string") return [{ path: filePath, source: raw[key] }];
  const oldKeys = ["old_string", "oldString"].filter((key) => Object.hasOwn(raw, key));
  const newKeys = ["new_string", "newString"].filter((key) => Object.hasOwn(raw, key));
  if (oldKeys.length > 1 || newKeys.length > 1) throw new Error("edit input contains duplicate field aliases");
  if (oldKeys.length > 0) {
    const oldText = raw[oldKeys[0]];
    const newText = newKeys.length > 0 ? raw[newKeys[0]] : undefined;
    if (typeof oldText !== "string") throw new Error("edit input must include string old_string");
    if (oldText.length === 0) throw new Error("edit input must include non-empty old_string");
    if (typeof newText !== "string") throw new Error("edit input must include string new_string");
    return [{ path: filePath, oldText, newText }];
  }
  throw new Error("tool input does not contain proposed file content");
}

function applyPatch(source, patchLines) {
  const finalNewline = source.endsWith("\n");
  const oldLines = source.replace(/\n$/, "").split("\n");
  if (oldLines.length === 1 && oldLines[0] === "") oldLines.length = 0;
  const output = []; let cursor = 0; let changed = false; let hunk = [];
  const flush = () => {
    const operations = hunk.filter((line) => line && !line.startsWith("@@") && line !== "\\ No newline at end of file");
    if (!operations.length) return;
    const oldBlock = operations.filter((line) => line[0] === " " || line[0] === "-").map((line) => line.slice(1));
    const newBlock = operations.filter((line) => line[0] === " " || line[0] === "+").map((line) => line.slice(1));
    if (operations.some((line) => ![" ", "+", "-"].includes(line[0]))) throw new Error(`unsupported hunk line ${JSON.stringify(operations.find((line) => ![" ", "+", "-"].includes(line[0])))}`);
    let match = -1;
    for (let candidate = cursor; candidate + oldBlock.length <= oldLines.length; candidate += 1) if (oldLines.slice(candidate, candidate + oldBlock.length).every((line, i) => line === oldBlock[i])) { match = candidate; break; }
    if (match < 0) throw new Error("hunk context was not found");
    output.push(...oldLines.slice(cursor, match), ...newBlock); cursor = match + oldBlock.length; changed = true;
  };
  for (const line of patchLines) { if (line.startsWith("@@")) { flush(); hunk = []; } hunk.push(line); }
  flush();
  if (!changed) throw new Error("patch contains no file operations");
  output.push(...oldLines.slice(cursor));
  return output.join("\n") + (finalNewline ? "\n" : "");
}

function difference(before, after) {
  const counts = new Map();
  for (const item of before) { const key = `${item.language}\0${item.category}\0${item.text}`; counts.set(key, (counts.get(key) || 0) + 1); }
  return after.filter((item) => { const key = `${item.language}\0${item.category}\0${item.text}`; const count = counts.get(key) || 0; if (count) { counts.set(key, count - 1); return false; } return true; });
}

async function findingsFor(root, filePath, source) {
  const language = lookup(filePath);
  return language ? scanSource(source, display(root, filePath), language, selected) : [];
}

async function evaluate(root, changes) {
  const findings = [];
  for (const change of changes) {
    const oldPath = resolve(root, change.path);
    let oldSource = "";
    try { oldSource = await fs.readFile(oldPath, "utf8"); } catch (error) { if (error.code !== "ENOENT") throw error; }
    let newSource = change.source;
    if (change.oldText !== undefined) { if (!oldSource.includes(change.oldText)) throw new Error(`proposed edit could not find old text in ${change.path}`); newSource = oldSource.replace(change.oldText, change.newText); }
    if (change.patchLines?.length) newSource = applyPatch(oldSource, change.patchLines);
    if (change.delete) newSource = "";
    const newPath = change.newPath ? resolve(root, change.newPath) && change.newPath : change.path;
    findings.push(...difference(await findingsFor(root, change.path, oldSource), await findingsFor(root, newPath, newSource)));
  }
  return findings;
}

function formatFindings(findings) {
  return `ban-code-comments found newly introduced comments: ${[...findings].sort((a, b) => a.path.localeCompare(b.path) || a.range.start.line - b.range.start.line || a.range.start.column - b.range.start.column).map((item) => `${item.path}:${item.range.start.line}:${item.range.start.column} [${item.category}] ${item.text.trim()}`).join("; ")}`;
}

function findingResponse(mode, findings) {
  if (!findings.length) return {};
  const reason = formatFindings(findings);
  if (mode === "hard") return { decision: "block", reason };
  const guidance = `${reason} Use Git history for history, pull-request descriptions for rationale, simplified code or a nearby README for complexity, and Markdown for general documentation.`;
  return { systemMessage: guidance, hookSpecificOutput: { hookEventName: "PreToolUse", additionalContext: guidance } };
}

export async function evaluateHook(event, mode = "hard") {
  if (mode !== "hard" && mode !== "warn") return errorResponse(mode, "invalid_mode", `unsupported hook mode ${JSON.stringify(mode)}`);
  const eventName = String(event?.hook_event_name || "").trim().toLowerCase().replaceAll("_", "");
  const toolName = String(event?.tool_name || "").trim().toLowerCase();
  if (eventName !== "pretooluse" || !tools.has(toolName)) return {};
  try { return findingResponse(mode, await evaluate(path.resolve(event.cwd || process.cwd()), reconstruct(event.tool_input))); }
  catch (error) { return errorResponse(mode, "proposal_unreadable", error.message); }
}

async function decodeInput(input) {
  if (typeof input === "string") return JSON.parse(input);
  if (Buffer.isBuffer(input)) return JSON.parse(input.toString("utf8"));
  if (input && typeof input === "object" && typeof input[Symbol.asyncIterator] === "function") {
    const chunks = [];
    for await (const chunk of input) chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk)));
    return JSON.parse(Buffer.concat(chunks).toString("utf8"));
  }
  if (input && typeof input === "object" && !Array.isArray(input)) return input;
  throw new Error("hook input is neither JSON text nor an event object");
}

export async function runHook(input, mode = "hard", output = process.stdout) {
  let event;
  try { event = await decodeInput(input); } catch (error) { output.write(`${JSON.stringify(errorResponse(mode, "invalid_event", `could not decode hook event: ${error.message}`))}\n`); return 0; }
  try { output.write(`${JSON.stringify(await evaluateHook(event, mode))}\n`); }
  catch (error) { output.write(`${JSON.stringify(errorResponse(mode, "hook_failure", error.message))}\n`); }
  return 0;
}

export { applyPatch, formatFindings, parsePatch, reconstruct, resolve };
