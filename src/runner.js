import { spawn } from "node:child_process";

function runCLI(executable, args, cwd = process.env.GITHUB_WORKSPACE || process.cwd()) {
  return new Promise((resolve) => {
    const child = spawn(executable, args, { cwd, stdio: "inherit", windowsHide: true });
    let settled = false;
    const finish = (code) => {
      if (settled) return;
      settled = true;
      resolve(typeof code === "number" ? code : 2);
    };
    child.on("error", (error) => {
      console.error(error instanceof Error ? error.message : String(error));
      finish(2);
    });
    child.on("close", (code) => finish(code));
  });
}

export { runCLI };
