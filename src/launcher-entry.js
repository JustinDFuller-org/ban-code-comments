import { main } from "./hook-launcher.js";

main().then((code) => {
  process.exitCode = code;
});
