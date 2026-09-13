import { runClaudeHook } from "./claude-hook.js";

async function inputText() {
  const chunks = [];
  for await (const chunk of process.stdin) chunks.push(chunk);
  return Buffer.concat(chunks).toString("utf8");
}

async function main(modeOverride) {
  const modeArgumentIndex = process.argv.indexOf("--mode");
  const requestedMode = modeArgumentIndex >= 0 ? process.argv[modeArgumentIndex + 1] : process.argv[2];
  const mode = modeOverride || requestedMode || "hard";
  return runClaudeHook(await inputText(), mode);
}

if (process.argv[1] && import.meta.url === new URL(process.argv[1], "file:").href) main().then((code) => { process.exitCode = code; });

export { main, inputText };
