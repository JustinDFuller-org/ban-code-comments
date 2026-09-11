package discovery

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/JustinDFuller/ban-code-comments/internal/languages"
	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

type Config struct {
	Paths     []string
	Languages map[model.Language]bool
	Includes  []string
	Excludes  []string
	Debug     bool
}

type Candidate struct {
	Path     string
	RelPath  string
	Language model.Language
}

type Diagnostic struct {
	Path   string
	Reason string
}

type Result struct {
	Candidates []Candidate
	Skipped    int
	Debug      []Diagnostic
}

var fixedDirectories = map[string]bool{
	".git": true, ".hg": true, ".svn": true, "node_modules": true, "vendor": true,
	"dist": true, "build": true, "target": true, ".build": true, ".next": true,
	"coverage": true, "tmp": true, ".cache": true,
}

func Discover(config Config) (Result, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return Result{}, err
	}
	workingDirectory, err = filepath.Abs(workingDirectory)
	if err != nil {
		return Result{}, err
	}
	gitRoot := findGitRoot(workingDirectory)
	var candidates []Candidate
	var diagnostics []Diagnostic
	skipped := 0
	for _, inputPath := range config.Paths {
		absolutePath, err := filepath.Abs(inputPath)
		if err != nil {
			return Result{}, err
		}
		info, err := os.Stat(absolutePath)
		if err != nil {
			return Result{}, fmt.Errorf("stat %s: %w", inputPath, err)
		}
		if info.Mode().IsRegular() {
			candidate, reason := candidateFor(absolutePath, workingDirectory, gitRoot, config)
			if candidate != nil {
				candidates = append(candidates, *candidate)
			} else {
				skipped++
				if config.Debug {
					diagnostics = append(diagnostics, Diagnostic{Path: displayPath(absolutePath, workingDirectory), Reason: reason})
				}
			}
			continue
		}
		err = filepath.WalkDir(absolutePath, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				if path != absolutePath && fixedDirectories[entry.Name()] {
					if config.Debug {
						diagnostics = append(diagnostics, Diagnostic{Path: displayPath(path, workingDirectory), Reason: "excluded directory"})
					}
					skipped++
					return filepath.SkipDir
				}
				return nil
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			candidate, reason := candidateFor(path, workingDirectory, gitRoot, config)
			if candidate != nil {
				candidates = append(candidates, *candidate)
			} else {
				skipped++
				if config.Debug {
					diagnostics = append(diagnostics, Diagnostic{Path: displayPath(path, workingDirectory), Reason: reason})
				}
			}
			return nil
		})
		if err != nil {
			return Result{}, fmt.Errorf("walk %s: %w", inputPath, err)
		}
	}
	ignored, err := ignoredPaths(candidates, gitRoot)
	if err != nil {
		return Result{}, err
	}
	filtered := candidates[:0]
	for _, candidate := range candidates {
		if ignored[candidate.Path] {
			skipped++
			if config.Debug {
				diagnostics = append(diagnostics, Diagnostic{Path: candidate.RelPath, Reason: "matched .gitignore"})
			}
			continue
		}
		filtered = append(filtered, candidate)
	}
	return Result{Candidates: filtered, Skipped: skipped, Debug: diagnostics}, nil
}

func candidateFor(path, workingDirectory, gitRoot string, config Config) (*Candidate, string) {
	language, ok := languages.Lookup(path)
	if !ok {
		return nil, "unsupported file type"
	}
	if len(config.Languages) > 0 && !config.Languages[language] {
		return nil, "language filter"
	}
	relativePath := displayPath(path, workingDirectory)
	if matchesAny(config.Excludes, relativePath) {
		return nil, "exclude glob"
	}
	if len(config.Includes) > 0 && !matchesAny(config.Includes, relativePath) {
		return nil, "include glob"
	}
	return &Candidate{Path: path, RelPath: relativePath, Language: language}, ""
}

func findGitRoot(path string) string {
	command := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	output, err := command.Output()
	if err != nil {
		return ""
	}
	root, err := filepath.Abs(strings.TrimSpace(string(output)))
	if err != nil {
		return ""
	}
	return root
}

func ignoredPaths(candidates []Candidate, gitRoot string) (map[string]bool, error) {
	ignored := make(map[string]bool)
	if gitRoot == "" || len(candidates) == 0 {
		return ignored, nil
	}
	var input bytes.Buffer
	paths := make([]string, 0, len(candidates))
	relativeToAbsolute := make(map[string]string, len(candidates))
	resolvedRoot, err := filepath.EvalSymlinks(gitRoot)
	if err != nil {
		resolvedRoot = gitRoot
	}
	for _, candidate := range candidates {
		resolvedPath, resolveErr := filepath.EvalSymlinks(candidate.Path)
		if resolveErr != nil {
			resolvedPath = candidate.Path
		}
		relative, err := filepath.Rel(resolvedRoot, resolvedPath)
		if err != nil || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		relative = filepath.ToSlash(relative)
		paths = append(paths, relative)
		relativeToAbsolute[relative] = candidate.Path
		input.WriteString(relative)
		input.WriteByte(0)
	}
	if len(paths) == 0 {
		return ignored, nil
	}
	command := exec.Command("git", "-C", gitRoot, "check-ignore", "--stdin", "--no-index", "-z")
	command.Stdin = &input
	output, err := command.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return ignored, nil
		}
		return nil, fmt.Errorf("check .gitignore: %w", err)
	}
	for _, path := range bytes.Split(output, []byte{0}) {
		if len(path) == 0 {
			continue
		}
		if absolute, ok := relativeToAbsolute[string(path)]; ok {
			ignored[absolute] = true
		}
	}
	return ignored, nil
}

func displayPath(path, workingDirectory string) string {
	relative, err := filepath.Rel(workingDirectory, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(relative)
}

func matchesAny(patterns []string, path string) bool {
	for _, pattern := range patterns {
		for _, item := range strings.Split(pattern, ",") {
			if globMatch(strings.TrimSpace(item), path) {
				return true
			}
		}
	}
	return false
}

func globMatch(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)
	if pattern == "" {
		return false
	}
	if !strings.Contains(pattern, "/") {
		pattern = "**/" + pattern
	}
	expression := "^" + globExpression(pattern) + "$"
	matched, err := regexp.MatchString(expression, path)
	return err == nil && matched
}

func globExpression(pattern string) string {
	var result strings.Builder
	for index := 0; index < len(pattern); index++ {
		switch pattern[index] {
		case '*':
			if index+1 < len(pattern) && pattern[index+1] == '*' {
				index++
				if index+1 < len(pattern) && pattern[index+1] == '/' {
					index++
					result.WriteString("(?:.*/)?")
				} else {
					result.WriteString(".*")
				}
			} else {
				result.WriteString("[^/]*")
			}
		case '?':
			result.WriteString("[^/]")
		default:
			result.WriteString(regexp.QuoteMeta(string(pattern[index])))
		}
	}
	return result.String()
}
