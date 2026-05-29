package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
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
	if handled, err := runEmbeddedBundleIfPresent(); handled || err != nil {
		return err
	}

	args := os.Args[1:]

	if len(args) > 0 && args[0] == "build" {
		return runBuildCommand(args[1:])
	}

	if len(args) > 0 && isEvalOption(args[0]) {
		if len(args) != 2 {
			return fmt.Errorf("%s", usage())
		}

		interpreter := NewInterpreter()
		return runSourceString(interpreter, args[1], "<eval>")
	}

	switch len(args) {
	case 0:
		manifestPath, err := findManifestUpwards(".")
		if err != nil {
			return err
		}

		return runManifestFile(manifestPath)

	case 1:
		target := args[0]

		manifestPath, isManifest, err := manifestPathFromArgument(target)
		if err != nil {
			return err
		}

		if isManifest {
			return runManifestFile(manifestPath)
		}

		interpreter := NewInterpreter()
		return runSourceFile(interpreter, target)

	default:
		return fmt.Errorf("%s", usage())
	}
}

func isEvalOption(arg string) bool {
	return arg == "-e" || arg == "--eval"
}

func usage() string {
	return "Usage:\n" +
		"  stult\n" +
		"  stult -e <source>\n" +
		"  stult --eval <source>\n" +
		"  stult build [project-directory] -o <output-executable>\n" +
		"  stult <file.stult>\n" +
		"  stult <directory>\n" +
		"  stult <manifest.stulton>\n" +
		"  stult <manifest.json>"
}

func runManifestFile(filename string) error {
	manifest, files, err := loadManifestFileFromFS(filename)
	if err != nil {
		return err
	}

	interpreter := NewInterpreter()

	return runManifestFromFS(interpreter, files, manifest.RunFiles)
}

func loadManifestFileFromFS(filename string) (*Manifest, fs.FS, error) {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("Could not resolve manifest path %q: %w", filename, err)
	}

	dir := filepath.Dir(absolutePath)
	base := filepath.Base(absolutePath)

	files := os.DirFS(dir)

	manifest, err := LoadManifestFromFS(files, base)
	if err != nil {
		return nil, nil, err
	}

	return manifest, files, nil
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

func runManifestFromFS(interpreter *Interpreter, files fs.FS, runFiles []string) error {
	for _, filename := range runFiles {
		if filepath.IsAbs(filename) {
			if err := runSourceFile(interpreter, filename); err != nil {
				return err
			}

			continue
		}

		fsPath := cleanFSPath(filename)

		if err := runSourceFileFromFS(interpreter, files, fsPath); err != nil {
			return err
		}
	}

	return nil
}

func runSourceFile(interpreter *Interpreter, filename string) error {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("Could not resolve source path %q: %w", filename, err)
	}

	files := os.DirFS(filepath.Dir(absolutePath))
	fsPath := filepath.Base(absolutePath)

	return runSourceFileFromFSNamed(interpreter, files, fsPath, absolutePath)
}

func runSourceFileFromFS(interpreter *Interpreter, files fs.FS, filename string) error {
	return runSourceFileFromFSNamed(interpreter, files, filename, filename)
}

func runSourceFileFromFSNamed(interpreter *Interpreter, files fs.FS, filename string, displayName string) error {
	fsPath := cleanFSPath(filename)

	if path.Ext(fsPath) != expectedFileExtension {
		fmt.Fprintf(
			os.Stderr,
			"Warning: Expected %s file extension, got %q\n",
			expectedFileExtension,
			path.Ext(fsPath),
		)
	}

	sourceBytes, err := fs.ReadFile(files, fsPath)
	if err != nil {
		return fmt.Errorf("Could not read %q: %w", displayName, err)
	}

	return runSourceString(interpreter, string(sourceBytes), displayName)
}

func runSourceString(interpreter *Interpreter, source string, displayName string) error {
	lexer := NewLexer(source)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return formatParserErrors(displayName, source, parser.Errors())
	}

	if err := interpreter.EvalProgram(program); err != nil {
		return formatRuntimeError(displayName, source, err)
	}

	return nil
}

func cleanFSPath(filename string) string {
	cleaned := filepath.Clean(filename)
	cleaned = filepath.ToSlash(cleaned)
	cleaned = strings.TrimPrefix(cleaned, "./")

	return cleaned
}
