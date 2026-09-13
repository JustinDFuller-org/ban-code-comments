import { runHook as evaluateHookInput } from "./hook.js";

async function runHook(mode, options = {}) {
  return evaluateHookInput(options.input || process.stdin, mode, options.output || process.stdout);
}

async function main(modeOverride) {
  const modeArgumentIndex = process.argv.indexOf("--mode");
  const requestedMode = modeArgumentIndex >= 0 ? process.argv[modeArgumentIndex + 1] : process.argv[2];
  const mode = modeOverride || requestedMode || "hard";
  if (mode !== "hard" && mode !== "warn") {
    process.stdout.write(JSON.stringify({ decision: "block", reason: `ban-code-comments hook launcher failed: unsupported hook mode ${mode}` }) + "\n");
    return 0;
  }
  return runHook(mode);
}

export { main, runHook };
