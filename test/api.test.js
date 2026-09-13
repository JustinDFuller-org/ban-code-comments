import assert from "node:assert/strict";
import test from "node:test";
import * as api from "../src/api.js";

test("public API exports shared operations", () => {
  for (const name of ["check", "discover", "scanSource", "renderJSON", "renderText", "runCLI", "lookup", "parseArgs"]) assert.equal(typeof api[name], "function");
});
