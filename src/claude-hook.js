import { promises as fs } from "node:fs";
import path from "node:path";
import { lookup } from "./languages.js";
import { CATEGORIES } from "./model.js";
import { scanSource } from "./scanner.js";

const selected = new Set([CATEGORIES.ORDINARY, CATEGORIES.DOCUMENTATION]);

function operationalResponse(mode, code, message) {
  const reason = `ban-code-comments Claude hook operational error (${code}): ${message}`;
  return { systemMessage: reason, hookSpecificOutput: { hookEventName: "PreToolUse", additionalContext: reason } };
}

function targetPath(root, filePath) {
  if (typeof filePath !== "string" || !filePath.trim()) throw new Error("tool input does not identify a file");
  const target = path.isAbsolute(filePath) ? path.normalize(filePath) : path.resolve(root, filePath);
  if (path.isAbsolute(filePath)) return target;
  const relative = path.relative(root, target);
  if (relative === ".." || relative.startsWith(`..${path.sep}`) || path.isAbsolute(relative)) throw new Error(`proposed path escapes workspace: ${JSON.stringify(filePath)}`);
  return target;
}

function displayPath(root, filePath) {
  const target = targetPath(root, filePath);
  const relative = path.relative(root, target);
  return relative && relative !== ".." && !relative.startsWith(`..${path.sep}`) ? relative.replaceAll(path.sep, "/") : target.replaceAll(path.sep, "/");
}

function replacement(source, oldText, newText, replaceAll) {
  if (typeof oldText !== "string" || typeof newText !== "string" || oldText.length === 0) throw new Error("Edit input must include non-empty old_string and string new_string");
  if (!source.includes(oldText)) throw new Error("proposed edit could not find old_string in the target file");
  if (replaceAll === true) return source.replaceAll(oldText, newText);
  return source.replace(oldText, newText);
}

function proposedContent(toolName, input, oldSource) {
  if (!input || typeof input !== "object" || Array.isArray(input)) throw new Error("tool input is not an object");
  if (toolName === "Write") {
    if (typeof input.content !== "string") throw new Error("Write input must include string content");
    return input.content;
  }
  if (!("old_string" in input) || !("new_string" in input)) throw new Error("Edit input must include old_string and new_string");
  if (input.replace_all !== undefined && typeof input.replace_all !== "boolean") throw new Error("Edit replace_all must be a boolean");
  return replacement(oldSource, input.old_string, input.new_string, input.replace_all === true);
}

function difference(before, after) {
  const counts = new Map();
  for (const item of before) {
    const key = `${item.language}\0${item.category}\0${item.text}`;
    counts.set(key, (counts.get(key) || 0) + 1);
  }
  return after.filter((item) => {
    const key = `${item.language}\0${item.category}\0${item.text}`;
    const count = counts.get(key) || 0;
    if (!count) return true;
    counts.set(key, count - 1);
    return false;
  });
}

function findings(filePath, source, root) {
  const language = lookup(filePath);
  return language ? scanSource(source, displayPath(root, filePath), language, selected) : [];
}

async function evaluate(event) {
  const root = path.resolve(event.cwd || process.cwd());
  const input = event.tool_input;
  const filePath = input?.file_path;
  const target = targetPath(root, filePath);
  let oldSource = "";
  try {
    oldSource = await fs.readFile(target, "utf8");
  } catch (error) {
    if (error.code !== "ENOENT") throw error;
  }
  const newSource = proposedContent(event.tool_name, input, oldSource);
  return difference(findings(filePath, oldSource, root), findings(filePath, newSource, root));
}

function formatFindings(items) {
  return `ban-code-comments found newly introduced comments: ${[...items].sort((a, b) => a.path.localeCompare(b.path) || a.range.start.line - b.range.start.line || a.range.start.column - b.range.start.column).map((item) => `${item.path}:${item.range.start.line}:${item.range.start.column} [${item.category}] ${item.text.trim()}`).join("; ")}`;
}

function findingResponse(mode, items) {
  if (!items.length) return {};
  const reason = formatFindings(items);
  if (mode === "hard") return { hookSpecificOutput: { hookEventName: "PreToolUse", permissionDecision: "deny", permissionDecisionReason: reason } };
  const guidance = `${reason} Use Git history for history, pull-request descriptions for rationale, simplified code or a nearby README for complexity, and Markdown for general documentation.`;
  return { systemMessage: guidance, hookSpecificOutput: { hookEventName: "PreToolUse", additionalContext: guidance } };
}

export async function evaluateClaudeHook(event, mode = "hard") {
  if (mode !== "hard" && mode !== "warn") return operationalResponse(mode, "invalid_mode", `unsupported hook mode ${JSON.stringify(mode)}`);
  if (event?.hook_event_name !== "PreToolUse" || !["Edit", "Write"].includes(event?.tool_name)) return {};
  try {
    return findingResponse(mode, await evaluate(event));
  } catch (error) {
    return operationalResponse(mode, "proposal_unreadable", error.message);
  }
}

export async function runClaudeHook(input, mode = "hard", output = process.stdout) {
  let event;
  try { event = typeof input === "string" ? JSON.parse(input) : input; }
  catch (error) { output.write(`${JSON.stringify(operationalResponse(mode, "invalid_event", `could not decode hook event: ${error.message}`))}\n`); return 0; }
  try { output.write(`${JSON.stringify(await evaluateClaudeHook(event, mode))}\n`); }
  catch (error) { output.write(`${JSON.stringify(operationalResponse(mode, "hook_failure", error.message))}\n`); }
  return 0;
}

export { difference, formatFindings, proposedContent, replacement, targetPath };
