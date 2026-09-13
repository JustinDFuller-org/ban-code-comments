export function renderJSON(result) {
  return `${JSON.stringify(result)}\n`;
}

export function renderText(result) {
  return result.findings.map((item) => `${item.path}:${item.range.start.line}:${item.range.start.column}: ${item.text}`).join("\n") + (result.findings.length ? "\n" : "");
}
