import assert from "node:assert/strict";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { aggregateCoverage, readReports } from "../scripts/check-coverage.mjs";

const report = (valid, covered) => `<coverage lines-valid="${valid}" lines-covered="${covered}"></coverage>`;

test("aggregateCoverage weights reports by valid lines", () => {
  assert.deepEqual(aggregateCoverage([{ valid: 100, covered: 90 }, { valid: 300, covered: 285 }]), {
    valid: 400,
    covered: 375,
    percentage: 93.75,
  });
});

test("readReports rejects invalid totals", async () => {
  const directory = await mkdtemp(path.join(os.tmpdir(), "ban-code-comments-coverage-test-"));
  const reportPath = path.join(directory, "invalid.xml");
  try {
    await writeFile(reportPath, report(10, 11));
    await assert.rejects(readReports([reportPath]), /invalid line totals/);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});
