import { access } from "node:fs/promises";
import { downloadPluginCLI } from "./plugin-cache.js";
import { runCLI } from "./runner.js";

function launchFailure(mode, error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stdout.write(JSON.stringify(mode === "hard"
    ? { decision: "block", reason: `ban-code-comments hook launcher failed: ${message}` }
    : { systemMessage: `ban-code-comments hook launcher failed: ${message}` }) + "\n");
}

async function runHook(mode, options = {}) {
  const executable = await downloadPluginCLI(options);
  await access(executable);
  return runCLI(executable, ["hook", "--mode", mode], options.cwd || process.cwd());
}

async function main(modeOverride) {
  const modeArgumentIndex = process.argv.indexOf("--mode");
  const requestedMode = modeArgumentIndex >= 0 ? process.argv[modeArgumentIndex + 1] : process.argv[2];
  const mode = modeOverride || requestedMode || "hard";
  if (mode !== "hard" && mode !== "warn") {
    launchFailure(mode, new Error(`unsupported hook mode ${mode}`));
    return 0;
  }
  try {
    return await runHook(mode);
  } catch (error) {
    launchFailure(mode, error);
    return 0;
  }
}

export { main, runHook };
