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
	manifestDirs := manifestExampleDirsForTest(t, examplesDir)

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

		if shouldSkipStandaloneExample(relativePath, manifestDirs) {
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

func TestManifestExamplesRun(t *testing.T) {
	examplesDir := examplesDirForTest(t)

	manifestFiles := manifestExampleFilesForTest(t, examplesDir)

	if len(manifestFiles) == 0 {
		t.Fatal("expected at least one manifest example")
	}

	for _, manifestFile := range manifestFiles {
		relativePath, err := filepath.Rel(examplesDir, manifestFile)
		if err != nil {
			t.Fatalf("could not make manifest path relative: %v", err)
		}

		relativePath = filepath.ToSlash(relativePath)

		t.Run(relativePath, func(t *testing.T) {
			if err := runManifestFile(manifestFile); err != nil {
				t.Fatalf("manifest example failed: %v", err)
			}
		})
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

func manifestExampleFilesForTest(t *testing.T, examplesDir string) []string {
	t.Helper()

	manifestFiles := []string{}

	err := filepath.WalkDir(examplesDir, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			if shouldSkipExampleDir(filename, examplesDir) {
				return filepath.SkipDir
			}

			return nil
		}

		if isManifestFilename(filename) {
			manifestFiles = append(manifestFiles, filename)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("could not walk examples directory for manifests: %v", err)
	}

	return manifestFiles
}

func manifestExampleDirsForTest(t *testing.T, examplesDir string) map[string]bool {
	t.Helper()

	manifestDirs := map[string]bool{}

	for _, manifestFile := range manifestExampleFilesForTest(t, examplesDir) {
		dir := filepath.Dir(manifestFile)

		relativeDir, err := filepath.Rel(examplesDir, dir)
		if err != nil {
			t.Fatalf("could not make manifest dir relative: %v", err)
		}

		relativeDir = filepath.ToSlash(relativeDir)

		if relativeDir == "." {
			relativeDir = ""
		}

		manifestDirs[relativeDir] = true
	}

	return manifestDirs
}

func shouldSkipStandaloneExample(relativePath string, manifestDirs map[string]bool) bool {
	if strings.HasPrefix(relativePath, "__ignore/") {
		return true
	}

	for manifestDir := range manifestDirs {
		if manifestDir == "" {
			continue
		}

		if relativePath == manifestDir || strings.HasPrefix(relativePath, manifestDir+"/") {
			return true
		}
	}

	return false
}

func shouldSkipExampleDir(filename string, examplesDir string) bool {
	relativePath, err := filepath.Rel(examplesDir, filename)
	if err != nil {
		return false
	}

	relativePath = filepath.ToSlash(relativePath)

	return relativePath == "__ignore"
}
