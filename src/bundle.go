package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const bundleMagic = "STULTBUNDLE1"

const bundleZipSizeBytes = 8

type embeddedBundleFS struct {
	*zip.Reader
	file *os.File
}

func (bundle *embeddedBundleFS) Close() error {
	if bundle.file == nil {
		return nil
	}

	return bundle.file.Close()
}

func runEmbeddedBundleIfPresent() (bool, error) {
	files, found, err := openEmbeddedBundle()
	if err != nil {
		return false, err
	}

	if !found {
		return false, nil
	}

	if closer, ok := files.(interface{ Close() error }); ok {
		defer closer.Close()
	}

	return true, runEmbeddedBundle(files)
}

func runEmbeddedBundle(files fs.FS) error {
	manifestPath, err := findManifestInFS(files)
	if err != nil {
		return err
	}

	manifest, err := LoadManifestFromFS(files, manifestPath)
	if err != nil {
		return err
	}

	interpreter := NewInterpreter()

	return runManifestFromFS(interpreter, files, manifest.RunFiles)
}

func openEmbeddedBundle() (fs.FS, bool, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, false, fmt.Errorf("Could not locate executable: %w", err)
	}

	file, err := os.Open(executablePath)
	if err != nil {
		return nil, false, fmt.Errorf("Could not open executable %q: %w", executablePath, err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, false, fmt.Errorf("Could not inspect executable %q: %w", executablePath, err)
	}

	executableSize := info.Size()
	footerSize := int64(bundleFooterSize())

	if executableSize < footerSize {
		file.Close()
		return nil, false, nil
	}

	footer := make([]byte, footerSize)
	if _, err := file.ReadAt(footer, executableSize-footerSize); err != nil {
		file.Close()
		return nil, false, fmt.Errorf("Could not read executable bundle footer: %w", err)
	}

	zipSize, found, err := parseBundleFooter(footer)
	if err != nil {
		file.Close()
		return nil, false, err
	}

	if !found {
		file.Close()
		return nil, false, nil
	}

	zipOffset := executableSize - footerSize - int64(zipSize)
	if zipOffset < 0 {
		file.Close()
		return nil, false, fmt.Errorf("Invalid embedded bundle: archive offset is negative")
	}

	reader := io.NewSectionReader(file, zipOffset, int64(zipSize))

	zipReader, err := zip.NewReader(reader, int64(zipSize))
	if err != nil {
		file.Close()
		return nil, false, fmt.Errorf("Could not open embedded bundle: %w", err)
	}

	return &embeddedBundleFS{
		Reader: zipReader,
		file:   file,
	}, true, nil
}

func findManifestInFS(files fs.FS) (string, error) {
	hasStulton, err := fsFileExists(files, ManifestStultonFilename)
	if err != nil {
		return "", err
	}

	hasJSON, err := fsFileExists(files, ManifestJSONFilename)
	if err != nil {
		return "", err
	}

	if hasStulton && hasJSON {
		return "", fmt.Errorf(
			"Found both %q and %q in bundle; use only one manifest file",
			ManifestStultonFilename,
			ManifestJSONFilename,
		)
	}

	if hasStulton {
		return ManifestStultonFilename, nil
	}

	if hasJSON {
		return ManifestJSONFilename, nil
	}

	return "", fmt.Errorf(
		"Embedded bundle does not contain %s or %s",
		ManifestStultonFilename,
		ManifestJSONFilename,
	)
}

func fsFileExists(files fs.FS, filename string) (bool, error) {
	info, err := fs.Stat(files, filename)

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
			return "", "", fmt.Errorf(buildUsage())

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
		"  interpreter build [project-directory] -o <output-executable>\n" +
		"  interpreter build [project-directory]"
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

func embeddedBundleStart(bytes []byte) (int, bool, error) {
	footerSize := bundleFooterSize()

	if len(bytes) < footerSize {
		return 0, false, nil
	}

	footer := bytes[len(bytes)-footerSize:]

	zipSize, found, err := parseBundleFooter(footer)
	if err != nil {
		return 0, false, err
	}

	if !found {
		return 0, false, nil
	}

	if zipSize > uint64(len(bytes)-footerSize) {
		return 0, false, fmt.Errorf("Invalid existing embedded bundle: archive size is larger than executable")
	}

	return len(bytes) - footerSize - int(zipSize), true, nil
}

func createProjectBundleArchive(projectDir string) ([]byte, error) {
	var buffer bytes.Buffer

	zipWriter := zip.NewWriter(&buffer)

	err := filepath.WalkDir(projectDir, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(projectDir, filename)
		if err != nil {
			return err
		}

		bundlePath := filepath.ToSlash(relativePath)

		if !shouldIncludeBundleFile(bundlePath) {
			return nil
		}

		bytes, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("Could not read project file %q: %w", filename, err)
		}

		writer, err := zipWriter.Create(bundlePath)
		if err != nil {
			return fmt.Errorf("Could not add %q to bundle: %w", bundlePath, err)
		}

		if _, err := writer.Write(bytes); err != nil {
			return fmt.Errorf("Could not write %q to bundle: %w", bundlePath, err)
		}

		return nil
	})
	if err != nil {
		zipWriter.Close()
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("Could not finish bundle archive: %w", err)
	}

	return buffer.Bytes(), nil
}

func shouldIncludeBundleFile(bundlePath string) bool {
	base := path.Base(bundlePath)

	if bundlePath == ManifestStultonFilename || bundlePath == ManifestJSONFilename {
		return true
	}

	if base == ManifestStultonFilename || base == ManifestJSONFilename {
		return false
	}

	return path.Ext(bundlePath) == expectedFileExtension
}

func makeBundleFooter(zipSize uint64) []byte {
	footer := make([]byte, bundleFooterSize())

	binary.LittleEndian.PutUint64(footer[:bundleZipSizeBytes], zipSize)
	copy(footer[bundleZipSizeBytes:], []byte(bundleMagic))

	return footer
}

func parseBundleFooter(footer []byte) (uint64, bool, error) {
	if len(footer) != bundleFooterSize() {
		return 0, false, fmt.Errorf("Invalid bundle footer size")
	}

	magic := string(footer[bundleZipSizeBytes:])
	if magic != bundleMagic {
		return 0, false, nil
	}

	zipSize := binary.LittleEndian.Uint64(footer[:bundleZipSizeBytes])

	if zipSize == 0 {
		return 0, false, fmt.Errorf("Invalid embedded bundle: archive is empty")
	}

	return zipSize, true, nil
}

func bundleFooterSize() int {
	return bundleZipSizeBytes + len(bundleMagic)
}
