import assert from "node:assert/strict";
import test from "node:test";
import { cliArguments, parseBoolean } from "../src/inputs.js";
import { main as actionMain } from "../src/index.js";

test("maps Action inputs to shared CLI arguments", () => {
  assert.deepEqual(cliArguments({ paths: "src folder\nfile.go", languages: "go,python\nrust", categories: "ordinary\ndirective", include: "src/**/*.go", exclude: "vendor/**", format: "text", debug: "true" }), ["--languages", "go,python", "--languages", "rust", "--categories", "ordinary,directive", "--include", "src/**/*.go", "--exclude", "vendor/**", "--format", "text", "--debug", "src folder", "file.go"]);
  assert.deepEqual(cliArguments({ format: "json", debug: "false" }), ["--format", "json", "."]);
  assert.throws(() => parseBoolean("sometimes"), /invalid boolean/);
});

test("Action invokes the shared CLI without a downloader", async () => {
  const calls = [];
  const code = await actionMain({
    core: { getInput: (name) => ({ paths: "src", format: "text" }[name] || "") },
    run: async (...args) => { calls.push(args); return 1; },
  });
  assert.equal(code, 1);
  assert.deepEqual(calls[0][0], ["--format", "text", "src"]);
  assert.equal(typeof calls[0][1].cwd, "string");
});
