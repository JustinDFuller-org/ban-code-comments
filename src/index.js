import * as core from "@actions/core";
import { CLI_VERSION } from "./version.js";
import { downloadCLI } from "./downloader.js";
import { cliArguments } from "./inputs.js";
import { runCLI } from "./runner.js";

async function main() {
  const inputs = {
    paths: core.getInput("paths"),
    languages: core.getInput("languages"),
    categories: core.getInput("categories"),
    include: core.getInput("include"),
    exclude: core.getInput("exclude"),
    format: core.getInput("format") || "json",
    debug: core.getInput("debug") || "false",
  };
  const executable = await downloadCLI(CLI_VERSION);
  return runCLI(executable, cliArguments(inputs));
}

main()
  .then((code) => {
    process.exitCode = code;
  })
  .catch((error) => {
    console.error(error instanceof Error ? error.message : String(error));
    process.exitCode = 2;
  });

export { main };
