package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type BuildBundleOptions struct {
	RunBytecode bool
}

func runBuildCommand(args []string) error {
	targetPath, outputPath, options, err := parseBuildCommandArgs(args)
	if err != nil {
		return err
	}

	if err := BuildBundleWithOptions(targetPath, outputPath, options); err != nil {
		return err
	}

	fmt.Printf("Built %s\n", outputPath)

	return nil
}

func parseBuildCommandArgs(args []string) (string, string, BuildBundleOptions, error) {
	targetPath := "."
	targetPathSet := false
	outputPath := ""
	options := BuildBundleOptions{
		RunBytecode: true,
	}
	modeSet := false

	for index := 0; index < len(args); index++ {
		arg := args[index]

		switch arg {
		case "--bytecode":
			if modeSet && !options.RunBytecode {
				return "", "", options, fmt.Errorf("build cannot use both --bytecode and --interpreter")
			}

			options.RunBytecode = true
			modeSet = true

		case "--interpreter":
			if modeSet && options.RunBytecode {
				return "", "", options, fmt.Errorf("build cannot use both --bytecode and --interpreter")
			}

			options.RunBytecode = false
			modeSet = true

		case "-o", "--output":
			if index+1 >= len(args) {
				return "", "", options, fmt.Errorf("build option %s requires an output path", arg)
			}

			outputPath = args[index+1]
			index++

		case "-h", "--help":
			return "", "", options, errors.New(buildUsage())

		default:
			if strings.HasPrefix(arg, "-") {
				return "", "", options, fmt.Errorf("unknown build option %q\n%s", arg, buildUsage())
			}

			if targetPathSet {
				return "", "", options, fmt.Errorf("build expected at most one project directory or source file\n%s", buildUsage())
			}

			targetPath = arg
			targetPathSet = true
		}
	}

	if outputPath == "" {
		outputPath = defaultBundleOutputPath(targetPath)
	}

	return targetPath, outputPath, options, nil
}

func buildUsage() string {
	return "Usage:\n" +
		"  stult build [--bytecode|--interpreter] [project-directory-or-file.stult] -o <output-executable>\n" +
		"  stult build [project-directory-or-file.stult]"
}

func defaultBundleOutputPath(targetPath string) string {
	cleaned := filepath.Clean(targetPath)
	name := filepath.Base(cleaned)

	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "stult-app"
	}

	if filepath.Ext(name) == expectedFileExtension {
		name = strings.TrimSuffix(name, filepath.Ext(name))
	}

	if runtime.GOOS == "windows" && filepath.Ext(name) != ".exe" {
		name += ".exe"
	}

	return name
}

func BuildBundle(projectDir string, outputPath string) error {
	return BuildBundleWithOptions(projectDir, outputPath, BuildBundleOptions{
		RunBytecode: true,
	})
}

func BuildBundleWithOptions(targetPath string, outputPath string, options BuildBundleOptions) error {
	absoluteTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("Could not resolve build target %q: %w", targetPath, err)
	}

	info, err := os.Stat(absoluteTargetPath)
	if err != nil {
		return fmt.Errorf("Could not inspect build target %q: %w", targetPath, err)
	}

	var archiveBytes []byte

	if info.IsDir() {
		if _, found, err := findManifestInDirectory(absoluteTargetPath); err != nil {
			return err
		} else if !found {
			return fmt.Errorf(
				"Project directory %q must contain %s or %s",
				absoluteTargetPath,
				ManifestStultonFilename,
				ManifestJSONFilename,
			)
		}

		archiveBytes, err = createProjectBundleArchiveWithOptions(absoluteTargetPath, options)
		if err != nil {
			return err
		}
	} else {
		if filepath.Ext(absoluteTargetPath) != expectedFileExtension {
			return fmt.Errorf(
				"Expected build target %q to be a project directory or %s source file",
				targetPath,
				expectedFileExtension,
			)
		}

		archiveBytes, err = createSingleSourceBundleArchive(absoluteTargetPath, options)
		if err != nil {
			return err
		}
	}

	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Could not locate current executable: %w", err)
	}

	runnerBytes, err := readExecutableWithoutBundle(executablePath)
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
