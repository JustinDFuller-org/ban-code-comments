import { DEFAULT_CATEGORIES, CATEGORIES } from "./model.js";
import { parseSelection } from "./languages.js";

const categoryNames = new Set(Object.values(CATEGORIES));

function valuesFor(args, index, option) {
  const value = args[index + 1];
  if (value === undefined || value.startsWith("-")) throw new Error(`${option} requires a value`);
  return [value, index + 1];
}

export function parseArgs(args = []) {
  const languageValues = [];
  const categoryValues = [];
  const includes = [];
  const excludes = [];
  const paths = [];
  let format = "json";
  let debug = false;
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === "--debug") debug = true;
    else if (argument === "--help" || argument === "-h") return { help: true };
    else if (argument === "--version" || argument === "-v") return { version: true };
    else if (["--language", "--languages"].includes(argument)) {
      const [value, next] = valuesFor(args, index, argument); languageValues.push(value); index = next;
    } else if (argument === "--categories") {
      const [value, next] = valuesFor(args, index, argument); categoryValues.push(value); index = next;
    } else if (["--include", "--exclude"].includes(argument)) {
      const [value, next] = valuesFor(args, index, argument); (argument === "--include" ? includes : excludes).push(value); index = next;
    } else if (argument === "--format") {
      const [value, next] = valuesFor(args, index, argument); format = value.toLowerCase(); index = next;
    } else if (argument.startsWith("-")) throw new Error(`unknown option ${argument}`);
    else paths.push(argument);
  }
  if (!["json", "text"].includes(format)) throw new Error(`unsupported format ${JSON.stringify(format)}`);
  const categories = categoryValues.flatMap((value) => String(value).split(",").map((item) => item.trim().toLowerCase())).filter(Boolean);
  for (const category of categories) if (!categoryNames.has(category)) throw new Error(`unsupported category ${JSON.stringify(category)}`);
  return { paths: paths.length ? paths : ["."], languages: parseSelection(languageValues), categories: new Set(categories.length ? categories : DEFAULT_CATEGORIES), includes, excludes, format, debug };
}

export function exitStatus(findings, error = null) {
  return error ? 2 : findings.length ? 1 : 0;
}
