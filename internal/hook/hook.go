package hook

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/JustinDFuller/ban-code-comments/internal/languages"
	"github.com/JustinDFuller/ban-code-comments/internal/model"
	"github.com/JustinDFuller/ban-code-comments/internal/scanner"
)

type Mode string

const (
	ModeHard Mode = "hard"
	ModeWarn Mode = "warn"
)

type Event struct {
	SessionID    string          `json:"session_id"`
	TurnID       string          `json:"turn_id"`
	Transcript   string          `json:"transcript_path"`
	CWD          string          `json:"cwd"`
	EventName    string          `json:"hook_event_name"`
	ToolName     string          `json:"tool_name"`
	ToolInput    json.RawMessage `json:"tool_input"`
	ToolResponse json.RawMessage `json:"tool_response"`
	ToolCallID   string          `json:"tool_use_id"`
}

type Response struct {
	Decision           string              `json:"decision,omitempty"`
	Reason             string              `json:"reason,omitempty"`
	SystemMessage      string              `json:"systemMessage,omitempty"`
	HookSpecificOutput *HookSpecificOutput `json:"hookSpecificOutput,omitempty"`
}

type HookSpecificOutput struct {
	HookEventName     string `json:"hookEventName,omitempty"`
	AdditionalContext string `json:"additionalContext,omitempty"`
}

type OperationalError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Evaluation struct {
	Findings []model.Finding
	Error    *OperationalError
}

type fileChange struct {
	Path       string
	NewPath    string
	Source     []byte
	OldText    string
	NewText    string
	PatchLines []string
	Delete     bool
}

var selectedCategories = map[model.Category]bool{
	model.CategoryOrdinary:      true,
	model.CategoryDocumentation: true,
}

func ParseMode(value string) (Mode, error) {
	mode := Mode(strings.ToLower(strings.TrimSpace(value)))
	if mode != ModeHard && mode != ModeWarn {
		return "", fmt.Errorf("unsupported hook mode %q", value)
	}
	return mode, nil
}

func Run(input io.Reader, output io.Writer, mode Mode) error {
	decoder := json.NewDecoder(input)
	var event Event
	if err := decoder.Decode(&event); err != nil {
		return writeResponse(output, responseForError(mode, "invalid_event", fmt.Sprintf("could not decode hook event: %v", err)))
	}
	response := Process(event, mode)
	return writeResponse(output, response)
}

func Process(event Event, mode Mode) Response {
	if mode != ModeHard && mode != ModeWarn {
		return responseForError(mode, "invalid_mode", fmt.Sprintf("unsupported hook mode %q", mode))
	}

	eventName := strings.ToLower(strings.TrimSpace(event.EventName))
	toolName := strings.ToLower(strings.TrimSpace(event.ToolName))
	if eventName != "pretooluse" && eventName != "pre_tool_use" {
		return Response{}
	}
	if event.CWD == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return responseForError(mode, "cwd_unavailable", err.Error())
		}
		event.CWD = cwd
	}
	root, err := filepath.Abs(event.CWD)
	if err != nil {
		return responseForError(mode, "cwd_invalid", err.Error())
	}

	return processPre(event, toolName, root, mode)
}

func processPre(event Event, toolName, root string, mode Mode) Response {
	if toolName != "apply_patch" && toolName != "edit" && toolName != "write" && toolName != "write_file" && toolName != "file_write" {
		return Response{}
	}
	changes, err := reconstructChanges(event.ToolInput)
	if err != nil {
		return responseForError(mode, "proposal_unreadable", err.Error())
	}
	findings, err := evaluateChanges(root, changes)
	if err != nil {
		return responseForError(mode, "proposal_evaluation_failed", err.Error())
	}
	return responseForFindings(mode, findings)
}

func responseForFindings(mode Mode, findings []model.Finding) Response {
	if len(findings) == 0 {
		return Response{}
	}
	reason := formatFindings(findings)
	if mode == ModeHard {
		return Response{Decision: "block", Reason: reason}
	}
	guidance := reason + " Use Git history for history, pull-request descriptions for rationale, simplified code or a nearby README for complexity, and Markdown for general documentation."
	return Response{
		SystemMessage: guidance,
		HookSpecificOutput: &HookSpecificOutput{
			HookEventName:     "PreToolUse",
			AdditionalContext: guidance,
		},
	}
}

func responseForError(mode Mode, code, message string) Response {
	text := fmt.Sprintf("ban-code-comments hook operational error (%s): %s", code, message)
	if mode == ModeHard {
		return Response{Decision: "block", Reason: text}
	}
	return Response{SystemMessage: text}
}

func writeResponse(output io.Writer, response Response) error {
	encoder := json.NewEncoder(output)
	return encoder.Encode(response)
}

func formatFindings(findings []model.Finding) string {
	ordered := append([]model.Finding(nil), findings...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Path != ordered[j].Path {
			return ordered[i].Path < ordered[j].Path
		}
		if ordered[i].Range.Start.Line != ordered[j].Range.Start.Line {
			return ordered[i].Range.Start.Line < ordered[j].Range.Start.Line
		}
		return ordered[i].Range.Start.Column < ordered[j].Range.Start.Column
	})
	parts := make([]string, 0, len(ordered))
	for _, finding := range ordered {
		parts = append(parts, fmt.Sprintf("%s:%d:%d [%s] %s", finding.Path, finding.Range.Start.Line, finding.Range.Start.Column, finding.Category, strings.TrimSpace(finding.Text)))
	}
	return "ban-code-comments found newly introduced comments: " + strings.Join(parts, "; ")
}

func evaluateChanges(root string, changes []fileChange) ([]model.Finding, error) {
	findings := make([]model.Finding, 0)
	for _, change := range changes {
		oldPath, err := resolvePath(root, change.Path)
		if err != nil {
			return nil, err
		}
		oldSource, readErr := os.ReadFile(oldPath)
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return nil, readErr
		}
		newSource := change.Source
		if change.OldText != "" || change.NewText != "" {
			if !strings.Contains(string(oldSource), change.OldText) {
				return nil, fmt.Errorf("proposed edit could not find old text in %s", change.Path)
			}
			newSource = []byte(strings.Replace(string(oldSource), change.OldText, change.NewText, 1))
		}
		if len(change.PatchLines) > 0 {
			newSource, err = applyUpdatePatch(oldSource, change.PatchLines)
			if err != nil {
				return nil, fmt.Errorf("apply proposed patch to %s: %w", change.Path, err)
			}
		}
		if change.Delete {
			newSource = nil
		}
		oldFindingPath := displayPath(root, change.Path)
		oldFindings := findingsFor(oldFindingPath, oldSource)
		newPath := change.Path
		if change.NewPath != "" {
			if _, err := resolvePath(root, change.NewPath); err != nil {
				return nil, err
			}
			newPath = change.NewPath
		}
		newFindings := findingsFor(displayPath(root, newPath), newSource)
		findings = append(findings, newFindingsDifference(oldFindings, newFindings)...)
	}
	return findings, nil
}

func findingsFor(path string, source []byte) []model.Finding {
	language, ok := languages.Lookup(path)
	if !ok {
		return nil
	}
	return scanner.Scan(filepath.ToSlash(path), language, source, selectedCategories)
}

func newFindingsDifference(before, after []model.Finding) []model.Finding {
	counts := make(map[string]int, len(before))
	for _, finding := range before {
		counts[findingKey(finding)]++
	}
	result := make([]model.Finding, 0, len(after))
	for _, finding := range after {
		key := findingKey(finding)
		if counts[key] > 0 {
			counts[key]--
			continue
		}
		result = append(result, finding)
	}
	return result
}

func findingKey(finding model.Finding) string {
	return string(finding.Language) + "\x00" + string(finding.Category) + "\x00" + finding.Text
}

func resolvePath(root, name string) (string, error) {
	name = strings.TrimSpace(filepath.ToSlash(name))
	name = strings.TrimPrefix(name, "a/")
	name = strings.TrimPrefix(name, "b/")
	if name == "" {
		return "", fmt.Errorf("invalid proposed path %q", name)
	}
	path := filepath.Clean(filepath.Join(root, filepath.FromSlash(name)))
	if filepath.IsAbs(filepath.FromSlash(name)) {
		path = filepath.Clean(filepath.FromSlash(name))
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("proposed path escapes workspace: %q", name)
	}
	return path, nil
}

func displayPath(root, name string) string {
	path, err := resolvePath(root, name)
	if err != nil {
		return filepath.ToSlash(name)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(name)
	}
	return filepath.ToSlash(relative)
}

func reconstructChanges(raw json.RawMessage) ([]fileChange, error) {
	if len(raw) == 0 {
		return nil, errors.New("hook event has no tool input")
	}
	var inputText string
	if err := json.Unmarshal(raw, &inputText); err == nil {
		return parsePatch(inputText)
	}
	var input map[string]json.RawMessage
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, fmt.Errorf("tool input is not an object: %w", err)
	}
	if patch := rawString(input, "patch"); patch != "" {
		return parsePatch(patch)
	}
	for _, key := range []string{"input", "command"} {
		if value := rawString(input, key); strings.HasPrefix(strings.TrimSpace(value), "*** Begin Patch") {
			return parsePatch(value)
		}
	}
	path := firstString(input, "path", "file_path", "filePath", "filename")
	if path == "" {
		return nil, errors.New("tool input does not identify a file")
	}
	if content, ok := rawBytes(input, "content", "new_content", "newContent"); ok {
		return []fileChange{{Path: path, Source: content}}, nil
	}
	oldText, oldOK := rawStringOK(input, "old_string", "oldString")
	newText, newOK := rawStringOK(input, "new_string", "newString")
	if oldOK && newOK {
		return []fileChange{{Path: path, OldText: oldText, NewText: newText}}, nil
	}
	return nil, errors.New("tool input does not contain proposed file content")
}

func rawString(input map[string]json.RawMessage, key string) string {
	value, ok := input[key]
	if !ok {
		return ""
	}
	var result string
	if json.Unmarshal(value, &result) != nil {
		return ""
	}
	return result
}

func rawStringOK(input map[string]json.RawMessage, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := input[key]; ok {
			var result string
			if err := json.Unmarshal(value, &result); err == nil {
				return result, true
			}
		}
	}
	return "", false
}

func rawBytes(input map[string]json.RawMessage, keys ...string) ([]byte, bool) {
	for _, key := range keys {
		if value, ok := input[key]; ok {
			var result string
			if err := json.Unmarshal(value, &result); err == nil {
				return []byte(result), true
			}
		}
	}
	return nil, false
}

func firstString(input map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		if value := rawString(input, key); value != "" {
			return value
		}
	}
	return ""
}

func parsePatch(patch string) ([]fileChange, error) {
	lines := strings.Split(strings.ReplaceAll(patch, "\r\n", "\n"), "\n")
	start := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "*** Begin Patch" {
		start++
	}
	changes := make([]fileChange, 0)
	for start < len(lines) {
		line := lines[start]
		if strings.TrimSpace(line) == "*** End Patch" || line == "" {
			start++
			continue
		}
		switch {
		case strings.HasPrefix(line, "*** Add File: "):
			path := strings.TrimPrefix(line, "*** Add File: ")
			body, next := collectPatchBody(lines, start+1)
			changes = append(changes, fileChange{Path: path, Source: []byte(strings.Join(stripAddedLines(body), "\n"))})
			start = next
		case strings.HasPrefix(line, "*** Delete File: "):
			path := strings.TrimPrefix(line, "*** Delete File: ")
			changes = append(changes, fileChange{Path: path, Delete: true})
			start++
		case strings.HasPrefix(line, "*** Update File: "):
			path := strings.TrimPrefix(line, "*** Update File: ")
			body, next := collectPatchBody(lines, start+1)
			newPath := ""
			if len(body) > 0 && strings.HasPrefix(body[0], "*** Move to: ") {
				newPath = strings.TrimPrefix(body[0], "*** Move to: ")
				body = body[1:]
			}
			changes = append(changes, fileChange{Path: path, NewPath: newPath, PatchLines: body, Delete: false})
			start = next
		default:
			return nil, fmt.Errorf("unsupported patch line %q", line)
		}
	}
	return changes, nil
}

func collectPatchBody(lines []string, start int) ([]string, int) {
	end := start
	for end < len(lines) {
		if strings.HasPrefix(lines[end], "*** Add File: ") || strings.HasPrefix(lines[end], "*** Delete File: ") || strings.HasPrefix(lines[end], "*** Update File: ") || strings.TrimSpace(lines[end]) == "*** End Patch" {
			break
		}
		end++
	}
	return lines[start:end], end
}

func stripAddedLines(lines []string) []string {
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "+") {
			result = append(result, strings.TrimPrefix(line, "+"))
		}
	}
	return result
}

func applyUpdatePatch(source []byte, lines []string) ([]byte, error) {
	hasFinalNewline := bytes.HasSuffix(source, []byte("\n"))
	text := strings.TrimSuffix(string(source), "\n")
	oldLines := []string{}
	if text != "" {
		oldLines = strings.Split(text, "\n")
	}
	newLines := make([]string, 0, len(oldLines))
	cursor := 0
	seenOperation := false
	flush := func(hunk []string) error {
		if len(hunk) == 0 {
			return nil
		}
		oldBlock := make([]string, 0)
		newBlock := make([]string, 0)
		for _, line := range hunk {
			if line == "\\ No newline at end of file" || line == "" || strings.HasPrefix(line, "@@") {
				continue
			}
			if line[0] != ' ' && line[0] != '-' && line[0] != '+' {
				return fmt.Errorf("unsupported hunk line %q", line)
			}
			content := line[1:]
			switch line[0] {
			case ' ', '-':
				oldBlock = append(oldBlock, content)
			}
			switch line[0] {
			case ' ', '+':
				newBlock = append(newBlock, content)
			}
		}
		if len(oldBlock) == 0 && len(newBlock) == 0 {
			return nil
		}
		match := -1
		for index := cursor; index+len(oldBlock) <= len(oldLines); index++ {
			if equalLines(oldLines[index:index+len(oldBlock)], oldBlock) {
				match = index
				break
			}
		}
		if match < 0 {
			return fmt.Errorf("hunk context was not found")
		}
		newLines = append(newLines, oldLines[cursor:match]...)
		newLines = append(newLines, newBlock...)
		cursor = match + len(oldBlock)
		seenOperation = true
		return nil
	}
	current := make([]string, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			if err := flush(current); err != nil {
				return nil, err
			}
			current = current[:0]
		}
		current = append(current, line)
	}
	if err := flush(current); err != nil {
		return nil, err
	}
	if !seenOperation {
		return nil, errors.New("patch contains no file operations")
	}
	newLines = append(newLines, oldLines[cursor:]...)
	result := []byte(strings.Join(newLines, "\n"))
	if hasFinalNewline {
		result = append(result, '\n')
	}
	return result, nil
}

func equalLines(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
