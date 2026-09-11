package languages

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

var extensionLanguages = map[string]model.Language{
	".go": model.Language("go"),
	".js": model.Language("javascript"), ".jsx": model.Language("javascript"), ".mjs": model.Language("javascript"), ".cjs": model.Language("javascript"),
	".ts": model.Language("typescript"), ".tsx": model.Language("typescript"), ".mts": model.Language("typescript"), ".cts": model.Language("typescript"),
	".py": model.Language("python"), ".pyw": model.Language("python"),
	".rs":   model.Language("rust"),
	".java": model.Language("java"),
	".c":    model.Language("c"), ".h": model.Language("c"), ".cc": model.Language("cpp"), ".cpp": model.Language("cpp"), ".cxx": model.Language("cpp"), ".hh": model.Language("cpp"), ".hpp": model.Language("cpp"), ".hxx": model.Language("cpp"),
	".cs": model.Language("csharp"),
	".kt": model.Language("kotlin"), ".kts": model.Language("kotlin"),
	".swift": model.Language("swift"),
	".rb":    model.Language("ruby"), ".rake": model.Language("ruby"),
	".php": model.Language("php"),
	".sh":  model.Language("shell"), ".bash": model.Language("shell"), ".zsh": model.Language("shell"), ".fish": model.Language("shell"), ".ksh": model.Language("shell"), ".csh": model.Language("shell"),
	".sql":  model.Language("sql"),
	".html": model.Language("html"), ".htm": model.Language("html"), ".xhtml": model.Language("html"), ".xml": model.Language("xml"), ".svg": model.Language("xml"),
	".css": model.Language("css"), ".scss": model.Language("scss"), ".sass": model.Language("scss"),
	".yaml": model.Language("yaml"), ".yml": model.Language("yaml"),
	".toml": model.Language("toml"),
	".json": model.Language("json"), ".jsonc": model.Language("jsonc"),
	".hcl": model.Language("hcl"), ".tf": model.Language("terraform"), ".tfvars": model.Language("terraform"),
	".mk": model.Language("makefile"), ".mak": model.Language("makefile"),
	".ini": model.Language("ini"), ".cfg": model.Language("ini"), ".conf": model.Language("ini"),
}

var specialLanguages = map[string]model.Language{
	"dockerfile": model.Language("dockerfile"),
	"makefile":   model.Language("makefile"),
}

var aliases = map[string]model.Language{
	"c++": model.Language("cpp"), "c#": model.Language("csharp"), "cs": model.Language("csharp"),
	"js": model.Language("javascript"), "jsx": model.Language("javascript"), "ts": model.Language("typescript"), "tsx": model.Language("typescript"),
	"sh": model.Language("shell"), "bash": model.Language("shell"),
	"hcl": model.Language("hcl"), "tf": model.Language("terraform"),
}

func Lookup(path string) (model.Language, bool) {
	base := strings.ToLower(filepath.Base(path))
	if language, ok := specialLanguages[base]; ok {
		return language, true
	}
	language, ok := extensionLanguages[strings.ToLower(filepath.Ext(base))]
	return language, ok
}

func ParseSelection(values []string) (map[model.Language]bool, error) {
	selection := make(map[model.Language]bool)
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			name := strings.ToLower(strings.TrimSpace(item))
			if name == "" {
				continue
			}
			language, ok := aliases[name]
			if !ok {
				language = model.Language(name)
			}
			if !supported(language) {
				return nil, fmt.Errorf("unsupported language %q", item)
			}
			selection[language] = true
		}
	}
	return selection, nil
}

func Supported() []model.Language {
	seen := make(map[model.Language]bool)
	for _, language := range extensionLanguages {
		seen[language] = true
	}
	for _, language := range specialLanguages {
		seen[language] = true
	}
	result := make([]model.Language, 0, len(seen))
	for language := range seen {
		result = append(result, language)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func supported(language model.Language) bool {
	for _, candidate := range Supported() {
		if candidate == language {
			return true
		}
	}
	return false
}
