package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func runBuildCommand(args []string) error {
	projectDir, outputPath, err := parseBuildCommandArgs(args)
	if err != nil {
		return err
	}

	if err := BuildBundle(projectDir, outputPath); err != nil {
		return err
	}

	fmt.Printf("Built %s\n", outputPath)

	return nil
}

func parseBuildCommandArgs(args []string) (string, string, error) {
	projectDir := "."
	projectDirSet := false
	outputPath := ""

	for index := 0; index < len(args); index++ {
		arg := args[index]

		switch arg {
		case "-o", "--output":
			if index+1 >= len(args) {
				return "", "", fmt.Errorf("build option %s requires an output path", arg)
			}

			outputPath = args[index+1]
			index++

		case "-h", "--help":
			return "", "", errors.New(buildUsage())

		default:
			if strings.HasPrefix(arg, "-") {
				return "", "", fmt.Errorf("unknown build option %q\n%s", arg, buildUsage())
			}

			if projectDirSet {
				return "", "", fmt.Errorf("build expected at most one project directory\n%s", buildUsage())
			}

			projectDir = arg
			projectDirSet = true
		}
	}

	if outputPath == "" {
		outputPath = defaultBundleOutputPath(projectDir)
	}

	return projectDir, outputPath, nil
}

func buildUsage() string {
	return "Usage:\n" +
		"  stult build [project-directory] -o <output-executable>\n" +
		"  stult build [project-directory]"
}

func defaultBundleOutputPath(projectDir string) string {
	cleaned := filepath.Clean(projectDir)
	name := filepath.Base(cleaned)

	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "stult-app"
	}

	if runtime.GOOS == "windows" && filepath.Ext(name) != ".exe" {
		name += ".exe"
	}

	return name
}

func BuildBundle(projectDir string, outputPath string) error {
	absoluteProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		return fmt.Errorf("Could not resolve project directory %q: %w", projectDir, err)
	}

	info, err := os.Stat(absoluteProjectDir)
	if err != nil {
		return fmt.Errorf("Could not inspect project directory %q: %w", projectDir, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("Expected project directory %q to be a directory", projectDir)
	}

	if _, found, err := findManifestInDirectory(absoluteProjectDir); err != nil {
		return err
	} else if !found {
		return fmt.Errorf(
			"Project directory %q must contain %s or %s",
			absoluteProjectDir,
			ManifestStultonFilename,
			ManifestJSONFilename,
		)
	}

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Could not locate current executable: %w", err)
	}

	runnerBytes, err := readExecutableWithoutBundle(executablePath)
	if err != nil {
		return err
	}

	archiveBytes, err := createProjectBundleArchive(absoluteProjectDir)
	if err != nil {
		return err
	}

	absoluteOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("Could not resolve output path %q: %w", outputPath, err)
	}

	absoluteExecutablePath, err := filepath.Abs(executablePath)
	if err == nil && absoluteOutputPath == absoluteExecutablePath {
		return fmt.Errorf("Refusing to overwrite the running executable %q", absoluteOutputPath)
	}

	if err := os.MkdirAll(filepath.Dir(absoluteOutputPath), 0755); err != nil {
		return fmt.Errorf("Could not create output directory: %w", err)
	}

	outputFile, err := os.OpenFile(absoluteOutputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("Could not create output executable %q: %w", outputPath, err)
	}
	defer outputFile.Close()

	if _, err := outputFile.Write(runnerBytes); err != nil {
		return fmt.Errorf("Could not write runner executable: %w", err)
	}

	if _, err := outputFile.Write(archiveBytes); err != nil {
		return fmt.Errorf("Could not write embedded bundle archive: %w", err)
	}

	footer := makeBundleFooter(uint64(len(archiveBytes)))

	if _, err := outputFile.Write(footer); err != nil {
		return fmt.Errorf("Could not write embedded bundle footer: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := outputFile.Chmod(0755); err != nil {
			return fmt.Errorf("Could not mark output executable as executable: %w", err)
		}
	}

	return nil
}

func readExecutableWithoutBundle(filename string) ([]byte, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not read runner executable %q: %w", filename, err)
	}

	bundleStart, found, err := embeddedBundleStart(bytes)
	if err != nil {
		return nil, err
	}

	if found {
		return bytes[:bundleStart], nil
	}

	return bytes, nil
}
