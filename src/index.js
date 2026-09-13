import * as core from "@actions/core";
import path from "node:path";
import { pathToFileURL } from "node:url";
import { cliArguments } from "./inputs.js";
import { runCLI } from "./cli-runner.js";

async function main(dependencies = {}) {
  const coreAPI = dependencies.core || core;
  const execute = dependencies.run || ((args, options) => runCLI(args, options));
  const inputs = {
    paths: coreAPI.getInput("paths"),
    languages: coreAPI.getInput("languages"),
    categories: coreAPI.getInput("categories"),
    include: coreAPI.getInput("include"),
    exclude: coreAPI.getInput("exclude"),
    format: coreAPI.getInput("format") || "json",
    debug: coreAPI.getInput("debug") || "false",
  };
  return execute(cliArguments(inputs), { cwd: process.env.GITHUB_WORKSPACE || process.cwd() });
}

if (process.argv[1] && import.meta.url === pathToFileURL(path.resolve(process.argv[1])).href) {
  main()
    .then((code) => {
      process.exitCode = code;
    })
    .catch((error) => {
      console.error(error instanceof Error ? error.message : String(error));
      process.exitCode = 2;
    });
}

export { main };
