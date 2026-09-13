#!/usr/bin/env node
import { runCLI } from "../src/cli-runner.js";

process.exitCode = await runCLI(process.argv.slice(2));
