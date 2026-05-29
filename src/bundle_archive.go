package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

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
