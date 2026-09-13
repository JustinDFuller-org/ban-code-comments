const extensionLanguages = new Map([
  [".go", "go"], [".js", "javascript"], [".jsx", "javascript"], [".mjs", "javascript"], [".cjs", "javascript"],
  [".ts", "typescript"], [".tsx", "typescript"], [".mts", "typescript"], [".cts", "typescript"],
  [".py", "python"], [".pyw", "python"], [".rs", "rust"], [".java", "java"],
  [".c", "c"], [".h", "c"], [".cc", "cpp"], [".cpp", "cpp"], [".cxx", "cpp"], [".hh", "cpp"], [".hpp", "cpp"], [".hxx", "cpp"],
  [".cs", "csharp"], [".kt", "kotlin"], [".kts", "kotlin"], [".swift", "swift"], [".rb", "ruby"], [".rake", "ruby"], [".php", "php"],
  [".sh", "shell"], [".bash", "shell"], [".zsh", "shell"], [".fish", "shell"], [".ksh", "shell"], [".csh", "shell"], [".sql", "sql"],
  [".html", "html"], [".htm", "html"], [".xhtml", "html"], [".xml", "xml"], [".svg", "xml"], [".css", "css"], [".scss", "scss"], [".sass", "scss"],
  [".yaml", "yaml"], [".yml", "yaml"], [".toml", "toml"], [".json", "json"], [".jsonc", "jsonc"], [".hcl", "hcl"], [".tf", "terraform"], [".tfvars", "terraform"],
  [".mk", "makefile"], [".mak", "makefile"], [".ini", "ini"], [".cfg", "ini"], [".conf", "ini"],
]);

const specialLanguages = new Map([["dockerfile", "dockerfile"], ["makefile", "makefile"]]);

const aliases = new Map([
  ["c++", "cpp"], ["c#", "csharp"], ["cs", "csharp"], ["js", "javascript"], ["jsx", "javascript"], ["ts", "typescript"], ["tsx", "typescript"],
  ["sh", "shell"], ["bash", "shell"], ["hcl", "hcl"], ["tf", "terraform"],
]);

export function lookup(filePath) {
  const normalized = String(filePath).replaceAll("\\", "/").toLowerCase();
  const base = normalized.slice(normalized.lastIndexOf("/") + 1);
  if (specialLanguages.has(base)) return specialLanguages.get(base);
  const dot = base.lastIndexOf(".");
  return extensionLanguages.get(dot < 0 ? "" : base.slice(dot)) || null;
}

export function supported() {
  return [...new Set([...extensionLanguages.values(), ...specialLanguages.values()])].sort();
}

export function parseSelection(values = []) {
  const selection = new Set();
  for (const value of values) {
    for (const raw of String(value).split(",")) {
      const name = raw.trim().toLowerCase();
      if (!name) continue;
      const language = aliases.get(name) || name;
      if (!supported().includes(language)) throw new Error(`unsupported language ${JSON.stringify(raw.trim())}`);
      selection.add(language);
    }
  }
  return selection;
}

export const extensions = Object.freeze(Object.fromEntries(extensionLanguages));
export const languageAliases = Object.freeze(Object.fromEntries(aliases));
