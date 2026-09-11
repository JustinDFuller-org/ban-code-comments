function nonEmptyLines(value) {
  return String(value || "")
    .split(/\r?\n/)
    .filter((line) => line.length > 0);
}

function addRepeated(args, flag, value) {
  for (const item of nonEmptyLines(value)) {
    args.push(flag, item);
  }
}

function parseBoolean(value) {
  const normalized = String(value || "false").toLowerCase();
  if (normalized === "true") return true;
  if (normalized === "false" || normalized === "") return false;
  throw new Error(`invalid boolean input for debug: ${value}`);
}

function cliArguments(inputs) {
  const args = [];
  const paths = nonEmptyLines(inputs.paths);
  addRepeated(args, "--languages", inputs.languages);
  if (inputs.categories) {
    const categories = nonEmptyLines(inputs.categories).join(",");
    if (categories) args.push("--categories", categories);
  }
  addRepeated(args, "--include", inputs.include);
  addRepeated(args, "--exclude", inputs.exclude);
  args.push("--format", inputs.format || "json");
  if (parseBoolean(inputs.debug)) args.push("--debug");
  args.push(...(paths.length > 0 ? paths : ["."]));
  return args;
}

export { cliArguments, nonEmptyLines, parseBoolean };
