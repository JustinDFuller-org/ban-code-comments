package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/JustinDFuller/ban-code-comments/internal/model"
)

func TestCLIExitCodesAndOptions(t *testing.T) {
	binary := buildBinary(t)
	root := t.TempDir()
	write := func(name, contents string) string {
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	write("clean.go", "package main\nvar value = \"// literal\"\n")
	write("bad.go", "package main\n// ordinary\n")
	write("bad.py", "value = 1 # ordinary\n")
	write("directive.go", "package main\n//go:generate tool\n")
	write("notes.txt", "not source\n")

	clean := runBinary(t, binary, root, "clean.go")
	if clean.exitCode != 0 {
		t.Fatalf("clean exit = %d, stderr = %s", clean.exitCode, clean.stderr)
	}
	var cleanResult model.Result
	if err := json.Unmarshal(clean.stdout, &cleanResult); err != nil {
		t.Fatalf("clean JSON: %v; stdout = %s", err, clean.stdout)
	}
	if len(cleanResult.Findings) != 0 {
		t.Fatalf("clean findings = %#v", cleanResult.Findings)
	}

	finding := runBinary(t, binary, root, "bad.go")
	if finding.exitCode != 1 {
		t.Fatalf("finding exit = %d, stderr = %s", finding.exitCode, finding.stderr)
	}

	text := runBinary(t, binary, root, "--format", "text", "bad.go")
	if text.exitCode != 1 || string(text.stdout) != "bad.go:2:1: // ordinary\n" {
		t.Fatalf("text result = exit %d, stdout %q", text.exitCode, text.stdout)
	}

	filtered := runBinary(t, binary, root, "--language", "python", "--include", "**/*.py", "--exclude", "bad.go", ".")
	if filtered.exitCode != 1 {
		t.Fatalf("filtered exit = %d, stdout = %s", filtered.exitCode, filtered.stdout)
	}

	directive := runBinary(t, binary, root, "--categories", "directive", "directive.go")
	if directive.exitCode != 1 {
		t.Fatalf("directive exit = %d, stdout = %s", directive.exitCode, directive.stdout)
	}

	debug := runBinary(t, binary, root, "--debug", "notes.txt")
	if debug.exitCode != 0 || !bytes.Contains(debug.stderr, []byte("unsupported file type")) {
		t.Fatalf("debug result = exit %d, stderr %q", debug.exitCode, debug.stderr)
	}

	invalid := runBinary(t, binary, root, "--format", "yaml", "clean.go")
	if invalid.exitCode != 2 {
		t.Fatalf("invalid exit = %d, stderr = %s", invalid.exitCode, invalid.stderr)
	}

	missing := runBinary(t, binary, root, "missing.go")
	if missing.exitCode != 2 {
		t.Fatalf("missing exit = %d, stderr = %s", missing.exitCode, missing.stderr)
	}
}

type commandResult struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

func runBinary(t *testing.T, binary, directory string, args ...string) commandResult {
	command := exec.Command(binary, args...)
	command.Dir = directory
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	return commandResult{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: exitCode}
}

func buildBinary(t *testing.T) string {
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(sourceFile)))
	binary := filepath.Join(t.TempDir(), "ban-code-comments")
	command := exec.Command("go", "build", "-o", binary, "./cmd/ban-code-comments")
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v: %s", err, output)
	}
	return binary
}
