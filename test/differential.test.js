import test from "node:test";
import { run } from "../scripts/differential.mjs";

test("Go and JavaScript implementations agree on the migration matrix", async () => {
  await run();
});
