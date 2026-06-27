package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDumpEvalWritesOutputFile(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "dumps", "eval.txt")

	stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
		return runDumpCommand([]string{"-o", outputPath, "-e", `STD.IO.OUTPUT.WRITE_LINE("hi")`})
	})
	if err != nil {
		t.Fatalf("dump eval failed: %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected dump with output file not to write stdout, got: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	contents := readFileForDumpTest(t, outputPath)
	if !strings.Contains(contents, "STULT BYTECODE DISASSEMBLY") {
		t.Fatalf("expected dump file to contain disassembly heading, got: %q", contents)
	}
	if !strings.Contains(contents, "chunk: <eval>") {
		t.Fatalf("expected dump file to contain eval chunk name, got: %q", contents)
	}
}

func TestDumpSourceFileOutputOptionCanFollowTarget(t *testing.T) {
	workDir := t.TempDir()
	sourcePath := filepath.Join(workDir, "program.stult")
	outputPath := filepath.Join(workDir, "program.dump.txt")

	if err := os.WriteFile(sourcePath, []byte(`value : 1 + 2`), 0644); err != nil {
		t.Fatalf("could not write source file: %v", err)
	}

	stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
		return runDumpCommand([]string{sourcePath, "--output", outputPath})
	})
	if err != nil {
		t.Fatalf("dump source file failed: %v", err)
	}
	if stdout != "" {
		t.Fatalf("expected dump with output file not to write stdout, got: %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	contents := readFileForDumpTest(t, outputPath)
	if !strings.Contains(contents, "STULT BYTECODE DISASSEMBLY") {
		t.Fatalf("expected dump file to contain disassembly heading, got: %q", contents)
	}
	if !strings.Contains(contents, "chunk: ") {
		t.Fatalf("expected dump file to contain a chunk name, got: %q", contents)
	}
}

func TestDumpOutputOptionRequiresPath(t *testing.T) {
	_, _, err := captureStdoutAndStderrForTest(t, func() error {
		return runDumpCommand([]string{"-o"})
	})
	if err == nil {
		t.Fatal("expected dump output option without a path to fail")
	}
	if !strings.Contains(err.Error(), "dump option -o requires an output path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func readFileForDumpTest(t *testing.T, filename string) string {
	t.Helper()

	contents, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("could not read %q: %v", filename, err)
	}

	return string(contents)
}
