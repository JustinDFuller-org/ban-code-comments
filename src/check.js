import { promises as fs } from "node:fs";
import { discover } from "./discovery.js";
import { result } from "./model.js";
import { scanSource } from "./scanner.js";

export async function check(inputPaths = [], options = {}) {
  const discovered = await discover(inputPaths, options);
  if (options.onDiagnostic) for (const diagnostic of discovered.diagnostics) options.onDiagnostic(diagnostic);
  const findings = [];
  for (const candidate of discovered.candidates) {
    let source;
    try { source = await fs.readFile(candidate.path, "utf8"); }
    catch (error) { throw new Error(`${candidate.relative}: ${error.message}`); }
    findings.push(...scanSource(source, candidate.relative, candidate.language, options.categories));
  }
  return result(findings, { files_scanned: discovered.candidates.length, files_skipped: discovered.skipped, findings: findings.length });
}
