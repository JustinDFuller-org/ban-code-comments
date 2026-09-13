import { runHook as evaluateHookInput } from "./hook.js";

async function runHook(mode, options = {}) {
  return evaluateHookInput(options.input || process.stdin, mode, options.output || process.stdout);
}

async function main(modeOverride) {
  const modeArgumentIndex = process.argv.indexOf("--mode");
  const requestedMode = modeArgumentIndex >= 0 ? process.argv[modeArgumentIndex + 1] : process.argv[2];
  const mode = modeOverride || requestedMode || "hard";
  return runHook(mode);
}

export { main, runHook };
