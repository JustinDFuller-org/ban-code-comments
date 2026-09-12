package scanner

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/JustinDFuller-org/ban-code-comments/internal/model"
)

type blockDelimiter struct {
	start string
	end   string
}

type syntax struct {
	line       []string
	block      []blockDelimiter
	nested     bool
	heredoc    bool
	rubyBlock  bool
	shell      bool
	yaml       bool
	phpHeredoc bool
}

func Scan(path string, language model.Language, source []byte, selected map[model.Category]bool) []model.Finding {
	spec, ok := syntaxFor(language)
	if !ok {
		return nil
	}
	findings := make([]model.Finding, 0)
	skipPHPUntil := -1
	skipPHPLineEnd := -1
	skipYAMLUntil := -1
	skipYAMLLineEnd := -1
	for index := 0; index < len(source); {
		if skipPHPUntil > index && index >= skipPHPLineEnd {
			index = skipPHPUntil
			continue
		}
		if skipPHPUntil == index {
			skipPHPUntil = -1
			skipPHPLineEnd = -1
		}
		if skipYAMLUntil > index && index >= skipYAMLLineEnd {
			index = skipYAMLUntil
			continue
		}
		if skipYAMLUntil == index {
			skipYAMLUntil = -1
			skipYAMLLineEnd = -1
		}
		if spec.heredoc && atLineStart(source, index) {
			if end, found := skipShellHeredoc(source, index); found {
				index = end
				continue
			}
		}
		if spec.phpHeredoc && atLineStart(source, index) {
			if end, found := skipPHPHeredoc(source, index); found {
				skipPHPUntil = end
				skipPHPLineEnd = index + lineLength(source, index)
			}
		}
		if spec.yaml && atLineStart(source, index) {
			if end, found := skipYAMLBlockScalar(source, index); found {
				skipYAMLUntil = end
				skipYAMLLineEnd = index + lineLength(source, index)
			}
		}
		if spec.rubyBlock && atLineStart(source, index) && bytes.HasPrefix(source[index:], []byte("=begin")) {
			if end := findRubyBlockEnd(source, index); end > index {
				appendFinding(&findings, path, language, source, index, end, selected)
				index = end
				continue
			}
		}
		if end, found := skipString(source, index, language); found {
			index = end
			continue
		}
		matched := false
		for _, delimiter := range spec.block {
			if !bytes.HasPrefix(source[index:], []byte(delimiter.start)) {
				continue
			}
			end := findBlockEnd(source, index+len(delimiter.start), delimiter, spec.nested)
			appendFinding(&findings, path, language, source, index, end, selected)
			index = end
			matched = true
			break
		}
		if matched {
			continue
		}
		for _, delimiter := range spec.line {
			if !bytes.HasPrefix(source[index:], []byte(delimiter)) {
				continue
			}
			if spec.shell && delimiter == "#" && isEscaped(source, index) {
				continue
			}
			if spec.shell && delimiter == "#" && !isShellCommentStart(source, index) {
				continue
			}
			if spec.yaml && delimiter == "#" && !isYAMLCommentStart(source, index) {
				continue
			}
			if language == "sql" && delimiter == "#" && isSQLHashOperator(source, index) {
				continue
			}
			lineEnd := bytes.IndexAny(source[index:], "\r\n")
			var end int
			if lineEnd < 0 {
				end = len(source)
			} else {
				end = index + lineEnd
			}
			appendFinding(&findings, path, language, source, index, end, selected)
			index = end
			matched = true
			break
		}
		if matched {
			continue
		}
		index++
	}
	return findings
}

func syntaxFor(language model.Language) (syntax, bool) {
	cStyle := syntax{line: []string{"//"}, block: []blockDelimiter{{start: "/*", end: "*/"}}}
	switch language {
	case "go", "javascript", "typescript", "java", "c", "cpp", "csharp", "kotlin", "swift", "rust", "jsonc":
		cStyle.nested = language == "rust"
		return cStyle, true
	case "python":
		return syntax{line: []string{"#"}}, true
	case "ruby":
		return syntax{line: []string{"#"}, rubyBlock: true}, true
	case "php":
		return syntax{line: []string{"//", "#"}, block: []blockDelimiter{{start: "/*", end: "*/"}}, phpHeredoc: true}, true
	case "shell", "yaml", "toml", "dockerfile", "makefile":
		return syntax{line: []string{"#"}, heredoc: language == "shell", shell: language == "shell", yaml: language == "yaml"}, true
	case "ini":
		return syntax{line: []string{"#", ";"}}, true
	case "sql":
		return syntax{line: []string{"--", "#"}, block: []blockDelimiter{{start: "/*", end: "*/"}}}, true
	case "html", "xml":
		return syntax{block: []blockDelimiter{{start: "<!--", end: "-->"}}}, true
	case "css":
		return syntax{block: []blockDelimiter{{start: "/*", end: "*/"}}}, true
	case "scss":
		return syntax{line: []string{"//"}, block: []blockDelimiter{{start: "/*", end: "*/"}}}, true
	case "hcl", "terraform":
		return syntax{line: []string{"#", "//"}, block: []blockDelimiter{{start: "/*", end: "*/"}}}, true
	case "json":
		return syntax{}, true
	default:
		return syntax{}, false
	}
}

func appendFinding(findings *[]model.Finding, path string, language model.Language, source []byte, start, end int, selected map[model.Category]bool) {
	if end <= start {
		return
	}
	text := string(source[start:end])
	category := classify(text)
	if len(selected) > 0 && !selected[category] {
		return
	}
	*findings = append(*findings, model.Finding{Path: path, Language: language, Category: category, Range: model.Range{Start: positionAt(source, start), End: positionAt(source, end)}, Text: text})
}

func classify(text string) model.Category {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(trimmed, "///") || strings.HasPrefix(trimmed, "//!") || strings.HasPrefix(trimmed, "/**") {
		return model.CategoryDocumentation
	}
	if strings.HasPrefix(trimmed, "#!") || strings.HasPrefix(lower, "# syntax=") || strings.HasPrefix(lower, "# escape=") {
		return model.CategoryDirective
	}
	for _, marker := range []string{"go:", "cgo", "eslint", "ts-ignore", "ts-expect-error", "prettier-ignore", "nolint", "noinspection", "shellcheck", "yamllint", "yaml-language-server", "coding:", "type: ignore", "swift-tools-version", "region", "endregion"} {
		if strings.Contains(lower, marker) {
			return model.CategoryDirective
		}
	}
	for _, marker := range []string{"copyright", "spdx-license", "licensed", "license", "generated by", "code generated", "auto-generated"} {
		if strings.Contains(lower, marker) {
			return model.CategoryHeader
		}
	}
	return model.CategoryOrdinary
}

func skipYAMLBlockScalar(source []byte, index int) (int, bool) {
	lineEnd := bytes.IndexByte(source[index:], '\n')
	if lineEnd < 0 {
		lineEnd = len(source) - index
	}
	line := source[index : index+lineEnd]
	trimmed := strings.TrimSpace(string(line))
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return 0, false
	}
	indent := len(line) - len(bytes.TrimLeft(line, " "))
	value := ""
	if colon := bytes.LastIndexByte(line, ':'); colon >= 0 {
		value = strings.TrimSpace(string(line[colon+1:]))
	} else if strings.HasPrefix(trimmed, "- ") {
		value = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
	}
	if value == "" || (value[0] != '|' && value[0] != '>') {
		return 0, false
	}
	search := index + lineEnd
	if search < len(source) {
		search++
	}
	for search < len(source) {
		end := bytes.IndexByte(source[search:], '\n')
		if end < 0 {
			end = len(source) - search
		}
		candidate := source[search : search+end]
		if strings.TrimSpace(string(candidate)) != "" {
			candidateIndent := len(candidate) - len(bytes.TrimLeft(candidate, " "))
			if candidateIndent <= indent {
				break
			}
		}
		search += end
		if search < len(source) {
			search++
		}
	}
	return search, true
}

func lineLength(source []byte, index int) int {
	if end := bytes.IndexByte(source[index:], '\n'); end >= 0 {
		return end
	}
	return len(source) - index
}

func skipString(source []byte, index int, language model.Language) (int, bool) {
	if end, found := skipRawString(source, index, language); found {
		return end, true
	}
	if language == "swift" {
		if end, found := skipHashString(source, index); found {
			return end, true
		}
	}
	if language == "csharp" {
		if end, found := skipCSharpRawString(source, index); found {
			return end, true
		}
	}
	quoteIndex := index
	verbatim := false
	if language == "csharp" && index < len(source) {
		if source[index] == '@' && index+1 < len(source) && source[index+1] == '"' {
			quoteIndex = index + 1
			verbatim = true
		} else if source[index] == '@' && index+2 < len(source) && source[index+1] == '$' && source[index+2] == '"' {
			quoteIndex = index + 2
			verbatim = true
		} else if (source[index] == '$' || source[index] == '@') && index+2 < len(source) && source[index+1] == '@' && source[index+2] == '"' {
			quoteIndex = index + 2
			verbatim = true
		}
	}
	if quoteIndex == index && index < len(source) && (source[index] == 'f' || source[index] == 'F' || source[index] == 'u' || source[index] == 'b' || source[index] == 'U' || source[index] == 'B') && index+1 < len(source) {
		if source[index+1] == '"' || source[index+1] == '\'' {
			quoteIndex = index + 1
		}
	}
	if quoteIndex >= len(source) || (source[quoteIndex] != '\'' && source[quoteIndex] != '"' && source[quoteIndex] != '`') {
		return 0, false
	}
	quote := source[quoteIndex]
	if quote == '`' {
		for cursor := quoteIndex + 1; cursor < len(source); cursor++ {
			if source[cursor] == '`' && !isEscaped(source, cursor) {
				return cursor + 1, true
			}
		}
		return len(source), true
	}
	triple := quoteIndex+2 < len(source) && source[quoteIndex+1] == quote && source[quoteIndex+2] == quote
	if triple {
		for cursor := quoteIndex + 3; cursor+2 < len(source); cursor++ {
			if source[cursor] == quote && source[cursor+1] == quote && source[cursor+2] == quote && !isEscaped(source, cursor) {
				return cursor + 3, true
			}
		}
		return len(source), true
	}
	if verbatim {
		for cursor := quoteIndex + 1; cursor < len(source); cursor++ {
			if source[cursor] != quote {
				continue
			}
			if cursor+1 < len(source) && source[cursor+1] == quote {
				cursor++
				continue
			}
			return cursor + 1, true
		}
		return len(source), true
	}
	for cursor := quoteIndex + 1; cursor < len(source); cursor++ {
		if source[cursor] == '\\' {
			cursor++
			continue
		}
		if source[cursor] == quote {
			return cursor + 1, true
		}
		if source[cursor] == '\n' && quote != '`' {
			return cursor, true
		}
	}
	return len(source), true
}

func skipCSharpRawString(source []byte, index int) (int, bool) {
	if index >= len(source) || source[index] != '"' {
		return 0, false
	}
	opening := 0
	for index+opening < len(source) && source[index+opening] == '"' {
		opening++
	}
	if opening < 3 {
		return 0, false
	}
	for cursor := index + opening; cursor < len(source); cursor++ {
		if source[cursor] != '"' {
			continue
		}
		closing := 0
		for cursor+closing < len(source) && source[cursor+closing] == '"' {
			closing++
		}
		if closing >= opening {
			return cursor + closing, true
		}
		cursor += closing - 1
	}
	return len(source), true
}

func skipHashString(source []byte, index int) (int, bool) {
	if index >= len(source) || source[index] != '#' {
		return 0, false
	}
	hashEnd := index
	for hashEnd < len(source) && source[hashEnd] == '#' {
		hashEnd++
	}
	if hashEnd >= len(source) || source[hashEnd] != '"' {
		return 0, false
	}
	triple := hashEnd+2 < len(source) && source[hashEnd+1] == '"' && source[hashEnd+2] == '"'
	closing := make([]byte, 0, hashEnd-index+3)
	if triple {
		closing = append(closing, '"', '"', '"')
	} else {
		closing = append(closing, '"')
	}
	closing = append(closing, source[index:hashEnd]...)
	contentStart := hashEnd + 1
	if triple {
		contentStart += 2
	}
	if end := bytes.Index(source[contentStart:], closing); end >= 0 {
		return contentStart + end + len(closing), true
	}
	return len(source), true
}

func skipRawString(source []byte, index int, language model.Language) (int, bool) {
	if language == "cpp" && index+1 < len(source) && source[index] == 'R' && source[index+1] == '"' {
		openEnd := bytes.IndexByte(source[index+2:], '(')
		if openEnd < 0 || openEnd > 16 {
			return 0, false
		}
		delimiter := source[index+2 : index+2+openEnd]
		for _, character := range delimiter {
			if character == ' ' || character == '\t' || character == '\r' || character == '\n' || character == ')' || character == '(' || character == '\\' {
				return 0, false
			}
		}
		closing := make([]byte, 0, len(delimiter)+2)
		closing = append(closing, ')')
		closing = append(closing, delimiter...)
		closing = append(closing, '"')
		contentStart := index + 2 + openEnd + 1
		if end := bytes.Index(source[contentStart:], closing); end >= 0 {
			return contentStart + end + len(closing), true
		}
		return len(source), true
	}
	if (language != "rust") || index >= len(source) || (source[index] != 'r' && source[index] != 'R') || index+1 >= len(source) {
		return 0, false
	}
	hashEnd := index + 1
	for hashEnd < len(source) && source[hashEnd] == '#' {
		hashEnd++
	}
	if hashEnd >= len(source) || source[hashEnd] != '"' {
		return 0, false
	}
	closing := make([]byte, 0, hashEnd-index+1)
	closing = append(closing, '"')
	closing = append(closing, source[index+1:hashEnd]...)
	contentStart := hashEnd + 1
	if end := bytes.Index(source[contentStart:], closing); end >= 0 {
		return contentStart + end + len(closing), true
	}
	return len(source), true
}

func findBlockEnd(source []byte, contentStart int, delimiter blockDelimiter, nested bool) int {
	depth := 1
	for index := contentStart; index < len(source); {
		if nested && bytes.HasPrefix(source[index:], []byte(delimiter.start)) {
			depth++
			index += len(delimiter.start)
			continue
		}
		if bytes.HasPrefix(source[index:], []byte(delimiter.end)) {
			depth--
			index += len(delimiter.end)
			if depth == 0 {
				return index
			}
			continue
		}
		index++
	}
	return len(source)
}

func skipShellHeredoc(source []byte, index int) (int, bool) {
	lineEnd := bytes.IndexByte(source[index:], '\n')
	if lineEnd < 0 {
		lineEnd = len(source) - index
	}
	line := source[index : index+lineEnd]
	markers := shellHeredocTerminators(line)
	if len(markers) == 0 {
		return 0, false
	}
	search := index + lineEnd
	if search < len(source) {
		search++
	}
	for _, marker := range markers {
		var found bool
		search, found = findShellHeredocEnd(source, search, marker.terminator, marker.stripTabs)
		if !found {
			return len(source), true
		}
	}
	return search, true
}

type shellHeredoc struct {
	terminator string
	stripTabs  bool
}

func findShellHeredocEnd(source []byte, search int, terminator string, stripTabs bool) (int, bool) {
	for search < len(source) {
		end := bytes.IndexByte(source[search:], '\n')
		if end < 0 {
			end = len(source) - search
		}
		candidate := strings.TrimSuffix(string(source[search:search+end]), "\r")
		if stripTabs {
			candidate = strings.TrimLeft(candidate, "\t")
		}
		if candidate == terminator {
			endIndex := search + end
			if endIndex < len(source) {
				endIndex++
			}
			return endIndex, true
		}
		search += end
		if search < len(source) {
			search++
		}
	}
	return len(source), false
}

func skipPHPHeredoc(source []byte, index int) (int, bool) {
	lineEnd := bytes.IndexByte(source[index:], '\n')
	if lineEnd < 0 {
		lineEnd = len(source) - index
	}
	line := source[index : index+lineEnd]
	marker := phpHeredocMarker(line)
	if marker < 0 {
		return 0, false
	}
	marker += 3
	if marker < len(line) && (line[marker] == '\'' || line[marker] == '"') {
		quote := line[marker]
		marker++
		start := marker
		for marker < len(line) && line[marker] != quote {
			marker++
		}
		if marker >= len(line) {
			return 0, false
		}
		terminator := string(line[start:marker])
		return findPHPHeredocEnd(source, index+lineEnd, terminator)
	}
	start := marker
	for marker < len(line) && (unicode.IsLetter(rune(line[marker])) || unicode.IsDigit(rune(line[marker])) || line[marker] == '_') {
		marker++
	}
	if marker == start {
		return 0, false
	}
	return findPHPHeredocEnd(source, index+lineEnd, string(line[start:marker]))
}

func phpHeredocMarker(line []byte) int {
	var quote byte
	escaped := false
	blockComment := false
	for index := 0; index+2 < len(line); index++ {
		character := line[index]
		if blockComment {
			if character == '*' && index+1 < len(line) && line[index+1] == '/' {
				blockComment = false
				index++
			}
			continue
		}
		if escaped {
			escaped = false
			continue
		}
		if quote != 0 {
			if character == '\\' && quote == '"' {
				escaped = true
			} else if character == quote {
				quote = 0
			}
			continue
		}
		if character == '\'' || character == '"' {
			quote = character
			continue
		}
		if character == '/' && index+1 < len(line) && line[index+1] == '*' {
			blockComment = true
			index++
			continue
		}
		if character == '#' || (character == '/' && index+1 < len(line) && line[index+1] == '/') {
			break
		}
		if character == '<' && line[index+1] == '<' && line[index+2] == '<' {
			return index
		}
	}
	return -1
}

func findPHPHeredocEnd(source []byte, lineEnd int, terminator string) (int, bool) {
	search := lineEnd
	if search < len(source) {
		search++
	}
	for search < len(source) {
		end := bytes.IndexByte(source[search:], '\n')
		if end < 0 {
			end = len(source) - search
		}
		candidate := strings.TrimSpace(strings.TrimSuffix(string(source[search:search+end]), "\r"))
		candidate = strings.TrimSuffix(candidate, ";")
		if candidate == terminator {
			endIndex := search + end
			if endIndex < len(source) {
				endIndex++
			}
			return endIndex, true
		}
		search += end
		if search < len(source) {
			search++
		}
	}
	return len(source), true
}

func shellHeredocTerminator(line []byte) (string, bool, bool) {
	markers := shellHeredocTerminators(line)
	if len(markers) == 0 {
		return "", false, false
	}
	return markers[0].terminator, markers[0].stripTabs, true
}

func shellHeredocTerminators(line []byte) []shellHeredoc {
	markers := make([]shellHeredoc, 0, 1)
	var quote byte
	escaped := false
	for index := 0; index+1 < len(line); index++ {
		character := line[index]
		if escaped {
			escaped = false
			continue
		}
		if character == '\\' {
			escaped = true
			continue
		}
		if character == '#' {
			break
		}
		if quote != 0 {
			if character == quote {
				quote = 0
			}
			continue
		}
		if character == '\'' || character == '"' {
			quote = character
			continue
		}
		if character != '<' || line[index+1] != '<' {
			continue
		}
		index += 2
		stripTabs := false
		if index < len(line) && line[index] == '-' {
			stripTabs = true
			index++
		}
		for index < len(line) && (line[index] == ' ' || line[index] == '\t') {
			index++
		}
		if index >= len(line) || line[index] == '<' {
			continue
		}
		terminator := ""
		if line[index] == '\\' && index+1 < len(line) {
			index++
			start := index
			for index < len(line) && isShellHeredocDelimiterChar(line[index]) {
				index++
			}
			if index == start {
				continue
			}
			terminator = string(line[start:index])
		}
		if terminator == "" && (line[index] == '\'' || line[index] == '"') {
			quote := line[index]
			index++
			start := index
			for index < len(line) && line[index] != quote {
				index++
			}
			if index >= len(line) {
				return markers
			}
			if index == start {
				return markers
			}
			terminator = string(line[start:index])
		}
		if terminator == "" {
			start := index
			for index < len(line) && isShellHeredocDelimiterChar(line[index]) {
				index++
			}
			if index == start {
				continue
			}
			terminator = string(line[start:index])
		}
		markers = append(markers, shellHeredoc{terminator: terminator, stripTabs: stripTabs})
	}
	return markers
}

func isShellHeredocDelimiterChar(character byte) bool {
	return unicode.IsLetter(rune(character)) || unicode.IsDigit(rune(character)) || strings.ContainsRune("_-.+=", rune(character))
}

func isEscaped(source []byte, index int) bool {
	backslashes := 0
	for index > 0 && source[index-1] == '\\' {
		backslashes++
		index--
	}
	return backslashes%2 == 1
}

func isShellCommentStart(source []byte, index int) bool {
	if index == 0 || source[index-1] == '\n' || unicode.IsSpace(rune(source[index-1])) {
		return true
	}
	switch source[index-1] {
	case ';', '|', '&', '(', ')', '{', '}', '<', '>':
		return true
	default:
		return false
	}
}

func isYAMLCommentStart(source []byte, index int) bool {
	return index == 0 || source[index-1] == '\n' || unicode.IsSpace(rune(source[index-1]))
}

func isSQLHashOperator(source []byte, index int) bool {
	if index+1 >= len(source) {
		return false
	}
	switch source[index+1] {
	case '>', '<', '-', '?':
		return true
	default:
		return false
	}
}

func findRubyBlockEnd(source []byte, index int) int {
	search := index
	for search < len(source) {
		lineEnd := bytes.IndexByte(source[search:], '\n')
		if lineEnd < 0 {
			lineEnd = len(source) - search
		}
		if search > index && bytes.HasPrefix(bytes.TrimSpace(source[search:search+lineEnd]), []byte("=end")) {
			end := search + lineEnd
			if end < len(source) {
				end++
			}
			return end
		}
		search += lineEnd
		if search < len(source) {
			search++
		}
	}
	return len(source)
}

func atLineStart(source []byte, index int) bool {
	return index == 0 || source[index-1] == '\n'
}

func positionAt(source []byte, offset int) model.Position {
	line, column := 1, 1
	for index := 0; index < offset && index < len(source); index++ {
		if source[index] == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	return model.Position{Line: line, Column: column}
}
