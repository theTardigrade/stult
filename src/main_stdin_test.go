package main

import (
	"os"
	"strings"
	"testing"
)

func TestRunStdinTargetWithProgramArgs(t *testing.T) {
	source := `STD.IO.OUTPUT.WRITE_LINE(STD.SYSTEM.ARGS[0], ":", STD.SYSTEM.ARGS[1])`

	cases := []struct {
		Name string
		Mode RuntimeMode
	}{
		{Name: "bytecode", Mode: RuntimeModeBytecode},
		{Name: "interpreter", Mode: RuntimeModeInterpreter},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
				return runWithStdinForTest(t, source, func() error {
					return runCommandTargetWithMode(tc.Mode, []string{"-", "left", "right"})
				})
			})

			if err != nil {
				t.Fatalf("run stdin target failed: %v", err)
			}

			if stdout != "left:right\n" {
				t.Fatalf("unexpected stdout: %q", stdout)
			}

			if stderr != "" {
				t.Fatalf("unexpected stderr: %q", stderr)
			}
		})
	}
}

func TestDumpStdinTargetUsesStdinDisplayName(t *testing.T) {
	source := `STD.IO.OUTPUT.WRITE_LINE("hi")`

	stdout, stderr, err := captureStdoutAndStderrForTest(t, func() error {
		return runWithStdinForTest(t, source, func() error {
			return runDumpCommand([]string{"-"})
		})
	})

	if err != nil {
		t.Fatalf("dump stdin target failed: %v", err)
	}

	if !strings.Contains(stdout, "chunk: <stdin>") {
		t.Fatalf("expected dump output to use <stdin> display name, got: %q", stdout)
	}

	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func runWithStdinForTest(t *testing.T, input string, run func() error) error {
	t.Helper()

	oldStdin := os.Stdin

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create stdin pipe: %v", err)
	}

	if _, err := writer.WriteString(input); err != nil {
		_ = reader.Close()
		_ = writer.Close()
		t.Fatalf("could not write stdin input: %v", err)
	}

	if err := writer.Close(); err != nil {
		_ = reader.Close()
		t.Fatalf("could not close stdin writer: %v", err)
	}

	os.Stdin = reader
	defer func() {
		os.Stdin = oldStdin
		_ = reader.Close()
	}()

	return run()
}
