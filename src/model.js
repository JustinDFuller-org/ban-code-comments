export const CATEGORIES = Object.freeze({
  ORDINARY: "ordinary",
  DOCUMENTATION: "documentation",
  HEADER: "header",
  DIRECTIVE: "directive",
});

export const DEFAULT_CATEGORIES = Object.freeze([
  CATEGORIES.ORDINARY,
  CATEGORIES.DOCUMENTATION,
]);

export function position(line, column) {
  return { line, column };
}

export function sourceRange(start, end) {
  return { start, end };
}

export function finding(path, language, category, range, text) {
  return { path, language, category, range, text };
}

export function result(findings = [], summary = {}) {
  return {
    findings,
    summary: {
      files_scanned: summary.files_scanned ?? 0,
      files_skipped: summary.files_skipped ?? 0,
      findings: summary.findings ?? findings.length,
    },
  };
}

export function exitCode(findings = [], error = null) {
  if (error) return 2;
  return findings.length > 0 ? 1 : 0;
}
