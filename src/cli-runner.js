import { check } from "./check.js";
import { exitStatus, parseArgs } from "./cli.js";
import { renderJSON, renderText } from "./report.js";

export async function runCLI(args = [], streams = {}) {
  const output = streams.stdout || process.stdout;
  const errors = streams.stderr || process.stderr;
  try {
    const options = parseArgs(args);
    if (options.help) { output.write("Usage: ban-code-comments [options] [path ...]\n"); return 0; }
    if (options.version) { output.write("1.0.1\n"); return 0; }
    const value = await check(options.paths, { ...options, onDiagnostic: options.debug ? (item) => errors.write(`debug: ${item.path}: ${item.reason}\n`) : undefined });
    output.write(options.format === "text" ? renderText(value) : renderJSON(value));
    if (options.debug) errors.write(`scanned ${value.summary.files_scanned} source(s), skipped ${value.summary.files_skipped} path(s)\n`);
    return exitStatus(value.findings);
  } catch (error) {
    errors.write(`${error.message}\n`);
    return 2;
  }
}
