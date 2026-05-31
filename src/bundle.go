package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
)

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
	if embeddedBundleWantsBytecode(files) {
		return runEmbeddedBytecodeBundle(files)
	}

	manifestPath, err := findManifestInFS(files)
	if err != nil {
		return err
	}

	manifest, err := LoadManifestFromFS(files, manifestPath)
	if err != nil {
		return err
	}

	interpreter := NewInterpreterWithArgs(os.Args[1:])

	return runManifestFromFS(interpreter, files, manifest.RunFiles)
}

func runEmbeddedBytecodeBundle(files fs.FS) error {
	manifestPath, err := findManifestInFS(files)
	if err != nil {
		return err
	}

	manifest, err := LoadManifestFromFS(files, manifestPath)
	if err != nil {
		return err
	}

	vm := NewBytecodeVM(os.Args[1:])

	return runBundledBytecodeManifestFromFS(vm, files, manifest.RunFiles)
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
