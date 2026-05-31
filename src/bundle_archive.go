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

func createProjectBundleArchiveWithOptions(
	projectDir string,
	options BuildBundleOptions,
) ([]byte, error) {
	var buffer bytes.Buffer

	zipWriter := zip.NewWriter(&buffer)

	bytecodeRunFiles := map[string]string{}
	bytecodePathsByBundlePath := map[string]string{}
	bytecodePathsByAbsolutePath := map[string]string{}

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

		bundlePath := cleanFSPath(filepath.ToSlash(relativePath))

		if !shouldIncludeBundleFile(bundlePath) {
			return nil
		}

		bytes, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("Could not read project file %q: %w", filename, err)
		}

		if options.RunBytecode && path.Ext(bundlePath) == expectedFileExtension {
			bytecodePath := bytecodeBundlePathForSource(bundlePath)

			if err := writeCompiledBytecodeToBundle(
				zipWriter,
				bytes,
				bundlePath,
				bytecodePath,
			); err != nil {
				return err
			}

			absoluteSourcePath, err := filepath.Abs(filename)
			if err != nil {
				return fmt.Errorf("Could not resolve source path %q: %w", filename, err)
			}

			bytecodePathsByBundlePath[bundlePath] = bytecodePath
			bytecodePathsByAbsolutePath[filepath.Clean(absoluteSourcePath)] = bytecodePath

			return nil
		}

		if err := writeBundleArchiveFile(zipWriter, bundlePath, bytes); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		zipWriter.Close()
		return nil, err
	}

	if options.RunBytecode {
		if err := addProjectBundleBytecodeRunMap(
			zipWriter,
			projectDir,
			bytecodeRunFiles,
			bytecodePathsByBundlePath,
			bytecodePathsByAbsolutePath,
		); err != nil {
			zipWriter.Close()
			return nil, err
		}

		runMapBytes, err := encodeBundledBytecodeRunMap(bytecodeRunFiles)
		if err != nil {
			zipWriter.Close()
			return nil, err
		}

		if err := writeBundleArchiveFile(
			zipWriter,
			bytecodeBundleRunMapFilename,
			runMapBytes,
		); err != nil {
			zipWriter.Close()
			return nil, err
		}

		if err := writeBundleArchiveFile(
			zipWriter,
			bytecodeBundleModeFilename,
			[]byte(bytecodeBundleModeFileContents),
		); err != nil {
			zipWriter.Close()
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("Could not finish bundle archive: %w", err)
	}

	return buffer.Bytes(), nil
}

func addProjectBundleBytecodeRunMap(
	zipWriter *zip.Writer,
	projectDir string,
	bytecodeRunFiles map[string]string,
	bytecodePathsByBundlePath map[string]string,
	bytecodePathsByAbsolutePath map[string]string,
) error {
	manifestPath, found, err := findManifestInDirectory(projectDir)
	if err != nil {
		return err
	}

	if !found {
		return fmt.Errorf(
			"Project directory %q must contain %s or %s",
			projectDir,
			ManifestStultonFilename,
			ManifestJSONFilename,
		)
	}

	manifest, _, err := loadManifestFileFromFS(manifestPath)
	if err != nil {
		return err
	}

	for _, runFile := range manifest.RunFiles {
		key := bytecodeBundleRunFileKey(runFile)

		if _, exists := bytecodeRunFiles[key]; exists {
			continue
		}

		if filepath.IsAbs(runFile) {
			absoluteRunFile := filepath.Clean(runFile)

			if bytecodePath, ok := bytecodePathsByAbsolutePath[absoluteRunFile]; ok {
				bytecodeRunFiles[key] = bytecodePath
				continue
			}

			bytecodePath, err := addAbsoluteRunFileBytecodeToBundle(
				zipWriter,
				absoluteRunFile,
			)
			if err != nil {
				return err
			}

			bytecodePathsByAbsolutePath[absoluteRunFile] = bytecodePath
			bytecodeRunFiles[key] = bytecodePath

			continue
		}

		bundlePath := cleanFSPath(runFile)

		if bytecodePath, ok := bytecodePathsByBundlePath[bundlePath]; ok {
			bytecodeRunFiles[key] = bytecodePath
			continue
		}

		absoluteRunFile := filepath.Join(projectDir, filepath.FromSlash(bundlePath))
		bytecodePath := bytecodeBundlePathForSource(bundlePath)

		sourceBytes, err := os.ReadFile(absoluteRunFile)
		if err != nil {
			return fmt.Errorf("Could not read manifest run file %q: %w", runFile, err)
		}

		if err := writeCompiledBytecodeToBundle(
			zipWriter,
			sourceBytes,
			runFile,
			bytecodePath,
		); err != nil {
			return err
		}

		absoluteRunFile, err = filepath.Abs(absoluteRunFile)
		if err != nil {
			return fmt.Errorf("Could not resolve manifest run file %q: %w", runFile, err)
		}

		bytecodePathsByBundlePath[bundlePath] = bytecodePath
		bytecodePathsByAbsolutePath[filepath.Clean(absoluteRunFile)] = bytecodePath
		bytecodeRunFiles[key] = bytecodePath
	}

	return nil
}

func addAbsoluteRunFileBytecodeToBundle(
	zipWriter *zip.Writer,
	filename string,
) (string, error) {
	sourceBytes, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("Could not read absolute manifest run file %q: %w", filename, err)
	}

	bytecodePath := bytecodeBundlePathForAbsoluteSource(filename)

	if err := writeCompiledBytecodeToBundle(
		zipWriter,
		sourceBytes,
		filename,
		bytecodePath,
	); err != nil {
		return "", err
	}

	return bytecodePath, nil
}

func writeCompiledBytecodeToBundle(
	zipWriter *zip.Writer,
	sourceBytes []byte,
	displayName string,
	bytecodePath string,
) error {
	bytecodeBytes, err := compileSourceBytesToBundledBytecode(sourceBytes, displayName)
	if err != nil {
		return err
	}

	return writeBundleArchiveFile(zipWriter, bytecodePath, bytecodeBytes)
}

func createSingleSourceBundleArchive(
	filename string,
	options BuildBundleOptions,
) ([]byte, error) {
	var buffer bytes.Buffer

	zipWriter := zip.NewWriter(&buffer)

	sourceBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not read source file %q: %w", filename, err)
	}

	sourcePath := filepath.ToSlash(filepath.Base(filename))
	manifestBytes := []byte(fmt.Sprintf("{\n\t\"run\": [%q]\n}\n", sourcePath))

	if err := writeBundleArchiveFile(zipWriter, ManifestJSONFilename, manifestBytes); err != nil {
		zipWriter.Close()
		return nil, err
	}

	if options.RunBytecode {
		bytecodePath := bytecodeBundlePathForSource(sourcePath)

		if err := writeCompiledBytecodeToBundle(
			zipWriter,
			sourceBytes,
			sourcePath,
			bytecodePath,
		); err != nil {
			zipWriter.Close()
			return nil, err
		}

		runMapBytes, err := encodeBundledBytecodeRunMap(map[string]string{
			bytecodeBundleRunFileKey(sourcePath): bytecodePath,
		})
		if err != nil {
			zipWriter.Close()
			return nil, err
		}

		if err := writeBundleArchiveFile(
			zipWriter,
			bytecodeBundleRunMapFilename,
			runMapBytes,
		); err != nil {
			zipWriter.Close()
			return nil, err
		}

		if err := writeBundleArchiveFile(
			zipWriter,
			bytecodeBundleModeFilename,
			[]byte(bytecodeBundleModeFileContents),
		); err != nil {
			zipWriter.Close()
			return nil, err
		}

		if err := zipWriter.Close(); err != nil {
			return nil, fmt.Errorf("Could not finish bundle archive: %w", err)
		}

		return buffer.Bytes(), nil
	}

	if err := writeBundleArchiveFile(zipWriter, sourcePath, sourceBytes); err != nil {
		zipWriter.Close()
		return nil, err
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("Could not finish bundle archive: %w", err)
	}

	return buffer.Bytes(), nil
}

func writeBundleArchiveFile(zipWriter *zip.Writer, bundlePath string, bytes []byte) error {
	writer, err := zipWriter.Create(bundlePath)
	if err != nil {
		return fmt.Errorf("Could not add %q to bundle: %w", bundlePath, err)
	}

	if _, err := writer.Write(bytes); err != nil {
		return fmt.Errorf("Could not write %q to bundle: %w", bundlePath, err)
	}

	return nil
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
