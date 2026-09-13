import assert from "node:assert/strict";
import test from "node:test";
import { renderJSON, renderText } from "../src/report.js";

const value = { findings: [{ path: "src/a.js", language: "javascript", category: "ordinary", range: { start: { line: 2, column: 3 }, end: { line: 2, column: 12 } }, text: "// note" }], summary: { files_scanned: 1, files_skipped: 2, findings: 1 } };

test("reports preserve JSON fields and stable text ordering", () => {
  assert.deepEqual(JSON.parse(renderJSON(value)), value);
  assert.equal(renderText(value), "src/a.js:2:3: // note\n");
  assert.equal(renderText({ findings: [] }), "");
});

test("report writers expose stream failures", () => {
  const broken = { write: () => { throw new Error("closed"); } };
  assert.throws(() => broken.write(renderJSON(value)), /closed/);
});
