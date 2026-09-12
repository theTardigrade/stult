package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type commandResult struct {
	Stdout string
	Stderr string
	Err    error
}

var (
	integrationRunnerOnce    sync.Once
	integrationRunnerPath    string
	integrationRunnerTempDir string
	integrationRunnerErr     error
)

func TestMain(m *testing.M) {
	code := m.Run()

	if integrationRunnerTempDir != "" {
		_ = os.RemoveAll(integrationRunnerTempDir)
	}

	os.Exit(code)
}

func TestBuildSingleSourceBundleRunsInBytecodeMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bundle/build integration test in short mode")
	}

	runner := buildStultRunnerForIntegrationTest(t)
	workDir := t.TempDir()

	sourcePath := filepath.Join(workDir, "hello.stult")
	writeFileForIntegrationTest(t, sourcePath, `STD.IO.OUTPUT.WRITE_LINE("single:", STD.SYSTEM.ARGS[0], ":", STD.SYSTEM.ARGS[1])`)

	outputPath := executablePathForIntegrationTest(filepath.Join(workDir, "single-bytecode"))

	buildResult := runCommandForIntegrationTest(
		t,
		workDir,
		runner,
		"build",
		"--bytecode",
		sourcePath,
		"-o",
		outputPath,
	)
	if buildResult.Err != nil {
		t.Fatalf("build failed: %v\nstdout:\n%s\nstderr:\n%s", buildResult.Err, buildResult.Stdout, buildResult.Stderr)
	}
	if buildResult.Stderr != "" {
		t.Fatalf("unexpected build stderr: %q", buildResult.Stderr)
	}
	if !strings.Contains(buildResult.Stdout, "Built "+outputPath) {
		t.Fatalf("unexpected build stdout: %q", buildResult.Stdout)
	}

	assertExecutableExistsForIntegrationTest(t, outputPath)

	runResult := runCommandForIntegrationTest(t, workDir, outputPath, "left", "right")
	if runResult.Err != nil {
		t.Fatalf("bundled executable failed: %v\nstdout:\n%s\nstderr:\n%s", runResult.Err, runResult.Stdout, runResult.Stderr)
	}
	if runResult.Stdout != "single:left:right\n" {
		t.Fatalf("unexpected bundled stdout: %q", runResult.Stdout)
	}
	if runResult.Stderr != "" {
		t.Fatalf("unexpected bundled stderr: %q", runResult.Stderr)
	}
}

func TestBuildManifestProjectBundleRunsFilesInOrderInInterpreterMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bundle/build integration test in short mode")
	}

	runner := buildStultRunnerForIntegrationTest(t)
	workDir := t.TempDir()
	projectDir := filepath.Join(workDir, "project")

	writeFileForIntegrationTest(t, filepath.Join(projectDir, ManifestStultonFilename), `{
	"RUN": {
		"src/setup.stult"
		"src/main.stult"
	}
}
`)
	writeFileForIntegrationTest(t, filepath.Join(projectDir, "src", "setup.stult"), `PREFIX : "manifest"`)
	writeFileForIntegrationTest(t, filepath.Join(projectDir, "src", "main.stult"), `STD.IO.OUTPUT.WRITE_LINE(PREFIX, ":", STD.SYSTEM.ARGS[0])`)

	outputPath := executablePathForIntegrationTest(filepath.Join(workDir, "project-interpreter"))

	buildResult := runCommandForIntegrationTest(
		t,
		workDir,
		runner,
		"build",
		"--interpreter",
		projectDir,
		"-o",
		outputPath,
	)
	if buildResult.Err != nil {
		t.Fatalf("build failed: %v\nstdout:\n%s\nstderr:\n%s", buildResult.Err, buildResult.Stdout, buildResult.Stderr)
	}
	if buildResult.Stderr != "" {
		t.Fatalf("unexpected build stderr: %q", buildResult.Stderr)
	}

	assertExecutableExistsForIntegrationTest(t, outputPath)

	runResult := runCommandForIntegrationTest(t, projectDir, outputPath, "arg")
	if runResult.Err != nil {
		t.Fatalf("bundled executable failed: %v\nstdout:\n%s\nstderr:\n%s", runResult.Err, runResult.Stdout, runResult.Stderr)
	}
	if runResult.Stdout != "manifest:arg\n" {
		t.Fatalf("unexpected bundled stdout: %q", runResult.Stdout)
	}
	if runResult.Stderr != "" {
		t.Fatalf("unexpected bundled stderr: %q", runResult.Stderr)
	}
}

func TestBuildDirectoryWithoutManifestFails(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bundle/build integration test in short mode")
	}

	runner := buildStultRunnerForIntegrationTest(t)
	workDir := t.TempDir()
	projectDir := filepath.Join(workDir, "missing-manifest")
	outputPath := executablePathForIntegrationTest(filepath.Join(workDir, "missing-manifest-app"))

	writeFileForIntegrationTest(t, filepath.Join(projectDir, "main.stult"), `STD.IO.OUTPUT.WRITE_LINE("should not build")`)

	result := runCommandForIntegrationTest(t, workDir, runner, "build", projectDir, "-o", outputPath)
	if result.Err == nil {
		t.Fatalf("expected build to fail, but it succeeded\nstdout:\n%s\nstderr:\n%s", result.Stdout, result.Stderr)
	}
	if result.Stdout != "" {
		t.Fatalf("unexpected stdout for failed build: %q", result.Stdout)
	}
	if !strings.Contains(result.Stderr, "must contain "+ManifestStultonFilename+" or "+ManifestJSONFilename) {
		t.Fatalf("unexpected stderr for failed build: %q", result.Stderr)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Fatalf("expected failed build not to create output %q; stat error: %v", outputPath, err)
	}
}

func buildStultRunnerForIntegrationTest(t *testing.T) string {
	t.Helper()

	integrationRunnerOnce.Do(func() {
		integrationRunnerTempDir, integrationRunnerErr = os.MkdirTemp("", "stult-integration-runner-*")
		if integrationRunnerErr != nil {
			return
		}

		integrationRunnerPath = executablePathForIntegrationTest(filepath.Join(integrationRunnerTempDir, "stult-test-runner"))
		result := runCommandForIntegrationTestWithoutTB("", "go", "build", "-o", integrationRunnerPath, ".")
		if result.Err != nil {
			integrationRunnerErr = fmt.Errorf(
				"could not build stult test runner: %w\nstdout:\n%s\nstderr:\n%s",
				result.Err,
				result.Stdout,
				result.Stderr,
			)
			return
		}

		if err := checkExecutableExistsForIntegrationTest(integrationRunnerPath); err != nil {
			integrationRunnerErr = err
		}
	})

	if integrationRunnerErr != nil {
		t.Fatal(integrationRunnerErr)
	}

	return integrationRunnerPath
}

func runCommandForIntegrationTest(
	t *testing.T,
	dir string,
	name string,
	args ...string,
) commandResult {
	t.Helper()

	return runCommandForIntegrationTestWithoutTB(dir, name, args...)
}

func runCommandForIntegrationTestWithoutTB(
	dir string,
	name string,
	args ...string,
) commandResult {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return commandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
}

func writeFileForIntegrationTest(t *testing.T, filename string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		t.Fatalf("could not create directory for %q: %v", filename, err)
	}

	if err := os.WriteFile(filename, []byte(contents), 0644); err != nil {
		t.Fatalf("could not write %q: %v", filename, err)
	}
}

func assertExecutableExistsForIntegrationTest(t *testing.T, filename string) {
	t.Helper()

	if err := checkExecutableExistsForIntegrationTest(filename); err != nil {
		t.Fatal(err)
	}
}

func checkExecutableExistsForIntegrationTest(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return fmt.Errorf("expected executable %q to exist: %w", filename, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected executable %q to be a file", filename)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0 {
		return fmt.Errorf("expected executable %q to have at least one executable bit; mode is %s", filename, info.Mode())
	}

	return nil
}

func executablePathForIntegrationTest(path string) string {
	if runtime.GOOS == "windows" && filepath.Ext(path) != ".exe" {
		return path + ".exe"
	}

	return path
}

func TestBuildManifestProjectBundleReadsNamedAssetsInBytecodeMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bundle/build integration test in short mode")
	}

	runner := buildStultRunnerForIntegrationTest(t)
	workDir := t.TempDir()
	projectDir := filepath.Join(workDir, "asset-project")
	writeNamedAssetProjectForIntegrationTest(t, projectDir)

	outputPath := executablePathForIntegrationTest(filepath.Join(workDir, "asset-bytecode"))

	buildResult := runCommandForIntegrationTest(
		t,
		workDir,
		runner,
		"build",
		"--bytecode",
		projectDir,
		"-o",
		outputPath,
	)
	if buildResult.Err != nil {
		t.Fatalf("build failed: %v\nstdout:\n%s\nstderr:\n%s", buildResult.Err, buildResult.Stdout, buildResult.Stderr)
	}

	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("could not remove project directory before bundled run: %v", err)
	}

	runResult := runCommandForIntegrationTest(t, workDir, outputPath)
	if runResult.Err != nil {
		t.Fatalf("bundled executable failed: %v\nstdout:\n%s\nstderr:\n%s", runResult.Err, runResult.Stdout, runResult.Stderr)
	}
	if runResult.Stdout != namedAssetProjectExpectedStdout() {
		t.Fatalf("unexpected bundled stdout: %q", runResult.Stdout)
	}
	if runResult.Stderr != "" {
		t.Fatalf("unexpected bundled stderr: %q", runResult.Stderr)
	}
}

func TestBuildManifestProjectBundleReadsNamedAssetsInInterpreterMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bundle/build integration test in short mode")
	}

	runner := buildStultRunnerForIntegrationTest(t)
	workDir := t.TempDir()
	projectDir := filepath.Join(workDir, "asset-project")
	writeNamedAssetProjectForIntegrationTest(t, projectDir)

	outputPath := executablePathForIntegrationTest(filepath.Join(workDir, "asset-interpreter"))

	buildResult := runCommandForIntegrationTest(
		t,
		workDir,
		runner,
		"build",
		"--interpreter",
		projectDir,
		"-o",
		outputPath,
	)
	if buildResult.Err != nil {
		t.Fatalf("build failed: %v\nstdout:\n%s\nstderr:\n%s", buildResult.Err, buildResult.Stdout, buildResult.Stderr)
	}

	if err := os.RemoveAll(projectDir); err != nil {
		t.Fatalf("could not remove project directory before bundled run: %v", err)
	}

	runResult := runCommandForIntegrationTest(t, workDir, outputPath)
	if runResult.Err != nil {
		t.Fatalf("bundled executable failed: %v\nstdout:\n%s\nstderr:\n%s", runResult.Err, runResult.Stdout, runResult.Stderr)
	}
	if runResult.Stdout != namedAssetProjectExpectedStdout() {
		t.Fatalf("unexpected bundled stdout: %q", runResult.Stdout)
	}
	if runResult.Stderr != "" {
		t.Fatalf("unexpected bundled stderr: %q", runResult.Stderr)
	}
}

func TestManifestProjectRunReadsNamedAssetsFromProjectDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping run integration test in short mode")
	}

	runner := buildStultRunnerForIntegrationTest(t)
	workDir := t.TempDir()
	projectDir := filepath.Join(workDir, "asset-project")
	writeNamedAssetProjectForIntegrationTest(t, projectDir)

	for _, mode := range []string{"--bytecode", "--interpreter"} {
		runResult := runCommandForIntegrationTest(t, workDir, runner, "run", mode, projectDir)
		if runResult.Err != nil {
			t.Fatalf("%s run failed: %v\nstdout:\n%s\nstderr:\n%s", mode, runResult.Err, runResult.Stdout, runResult.Stderr)
		}
		if runResult.Stdout != namedAssetProjectExpectedStdout() {
			t.Fatalf("unexpected %s stdout: %q", mode, runResult.Stdout)
		}
		if runResult.Stderr != "" {
			t.Fatalf("unexpected %s stderr: %q", mode, runResult.Stderr)
		}
	}
}

func writeNamedAssetProjectForIntegrationTest(t *testing.T, projectDir string) {
	t.Helper()

	writeFileForIntegrationTest(t, filepath.Join(projectDir, ManifestStultonFilename), `{
	"RUN": "main.stult"
	"ASSETS": {
		"CONFIG": "./data/config.stulton"
		"BINARY": "./data/bytes.bin"
		"TEMPLATES": "./templates"
	}
}
`)
	writeFileForIntegrationTest(t, filepath.Join(projectDir, "main.stult"), `READ : STD.FILE.BUNDLED.READ
WRITE_LINE : STD.IO.OUTPUT.WRITE_LINE
JOIN : STD.TYPE.STRING.JOIN
PARSE : STD.DATA.STULTON.PARSE

config : PARSE(READ("CONFIG"))
WRITE_LINE(config.label)
WRITE_LINE(READ("TEMPLATES", "help.txt"))
bytes : READ("BINARY", _, +)
WRITE_LINE(bytes[0], "-", bytes[1])
WRITE_LINE(JOIN(STD.FILE.BUNDLED.KEYS(), ","))
WRITE_LINE(JOIN(STD.FILE.BUNDLED.LIST("TEMPLATES"), ","))
`)
	writeFileForIntegrationTest(t, filepath.Join(projectDir, "data", "config.stulton"), `{"label": "example"}`)
	writeFileForIntegrationTest(t, filepath.Join(projectDir, "data", "bytes.bin"), string([]byte{65, 66}))
	writeFileForIntegrationTest(t, filepath.Join(projectDir, "templates", "help.txt"), "Help text.")
	writeFileForIntegrationTest(t, filepath.Join(projectDir, "templates", "emails", "welcome.txt"), "Welcome.")
}

func namedAssetProjectExpectedStdout() string {
	return "example\n" +
		"Help text.\n" +
		"65-66\n" +
		"BINARY,CONFIG,TEMPLATES\n" +
		"emails/welcome.txt,help.txt\n"
}
