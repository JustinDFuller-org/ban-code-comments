import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { lookup } from "../src/languages.js";
import { CATEGORIES } from "../src/model.js";
import { scanSource } from "../src/scanner.js";

const fixtures = [
  ["go", "go"], ["javascript", "js"], ["typescript", "ts"], ["python", "py"], ["rust", "rs"], ["java", "java"], ["c", "c"], ["cpp", "cpp"], ["csharp", "cs"], ["kotlin", "kt"], ["swift", "swift"], ["ruby", "rb"], ["php", "php"], ["shell", "sh"], ["sql", "sql"], ["html", "html"], ["xml", "xml"], ["css", "css"], ["scss", "scss"], ["yaml", "yaml"], ["toml", "toml"], ["json", "json"], ["jsonc", "jsonc"], ["hcl", "hcl"], ["terraform", "tf"], ["dockerfile", "Dockerfile"], ["makefile", "Makefile"], ["ini", "ini"],
];

test("checked-in fixtures cover every supported language", async (t) => {
  for (const [language, extension] of fixtures) await t.test(language, async () => {
    const base = extension === "Dockerfile" || extension === "Makefile" ? extension : `fixture.${extension}`;
    for (const kind of ["finding", "clean", "false-positive"]) {
      const relative = path.join("test", "fixtures", language, kind, base);
      const source = await readFile(relative, "utf8");
      assert.equal(lookup(relative), language);
      const findings = scanSource(source, relative, language, new Set(Object.values(CATEGORIES)));
      const expected = kind === "finding" && language !== "json" ? 1 : 0;
      assert.equal(findings.length, expected, `${relative}: ${JSON.stringify(findings)}`);
      if (expected) {
        assert.equal(findings[0].category, CATEGORIES.ORDINARY);
        assert.match(findings[0].text, /finding/);
        assert.ok(findings[0].range.start.line >= 1 && findings[0].range.start.column >= 1);
      }
    }
  });
});
