package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const expectedFileExtension = ".stult"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]

	switch len(args) {
	case 0:
		manifestPath, err := findManifestUpwards(".")
		if err != nil {
			return err
		}

		manifest, err := LoadManifest(manifestPath)
		if err != nil {
			return err
		}

		return runManifest(manifest)

	case 1:
		target := args[0]

		manifestPath, isManifest, err := manifestPathFromArgument(target)
		if err != nil {
			return err
		}

		if isManifest {
			manifest, err := LoadManifest(manifestPath)
			if err != nil {
				return err
			}

			return runManifest(manifest)
		}

		interpreter := NewInterpreter()
		return runSourceFile(interpreter, target)

	default:
		return fmt.Errorf(
			"Usage:\n" +
				"  interpreter\n" +
				"  interpreter <file.stult>\n" +
				"  interpreter <directory>\n" +
				"  interpreter <manifest.stult>\n" +
				"  interpreter <manifest.json>",
		)
	}
}

func manifestPathFromArgument(target string) (string, bool, error) {
	info, err := os.Stat(target)

	if err == nil {
		if info.IsDir() {
			manifestPath, err := findManifestUpwards(target)
			if err != nil {
				return "", false, err
			}

			return manifestPath, true, nil
		}

		if isManifestFilename(target) {
			return target, true, nil
		}

		return "", false, nil
	}

	if os.IsNotExist(err) {
		if isManifestFilename(target) {
			return target, true, nil
		}

		return "", false, nil
	}

	return "", false, fmt.Errorf("Could not inspect %q: %w", target, err)
}

func findManifestUpwards(startDir string) (string, error) {
	absoluteDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("Could not resolve directory %q: %w", startDir, err)
	}

	info, err := os.Stat(absoluteDir)
	if err != nil {
		return "", fmt.Errorf("Could not inspect directory %q: %w", startDir, err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("Expected %q to be a directory", startDir)
	}

	dir := absoluteDir

	for {
		manifestPath, found, err := findManifestInDirectory(dir)
		if err != nil {
			return "", err
		}

		if found {
			return manifestPath, nil
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf(
		"Could not find %s or %s in %q or any parent directory",
		ManifestStultonFilename,
		ManifestJSONFilename,
		absoluteDir,
	)
}

func findManifestInDirectory(dir string) (string, bool, error) {
	stultonPath := filepath.Join(dir, ManifestStultonFilename)
	jsonPath := filepath.Join(dir, ManifestJSONFilename)

	hasStulton, err := manifestFileExists(stultonPath)
	if err != nil {
		return "", false, err
	}

	hasJSON, err := manifestFileExists(jsonPath)
	if err != nil {
		return "", false, err
	}

	if hasStulton && hasJSON {
		return "", false, fmt.Errorf(
			"Found both %q and %q in %q; use only one manifest file",
			ManifestStultonFilename,
			ManifestJSONFilename,
			dir,
		)
	}

	if hasStulton {
		return stultonPath, true, nil
	}

	if hasJSON {
		return jsonPath, true, nil
	}

	return "", false, nil
}

func manifestFileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)

	if err == nil {
		if info.IsDir() {
			return false, fmt.Errorf("Expected manifest %q to be a file, got directory", filename)
		}

		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("Could not inspect manifest %q: %w", filename, err)
}

func isManifestFilename(filename string) bool {
	base := filepath.Base(filename)

	return base == ManifestStultonFilename || base == ManifestJSONFilename
}

func runManifest(manifest *Manifest) error {
	interpreter := NewInterpreter()

	for _, filename := range manifest.RunFiles {
		if err := runSourceFile(interpreter, filename); err != nil {
			return err
		}
	}

	return nil
}

func runSourceFile(interpreter *Interpreter, filename string) error {
	if filepath.Ext(filename) != expectedFileExtension {
		fmt.Fprintf(
			os.Stderr,
			"Warning: Expected %s file extension, got %q\n",
			expectedFileExtension,
			filepath.Ext(filename),
		)
	}

	sourceBytes, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Could not read %q: %w", filename, err)
	}

	lexer := NewLexer(string(sourceBytes))
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return formatParserErrors(filename, parser.Errors())
	}

	if err := interpreter.EvalProgram(program); err != nil {
		return fmt.Errorf("Runtime error in %q: %w", filename, err)
	}

	return nil
}

func formatParserErrors(filename string, errors []string) error {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Parser errors in %q:", filename)

	for _, err := range errors {
		fmt.Fprintf(&builder, "\n  - %s", err)
	}

	return fmt.Errorf("%s", builder.String())
}
