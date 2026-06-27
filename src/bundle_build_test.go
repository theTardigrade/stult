package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type commandResult struct {
	Stdout string
	Stderr string
	Err    error
}

func TestBuildSingleSourceBundleRuns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping bundle/build integration test in short mode")
	}

	runner := buildStultRunnerForIntegrationTest(t)
	workDir := t.TempDir()

	sourcePath := filepath.Join(workDir, "hello.stult")
	writeFileForIntegrationTest(t, sourcePath, `STD.IO.OUTPUT.WRITE_LINE("single:", STD.SYSTEM.ARGS[0], ":", STD.SYSTEM.ARGS[1])`)

	cases := []struct {
		Name string
		Mode string
	}{
		{Name: "bytecode", Mode: "--bytecode"},
		{Name: "interpreter", Mode: "--interpreter"},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			outputPath := executablePathForIntegrationTest(filepath.Join(workDir, "single-"+tc.Name))

			buildResult := runCommandForIntegrationTest(
				t,
				workDir,
				runner,
				"build",
				tc.Mode,
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
		})
	}
}

func TestBuildManifestProjectBundleRunsFilesInOrder(t *testing.T) {
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

	cases := []struct {
		Name string
		Mode string
	}{
		{Name: "bytecode", Mode: "--bytecode"},
		{Name: "interpreter", Mode: "--interpreter"},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			outputPath := executablePathForIntegrationTest(filepath.Join(workDir, "project-"+tc.Name))

			buildResult := runCommandForIntegrationTest(
				t,
				workDir,
				runner,
				"build",
				tc.Mode,
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
		})
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

	outputPath := executablePathForIntegrationTest(filepath.Join(t.TempDir(), "stult-test-runner"))
	result := runCommandForIntegrationTest(t, "", "go", "build", "-o", outputPath, ".")
	if result.Err != nil {
		t.Fatalf("could not build stult test runner: %v\nstdout:\n%s\nstderr:\n%s", result.Err, result.Stdout, result.Stderr)
	}

	assertExecutableExistsForIntegrationTest(t, outputPath)

	return outputPath
}

func runCommandForIntegrationTest(
	t *testing.T,
	dir string,
	name string,
	args ...string,
) commandResult {
	t.Helper()

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

	info, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("expected executable %q to exist: %v", filename, err)
	}
	if info.IsDir() {
		t.Fatalf("expected executable %q to be a file", filename)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0 {
		t.Fatalf("expected executable %q to have at least one executable bit; mode is %s", filename, info.Mode())
	}
}

func executablePathForIntegrationTest(path string) string {
	if runtime.GOOS == "windows" && filepath.Ext(path) != ".exe" {
		return path + ".exe"
	}

	return path
}
