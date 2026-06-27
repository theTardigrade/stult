package main

import (
	"strings"
	"testing"
)

func TestBuildCommandAcceptsOutputLongOption(t *testing.T) {
	target, output, options, err := parseBuildCommandArgs([]string{
		"--interpreter",
		"example_project",
		"--output",
		"example-tool",
	})
	if err != nil {
		t.Fatalf("parse build args failed: %v", err)
	}
	if target != "example_project" {
		t.Fatalf("unexpected target: %q", target)
	}
	if output != "example-tool" {
		t.Fatalf("unexpected output path: %q", output)
	}
	if options.RunBytecode {
		t.Fatal("expected --interpreter to select source/interpreter bundle mode")
	}
}

func TestBuildCommandAcceptsOutputShortOption(t *testing.T) {
	target, output, options, err := parseBuildCommandArgs([]string{
		"--bytecode",
		"example_project",
		"-o",
		"example-tool",
	})
	if err != nil {
		t.Fatalf("parse build args failed: %v", err)
	}
	if target != "example_project" {
		t.Fatalf("unexpected target: %q", target)
	}
	if output != "example-tool" {
		t.Fatalf("unexpected output path: %q", output)
	}
	if !options.RunBytecode {
		t.Fatal("expected --bytecode to select bytecode bundle mode")
	}
}

func TestBuildOutputOptionsRequirePath(t *testing.T) {
	for _, option := range []string{"-o", "--output"} {
		t.Run(option, func(t *testing.T) {
			_, _, _, err := parseBuildCommandArgs([]string{option})
			if err == nil {
				t.Fatalf("expected %s without a path to fail", option)
			}
			expected := "build option " + option + " requires an output path"
			if !strings.Contains(err.Error(), expected) {
				t.Fatalf("expected error to contain %q, got: %v", expected, err)
			}
		})
	}
}

func TestBuildUsageMentionsOutputLongOption(t *testing.T) {
	for name, usage := range map[string]string{
		"commandUsage": commandUsage(),
		"buildUsage":   buildUsage(),
	} {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(usage, "[-o|--output <output-executable>]") {
				t.Fatalf("expected usage to mention -o and --output, got:\n%s", usage)
			}
		})
	}
}
