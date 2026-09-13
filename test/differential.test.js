import test from "node:test";
import { run } from "../scripts/differential.mjs";

if (process.env.RUN_DIFFERENTIAL === "1") test("Go and JavaScript implementations agree on the migration matrix", async () => {
  await run();
});
