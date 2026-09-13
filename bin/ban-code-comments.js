#!/usr/bin/env node
import { parseArgs } from "../src/cli.js";

try {
  const options = parseArgs(process.argv.slice(2));
  if (options.help) process.stdout.write("Usage: ban-code-comments [options] [path ...]\n");
  else if (options.version) process.stdout.write("1.0.1\n");
  else process.stdout.write(`${JSON.stringify({ findings: [], summary: { files_scanned: 0, files_skipped: 0, findings: 0 } })}\n`);
  process.exitCode = 0;
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
}
