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

	if len(args) == 0 {
		printUsage()
		return nil
	}

	switch args[0] {
	case "run":
		mode, runArgs, err := parseRuntimeModeFlag(args[1:])
		if err != nil {
			return err
		}

		return runCommandTargetWithMode(mode, runArgs)

	case "dump":
		return runDumpCommand(args[1:])

	case "build":
		return runBuildCommand(args[1:])

	case "-h", "--help", "help":
		printUsage()
		return nil

	default:
		return fmt.Errorf("unknown command %q\n%s", args[0], commandUsage())
	}
}

func runCommandTargetWithMode(mode RuntimeMode, args []string) error {
	if len(args) > 0 && isEvalFlag(args[0]) {
		return runEvalCommandWithMode(mode, args)
	}

	switch len(args) {
	case 0:
		manifestPath, err := findManifestUpwards(".")
		if err != nil {
			return err
		}

		return runManifestFileWithMode(mode, manifestPath, nil)

	default:
		target := args[0]
		programArgs := args[1:]

		manifestPath, isManifest, err := manifestPathFromArgument(target)
		if err != nil {
			return err
		}

		if isManifest {
			return runManifestFileWithMode(mode, manifestPath, programArgs)
		}

		return runSourceFileWithMode(mode, target, programArgs)
	}
}

func runEvalCommandWithMode(mode RuntimeMode, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("%s requires a source string", args[0])
	}

	source := args[1]
	programArgs := args[2:]

	return runSourceStringWithMode(mode, source, "<eval>", programArgs)
}

func runSourceFileWithMode(mode RuntimeMode, filename string, args []string) error {
	switch mode {
	case RuntimeModeBytecode:
		return runSourceFileWithBytecode(filename, args)

	case RuntimeModeInterpreter:
		interpreter := NewInterpreterWithArgs(args)

		return runSourceFile(interpreter, filename)

	default:
		return fmt.Errorf("unknown runtime mode %d", mode)
	}
}

func runManifestFileWithMode(mode RuntimeMode, filename string, args []string) error {
	switch mode {
	case RuntimeModeBytecode:
		return runManifestFileWithBytecode(filename, args)

	case RuntimeModeInterpreter:
		return runManifestFileWithArgs(filename, args)

	default:
		return fmt.Errorf("unknown runtime mode %d", mode)
	}
}

func runSourceStringWithMode(
	mode RuntimeMode,
	source string,
	displayName string,
	args []string,
) error {
	switch mode {
	case RuntimeModeBytecode:
		vm := NewBytecodeVM(args)

		return runSourceStringWithBytecodeVM(vm, source, displayName)

	case RuntimeModeInterpreter:
		interpreter := NewInterpreterWithArgs(args)

		return runSourceString(interpreter, source, displayName)

	default:
		return fmt.Errorf("unknown runtime mode %d", mode)
	}
}

func runDumpCommand(args []string) error {
	dumpArgs, err := parseDumpArgs(args)
	if err != nil {
		return err
	}

	if len(dumpArgs) > 0 && isEvalFlag(dumpArgs[0]) {
		if len(dumpArgs) != 2 {
			return fmt.Errorf("Usage: stult dump [--bytecode] -e|--eval <source-string>")
		}

		return dumpSourceStringBytecode(dumpArgs[1], "<eval>")
	}

	switch len(dumpArgs) {
	case 0:
		manifestPath, err := findManifestUpwards(".")
		if err != nil {
			return err
		}

		return dumpTargetBytecode(manifestPath)

	case 1:
		return dumpTargetBytecode(dumpArgs[0])

	default:
		return fmt.Errorf("Usage: stult dump [--bytecode] [file.stult|directory|manifest]")
	}
}

func dumpTargetBytecode(target string) error {
	manifestPath, isManifest, err := manifestPathFromArgument(target)
	if err != nil {
		return err
	}

	if isManifest {
		return dumpManifestBytecode(manifestPath)
	}

	return dumpSourceFileBytecode(target)
}

func dumpSourceStringBytecode(source string, displayName string) error {
	chunk, err := compileSourceStringToBytecode(source, displayName)
	if err != nil {
		return err
	}

	fmt.Print(FormatBytecode(chunk))

	return nil
}

func dumpSourceFileBytecode(filename string) error {
	chunk, err := compileSourceFileToBytecode(filename)
	if err != nil {
		return err
	}

	fmt.Print(FormatBytecode(chunk))

	return nil
}

func dumpManifestBytecode(filename string) error {
	manifest, files, err := loadManifestFileFromFS(filename)
	if err != nil {
		return err
	}

	for index, runFile := range manifest.RunFiles {
		if index > 0 {
			fmt.Println()
		}

		var chunk *BytecodeChunk

		if filepath.IsAbs(runFile) {
			chunk, err = compileSourceFileToBytecode(runFile)
		} else {
			fsPath := cleanFSPath(runFile)
			chunk, err = compileSourceFromFSToBytecode(files, fsPath, runFile)
		}

		if err != nil {
			return err
		}

		fmt.Print(FormatBytecode(chunk))
	}

	return nil
}

func runManifestFile(filename string) error {
	return runManifestFileWithArgs(filename, nil)
}

func runManifestFileWithArgs(filename string, args []string) error {
	manifest, files, err := loadManifestFileFromFS(filename)
	if err != nil {
		return err
	}

	interpreter := NewInterpreterWithArgs(args)

	return runManifestFromFS(interpreter, files, manifest.RunFiles)
}

func runManifestFileWithBytecode(filename string, args []string) error {
	manifest, files, err := loadManifestFileFromFS(filename)
	if err != nil {
		return err
	}

	vm := NewBytecodeVM(args)

	return runBytecodeManifestFromFS(vm, files, manifest.RunFiles)
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

	for {
		stulton := filepath.Join(absoluteDir, ManifestStultonFilename)
		jsonManifest := filepath.Join(absoluteDir, ManifestJSONFilename)

		hasStulton, err := fileExists(stulton)
		if err != nil {
			return "", err
		}

		hasJSON, err := fileExists(jsonManifest)
		if err != nil {
			return "", err
		}

		if hasStulton && hasJSON {
			return "", fmt.Errorf(
				"Found both %q and %q in %q; use only one manifest file",
				ManifestStultonFilename,
				ManifestJSONFilename,
				absoluteDir,
			)
		}

		if hasStulton {
			return stulton, nil
		}

		if hasJSON {
			return jsonManifest, nil
		}

		parent := filepath.Dir(absoluteDir)
		if parent == absoluteDir {
			break
		}

		absoluteDir = parent
	}

	return "", fmt.Errorf(
		"Could not find %s or %s from %q upward",
		ManifestStultonFilename,
		ManifestJSONFilename,
		startDir,
	)
}

func findManifestInDirectory(directory string) (string, bool, error) {
	stulton := filepath.Join(directory, ManifestStultonFilename)
	jsonManifest := filepath.Join(directory, ManifestJSONFilename)

	hasStulton, err := fileExists(stulton)
	if err != nil {
		return "", false, err
	}

	hasJSON, err := fileExists(jsonManifest)
	if err != nil {
		return "", false, err
	}

	if hasStulton && hasJSON {
		return "", false, fmt.Errorf(
			"Found both %q and %q in %q; use only one manifest file",
			ManifestStultonFilename,
			ManifestJSONFilename,
			directory,
		)
	}

	if hasStulton {
		return stulton, true, nil
	}

	if hasJSON {
		return jsonManifest, true, nil
	}

	return "", false, nil
}

func fileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)

	if err == nil {
		if info.IsDir() {
			return false, fmt.Errorf("Expected %q to be a file, got directory", filename)
		}

		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, fmt.Errorf("Could not inspect %q: %w", filename, err)
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

func runBytecodeManifestFromFS(vm *BytecodeVM, files fs.FS, runFiles []string) error {
	for _, filename := range runFiles {
		if filepath.IsAbs(filename) {
			if err := runSourceFileWithBytecodeVM(vm, filename); err != nil {
				return err
			}

			continue
		}

		fsPath := cleanFSPath(filename)

		if err := runSourceFileFromFSWithBytecodeVM(vm, files, fsPath, filename); err != nil {
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

func runSourceFileWithBytecode(filename string, args []string) error {
	vm := NewBytecodeVM(args)

	return runSourceFileWithBytecodeVM(vm, filename)
}

func runSourceFileWithBytecodeVM(vm *BytecodeVM, filename string) error {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("Could not resolve source path %q: %w", filename, err)
	}

	files := os.DirFS(filepath.Dir(absolutePath))
	fsPath := filepath.Base(absolutePath)

	return runSourceFileFromFSWithBytecodeVM(vm, files, fsPath, absolutePath)
}

func runSourceFileFromFS(interpreter *Interpreter, files fs.FS, filename string) error {
	return runSourceFileFromFSNamed(interpreter, files, filename, filename)
}

func runSourceFileFromFSNamed(
	interpreter *Interpreter,
	files fs.FS,
	filename string,
	displayName string,
) error {
	fsPath := cleanFSPath(filename)

	source, err := readSourceFromFS(files, fsPath, displayName)
	if err != nil {
		return err
	}

	return runSourceString(interpreter, source, displayName)
}

func runSourceFileFromFSWithBytecodeVM(
	vm *BytecodeVM,
	files fs.FS,
	filename string,
	displayName string,
) error {
	source, err := readSourceFromFS(files, filename, displayName)
	if err != nil {
		return err
	}

	return runSourceStringWithBytecodeVM(vm, source, displayName)
}

func readSourceFromFS(files fs.FS, filename string, displayName string) (string, error) {
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
		return "", fmt.Errorf("Could not read %q: %w", displayName, err)
	}

	return string(sourceBytes), nil
}

func runSourceString(interpreter *Interpreter, source string, displayName string) error {
	lexer := NewLexer(source)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return formatParserErrors(displayName, source, parser.Errors())
	}

	if err := interpreter.EvalProgram(program); err != nil {
		return fmt.Errorf("Runtime error in %q: %w", displayName, err)
	}

	return nil
}

func runSourceStringWithBytecodeVM(vm *BytecodeVM, source string, displayName string) error {
	chunk, err := compileSourceStringToBytecode(source, displayName)
	if err != nil {
		return err
	}

	if _, err := vm.Run(chunk); err != nil {
		return fmt.Errorf("Bytecode runtime error in %q: %w", displayName, err)
	}

	return nil
}

func compileSourceFileToBytecode(filename string) (*BytecodeChunk, error) {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not resolve source path %q: %w", filename, err)
	}

	files := os.DirFS(filepath.Dir(absolutePath))
	fsPath := filepath.Base(absolutePath)

	return compileSourceFromFSToBytecode(files, fsPath, absolutePath)
}

func compileSourceFromFSToBytecode(
	files fs.FS,
	filename string,
	displayName string,
) (*BytecodeChunk, error) {
	source, err := readSourceFromFS(files, filename, displayName)
	if err != nil {
		return nil, err
	}

	return compileSourceStringToBytecode(source, displayName)
}

func compileSourceStringToBytecode(source string, displayName string) (*BytecodeChunk, error) {
	lexer := NewLexer(source)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return nil, formatParserErrors(displayName, source, parser.Errors())
	}

	chunk, err := CompileBytecode(program, displayName)
	if err != nil {
		return nil, fmt.Errorf("Could not compile bytecode for %q: %w", displayName, err)
	}

	return chunk, nil
}

func cleanFSPath(filename string) string {
	cleaned := filepath.Clean(filename)
	cleaned = filepath.ToSlash(cleaned)
	cleaned = strings.TrimPrefix(cleaned, "./")

	return cleaned
}
