import { appendFile } from "node:fs/promises";
import { readFile } from "node:fs/promises";
import { fileURLToPath, pathToFileURL } from "node:url";

function reportTotals(xml, path) {
  const root = xml.match(/<coverage\b[^>]*>/)?.[0];
  const valid = Number(root?.match(/\blines-valid="(\d+)"/)?.[1]);
  const covered = Number(root?.match(/\blines-covered="(\d+)"/)?.[1]);
  if (!root || !Number.isFinite(valid) || !Number.isFinite(covered)) throw new Error(`coverage report is missing line totals: ${path}`);
  if (valid <= 0 || covered < 0 || covered > valid) throw new Error(`coverage report has invalid line totals: ${path}`);
  return { covered, valid };
}

export function aggregateCoverage(reports) {
  const totals = reports.reduce((sum, report) => ({
    covered: sum.covered + report.covered,
    valid: sum.valid + report.valid,
  }), { covered: 0, valid: 0 });
  return { ...totals, percentage: (totals.covered / totals.valid) * 100 };
}

export async function readReports(paths) {
  return Promise.all(paths.map(async path => reportTotals(await readFile(path, "utf8"), path)));
}

async function writeSummary(result, minimum, paths) {
  const summaryPath = process.env.GITHUB_STEP_SUMMARY;
  if (!summaryPath) return;
  const status = result.percentage >= minimum ? "PASS" : "FAIL";
  await appendFile(summaryPath, `## Coverage ${status}\n\n- Reports: ${paths.join(", ")}\n- Covered lines: ${result.covered}/${result.valid}\n- Aggregate line coverage: ${result.percentage.toFixed(2)}%\n- Minimum: ${minimum.toFixed(2)}%\n`);
}

export async function main(args = process.argv.slice(2)) {
  const minimumIndex = args.indexOf("--minimum");
  const minimum = minimumIndex >= 0 ? Number(args[minimumIndex + 1]) : 90;
  const paths = minimumIndex >= 0 ? args.slice(0, minimumIndex) : args;
  if (paths.length === 0 || !Number.isFinite(minimum) || minimum < 0 || minimum > 100) throw new Error("usage: check-coverage.mjs REPORT... [--minimum PERCENT]");
  const result = aggregateCoverage(await readReports(paths));
  await writeSummary(result, minimum, paths);
  process.stdout.write(`aggregate line coverage: ${result.percentage.toFixed(2)}% (${result.covered}/${result.valid}), minimum: ${minimum.toFixed(2)}%\n`);
  if (result.percentage < minimum) throw new Error(`aggregate line coverage is below ${minimum.toFixed(2)}%`);
  return result;
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  main().catch(error => {
    process.stderr.write(`${error.message}\n`);
    process.exitCode = 1;
  });
}
