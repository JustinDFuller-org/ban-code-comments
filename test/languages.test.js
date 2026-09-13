import assert from "node:assert/strict";
import test from "node:test";
import { extensions, languageAliases, lookup, parseSelection, supported } from "../src/languages.js";

test("registry covers every supported language and special filename", () => {
  const expected = ["c", "cpp", "csharp", "css", "dockerfile", "go", "hcl", "html", "ini", "java", "javascript", "json", "jsonc", "kotlin", "makefile", "php", "python", "ruby", "rust", "scss", "shell", "sql", "swift", "terraform", "toml", "typescript", "xml", "yaml"];
  assert.deepEqual(supported(), expected);
  assert.equal(lookup("Dockerfile"), "dockerfile");
  assert.equal(lookup("Makefile"), "makefile");
  for (const [extension, language] of Object.entries(extensions)) assert.equal(lookup(`fixture${extension}`), language);
});

test("registry resolves aliases and rejects unsupported selections", () => {
  assert.deepEqual([...parseSelection(["go,js", "c#", "terraform"])].sort(), ["csharp", "go", "javascript", "terraform"]);
  for (const [alias, language] of Object.entries(languageAliases)) assert.deepEqual([...parseSelection([alias])], [language]);
  assert.throws(() => parseSelection(["brainfuck"]), /unsupported language/);
});
