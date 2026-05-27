package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStandaloneExamplesRun(t *testing.T) {
	examplesDir := examplesDirForTest(t)

	err := filepath.WalkDir(examplesDir, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		if filepath.Ext(filename) != expectedFileExtension {
			return nil
		}

		relativePath, err := filepath.Rel(examplesDir, filename)
		if err != nil {
			return err
		}

		relativePath = filepath.ToSlash(relativePath)

		if shouldSkipStandaloneExample(relativePath) {
			return nil
		}

		t.Run(relativePath, func(t *testing.T) {
			interpreter := NewInterpreter()

			if err := runSourceFile(interpreter, filename); err != nil {
				t.Fatalf("example failed: %v", err)
			}
		})

		return nil
	})
	if err != nil {
		t.Fatalf("could not walk examples directory: %v", err)
	}
}

func TestBoolExampleRunsWithBindings(t *testing.T) {
	examplesDir := examplesDirForTest(t)

	interpreter := NewInterpreter()

	files := []string{
		filepath.Join(examplesDir, "bool", "bindings.stult"),
		filepath.Join(examplesDir, "bool", "bool.stult"),
	}

	for _, filename := range files {
		if err := runSourceFile(interpreter, filename); err != nil {
			t.Fatalf("example %q failed: %v", filename, err)
		}
	}
}

func examplesDirForTest(t *testing.T) string {
	t.Helper()

	candidates := []string{
		filepath.Join("..", "examples"),
		"examples",
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate
		}
	}

	t.Fatal("could not find examples directory")
	return ""
}

func shouldSkipStandaloneExample(relativePath string) bool {
	if strings.HasPrefix(relativePath, "__ignore/") {
		return true
	}

	switch relativePath {
	case "bool/bindings.stult":
		return true

	case "bool/bool.stult":
		return true

	default:
		return false
	}
}
