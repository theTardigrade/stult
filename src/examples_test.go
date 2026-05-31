package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var standaloneExamplesRequiringArgs = map[string]bool{
	"csv_to_json_converter.stult": true,
}

var examplesWithNondeterministicStdout = map[string]bool{
	"animated_sine_wave/manifest.stulton": true,
}

type exampleRunResult struct {
	Stdout string
	Stderr string
	Err    error
}

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
			compareExampleRunsForTest(
				t,
				relativePath,
				func() error {
					return runExampleSourceFileWithInterpreterForTest(filename)
				},
				func() error {
					return runExampleSourceFileWithBytecodeForTest(filename)
				},
			)
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
			compareExampleRunsForTest(
				t,
				relativePath,
				func() error {
					return runExampleManifestWithInterpreterForTest(manifestFile)
				},
				func() error {
					return runExampleManifestWithBytecodeForTest(manifestFile)
				},
			)
		})
	}
}

func compareExampleRunsForTest(
	t *testing.T,
	relativePath string,
	runInterpreter func() error,
	runBytecode func() error,
) {
	t.Helper()

	interpreterResult := captureExampleRunForTest(t, runInterpreter)
	bytecodeResult := captureExampleRunForTest(t, runBytecode)

	if !exampleErrorsMatch(interpreterResult.Err, bytecodeResult.Err) {
		t.Fatalf(
			"interpreter and bytecode errors differed\n\ninterpreter error:\n%s\n\nbytecode error:\n%s\n\ninterpreter stdout:\n%s\n\nbytecode stdout:\n%s\n\ninterpreter stderr:\n%s\n\nbytecode stderr:\n%s",
			exampleErrorString(interpreterResult.Err),
			exampleErrorString(bytecodeResult.Err),
			interpreterResult.Stdout,
			bytecodeResult.Stdout,
			interpreterResult.Stderr,
			bytecodeResult.Stderr,
		)
	}

	if shouldCompareExampleStdout(relativePath) && interpreterResult.Stdout != bytecodeResult.Stdout {
		t.Fatalf(
			"interpreter and bytecode stdout differed\n\n%s",
			formatOutputDifferenceForTest(interpreterResult.Stdout, bytecodeResult.Stdout),
		)
	}

	if interpreterResult.Stderr != bytecodeResult.Stderr {
		t.Fatalf(
			"interpreter and bytecode stderr differed\n\n%s",
			formatOutputDifferenceForTest(interpreterResult.Stderr, bytecodeResult.Stderr),
		)
	}
}

func shouldCompareExampleStdout(relativePath string) bool {
	return !examplesWithNondeterministicStdout[relativePath]
}

func captureExampleRunForTest(t *testing.T, run func() error) exampleRunResult {
	t.Helper()

	stdout, stderr, err := captureStdoutAndStderrForTest(t, run)

	return exampleRunResult{
		Stdout: stdout,
		Stderr: stderr,
		Err:    err,
	}
}

func captureStdoutAndStderrForTest(
	t *testing.T,
	run func() error,
) (string, string, error) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create stdout pipe: %v", err)
	}

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		_ = stdoutReader.Close()
		_ = stdoutWriter.Close()
		t.Fatalf("could not create stderr pipe: %v", err)
	}

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer

	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)

	go func() {
		_, err := io.Copy(&stdoutBuffer, stdoutReader)
		stdoutDone <- err
	}()

	go func() {
		_, err := io.Copy(&stderrBuffer, stderrReader)
		stderrDone <- err
	}()

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	runErr := run()

	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()

	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdoutCopyErr := <-stdoutDone
	stderrCopyErr := <-stderrDone

	_ = stdoutReader.Close()
	_ = stderrReader.Close()

	if stdoutCopyErr != nil {
		t.Fatalf("could not capture stdout: %v", stdoutCopyErr)
	}

	if stderrCopyErr != nil {
		t.Fatalf("could not capture stderr: %v", stderrCopyErr)
	}

	return stdoutBuffer.String(), stderrBuffer.String(), runErr
}

func exampleErrorsMatch(left error, right error) bool {
	if left == nil || right == nil {
		return left == right
	}

	return left.Error() == right.Error()
}

func exampleErrorString(err error) string {
	if err == nil {
		return "<nil>"
	}

	return err.Error()
}

func formatOutputDifferenceForTest(interpreterOutput string, bytecodeOutput string) string {
	differenceIndex := firstDifferentByteIndex(interpreterOutput, bytecodeOutput)

	return fmt.Sprintf(
		"first difference at byte %d\n\ninterpreter length: %d\nbytecode length: %d\n\ninterpreter excerpt:\n%q\n\nbytecode excerpt:\n%q",
		differenceIndex,
		len(interpreterOutput),
		len(bytecodeOutput),
		excerptAroundByteIndex(interpreterOutput, differenceIndex),
		excerptAroundByteIndex(bytecodeOutput, differenceIndex),
	)
}

func firstDifferentByteIndex(left string, right string) int {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}

	for index := 0; index < limit; index++ {
		if left[index] != right[index] {
			return index
		}
	}

	return limit
}

func excerptAroundByteIndex(text string, index int) string {
	if index < 0 {
		index = 0
	}

	start := index - 120
	if start < 0 {
		start = 0
	}

	end := index + 120
	if end > len(text) {
		end = len(text)
	}

	return text[start:end]
}

func runExampleSourceFileWithInterpreterForTest(filename string) error {
	interpreter := NewInterpreter()

	return runSourceFile(interpreter, filename)
}

func runExampleManifestWithInterpreterForTest(filename string) error {
	return runManifestFile(filename)
}

func runExampleSourceFileWithBytecodeForTest(filename string) error {
	vm := NewBytecodeVM(nil)

	return runExampleSourceFileWithBytecodeVMForTest(vm, filename)
}

func runExampleSourceFileWithBytecodeVMForTest(vm *BytecodeVM, filename string) error {
	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return fmt.Errorf("could not resolve source path %q: %w", filename, err)
	}

	files := os.DirFS(filepath.Dir(absolutePath))
	fsPath := filepath.Base(absolutePath)

	return runExampleSourceFileFromFSWithBytecodeVMForTest(vm, files, fsPath, absolutePath)
}

func runExampleManifestWithBytecodeForTest(filename string) error {
	manifest, files, err := loadManifestFileFromFS(filename)
	if err != nil {
		return err
	}

	vm := NewBytecodeVM(nil)

	for _, runFile := range manifest.RunFiles {
		if filepath.IsAbs(runFile) {
			if err := runExampleSourceFileWithBytecodeVMForTest(vm, runFile); err != nil {
				return err
			}

			continue
		}

		fsPath := cleanFSPath(runFile)

		if err := runExampleSourceFileFromFSWithBytecodeVMForTest(
			vm,
			files,
			fsPath,
			runFile,
		); err != nil {
			return err
		}
	}

	return nil
}

func runExampleSourceFileFromFSWithBytecodeVMForTest(
	vm *BytecodeVM,
	files fs.FS,
	filename string,
	displayName string,
) error {
	fsPath := cleanFSPath(filename)

	sourceBytes, err := fs.ReadFile(files, fsPath)
	if err != nil {
		return fmt.Errorf("could not read %q: %w", displayName, err)
	}

	source := string(sourceBytes)

	lexer := NewLexer(source)
	parser := NewParser(lexer)
	program := parser.ParseProgram()

	if len(parser.Errors()) > 0 {
		return formatParserErrors(displayName, source, parser.Errors())
	}

	chunk, err := CompileBytecode(program, displayName)
	if err != nil {
		return fmt.Errorf("could not compile bytecode for %q: %w", displayName, err)
	}

	if _, err := vm.Run(chunk); err != nil {
		return fmt.Errorf("bytecode runtime error in %q: %w", displayName, err)
	}

	return nil
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

	if standaloneExamplesRequiringArgs[relativePath] {
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
