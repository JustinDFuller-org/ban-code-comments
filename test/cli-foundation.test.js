import assert from "node:assert/strict";
import test from "node:test";
import { CATEGORIES, DEFAULT_CATEGORIES } from "../src/model.js";
import { exitStatus, parseArgs } from "../src/cli.js";

test("CLI parser preserves defaults and repeated selections", () => {
  const options = parseArgs(["--language", "go,js", "--languages", "python", "--categories", "directive,header", "--format", "text", "src"]);
  assert.deepEqual([...options.languages].sort(), ["go", "javascript", "python"]);
  assert.deepEqual([...options.categories].sort(), [CATEGORIES.DIRECTIVE, CATEGORIES.HEADER]);
  assert.equal(options.format, "text");
  assert.deepEqual(options.paths, ["src"]);
  assert.deepEqual([...new Set(DEFAULT_CATEGORIES)], [CATEGORIES.ORDINARY, CATEGORIES.DOCUMENTATION]);
});

test("CLI parser validates options and status mapping", () => {
  assert.deepEqual(parseArgs([]).paths, ["."]);
  assert.equal(parseArgs(["--debug"]).debug, true);
  assert.throws(() => parseArgs(["--format", "yaml"]), /unsupported format/);
  assert.throws(() => parseArgs(["--categories", "ordinary,nope"]), /unsupported category/);
  assert.throws(() => parseArgs(["--unknown"]), /unknown option/);
  assert.equal(exitStatus([], null), 0);
  assert.equal(exitStatus([{}], null), 1);
  assert.equal(exitStatus([], new Error("scan")), 2);
});
