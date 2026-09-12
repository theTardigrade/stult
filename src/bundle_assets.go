package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const bundleAssetsIndexFilename = ".stult-assets/index.json"
const bundleAssetFileDir = ".stult-assets/files"
const bundleAssetDirectoryDir = ".stult-assets/directories"

type BundleAssetKind string

const (
	BundleAssetKindFile      BundleAssetKind = "file"
	BundleAssetKindDirectory BundleAssetKind = "directory"
)

type BundleAsset struct {
	Kind   BundleAssetKind `json:"kind"`
	Source string          `json:"source,omitempty"`
	Path   string          `json:"path,omitempty"`
	Prefix string          `json:"prefix,omitempty"`
}

type bundledAssetIndex struct {
	Assets map[string]BundleAsset `json:"assets"`
}

type BundleAssetStoreMode int

const (
	BundleAssetStoreModeProject BundleAssetStoreMode = iota
	BundleAssetStoreModeEmbedded
)

type BundleAssetStore struct {
	mode   BundleAssetStoreMode
	files  fs.FS
	assets map[string]BundleAsset
}

func NewEmptyBundleAssetStore() *BundleAssetStore {
	return &BundleAssetStore{
		mode:   BundleAssetStoreModeProject,
		assets: map[string]BundleAsset{},
	}
}

func NewProjectBundleAssetStore(files fs.FS, assets map[string]string) *BundleAssetStore {
	storeAssets := map[string]BundleAsset{}

	for name, source := range assets {
		storeAssets[name] = BundleAsset{
			Source: source,
		}
	}

	return &BundleAssetStore{
		mode:   BundleAssetStoreModeProject,
		files:  files,
		assets: storeAssets,
	}
}

func NewEmbeddedBundleAssetStore(files fs.FS) (*BundleAssetStore, error) {
	index, err := readBundledAssetIndex(files)
	if err != nil {
		return nil, err
	}

	return &BundleAssetStore{
		mode:   BundleAssetStoreModeEmbedded,
		files:  files,
		assets: index.Assets,
	}, nil
}

func (store *BundleAssetStore) Keys() []string {
	if store == nil || len(store.assets) == 0 {
		return []string{}
	}

	names := make([]string, 0, len(store.assets))
	for name := range store.assets {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}

func (store *BundleAssetStore) Read(name string, relativePath string, hasRelativePath bool) ([]byte, error) {
	asset, ok, err := store.lookupAsset(name)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unknown bundle asset %q", name)
	}

	kind, err := store.assetKind(asset)
	if err != nil {
		return nil, fmt.Errorf("bundle asset %q: %w", name, err)
	}

	switch kind {
	case BundleAssetKindFile:
		if hasRelativePath {
			return nil, fmt.Errorf("bundle asset %q is a file asset; path argument must be omitted or _", name)
		}

		return fs.ReadFile(store.files, store.assetFilePath(asset))

	case BundleAssetKindDirectory:
		if !hasRelativePath {
			return nil, fmt.Errorf("bundle asset %q is a directory asset; path argument is required", name)
		}

		cleanRelativePath, err := cleanBundleAssetInnerPath(relativePath, false)
		if err != nil {
			return nil, fmt.Errorf("bundle asset %q path: %w", name, err)
		}

		return fs.ReadFile(store.files, store.assetDirectoryChildPath(asset, cleanRelativePath))

	default:
		return nil, fmt.Errorf("bundle asset %q has unknown kind %q", name, kind)
	}
}

func (store *BundleAssetStore) Exists(name string, relativePath string, hasRelativePath bool) (bool, error) {
	asset, ok, err := store.lookupAsset(name)
	if err != nil || !ok {
		return false, err
	}

	kind, err := store.assetKind(asset)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("bundle asset %q: %w", name, err)
	}

	var assetPath string

	switch kind {
	case BundleAssetKindFile:
		if hasRelativePath {
			return false, fmt.Errorf("bundle asset %q is a file asset; path argument must be omitted or _", name)
		}
		assetPath = store.assetFilePath(asset)

	case BundleAssetKindDirectory:
		if !hasRelativePath {
			assetPath = store.assetDirectoryPath(asset)
			break
		}

		cleanRelativePath, err := cleanBundleAssetInnerPath(relativePath, false)
		if err != nil {
			return false, fmt.Errorf("bundle asset %q path: %w", name, err)
		}
		assetPath = store.assetDirectoryChildPath(asset, cleanRelativePath)

	default:
		return false, fmt.Errorf("bundle asset %q has unknown kind %q", name, kind)
	}

	_, err = fs.Stat(store.files, assetPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func (store *BundleAssetStore) List(name string, relativePath string, hasRelativePath bool) ([]string, error) {
	asset, ok, err := store.lookupAsset(name)
	if err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unknown bundle asset %q", name)
	}

	kind, err := store.assetKind(asset)
	if err != nil {
		return nil, fmt.Errorf("bundle asset %q: %w", name, err)
	}
	if kind != BundleAssetKindDirectory {
		return nil, fmt.Errorf("bundle asset %q is a file asset; LIST requires a directory asset", name)
	}

	root := store.assetDirectoryPath(asset)
	listRoot := root
	prefix := ""
	if hasRelativePath {
		cleanRelativePath, err := cleanBundleAssetInnerPath(relativePath, true)
		if err != nil {
			return nil, fmt.Errorf("bundle asset %q path: %w", name, err)
		}
		if cleanRelativePath != "" {
			listRoot = store.assetDirectoryChildPath(asset, cleanRelativePath)
			prefix = cleanRelativePath
		}
	}

	entries := []string{}
	err = fs.WalkDir(store.files, listRoot, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		relative := strings.TrimPrefix(filename, root)
		relative = strings.TrimPrefix(relative, "/")
		if relative == filename || relative == "" || strings.HasPrefix(relative, "../") || relative == ".." {
			return nil
		}
		if prefix != "" && relative != prefix && !strings.HasPrefix(relative, prefix+"/") {
			return nil
		}

		entries = append(entries, relative)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(entries)
	return entries, nil
}

func (store *BundleAssetStore) lookupAsset(name string) (BundleAsset, bool, error) {
	if store == nil || len(store.assets) == 0 {
		return BundleAsset{}, false, nil
	}

	asset, ok := store.assets[name]
	return asset, ok, nil
}

func (store *BundleAssetStore) assetKind(asset BundleAsset) (BundleAssetKind, error) {
	if store.mode == BundleAssetStoreModeEmbedded {
		if asset.Kind == "" {
			return "", fmt.Errorf("missing kind")
		}
		return asset.Kind, nil
	}

	info, err := fs.Stat(store.files, asset.Source)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return BundleAssetKindDirectory, nil
	}
	if info.Mode().IsRegular() {
		return BundleAssetKindFile, nil
	}

	return "", fmt.Errorf("source is not a regular file or directory")
}

func (store *BundleAssetStore) assetFilePath(asset BundleAsset) string {
	if store.mode == BundleAssetStoreModeEmbedded {
		return asset.Path
	}

	return asset.Source
}

func (store *BundleAssetStore) assetDirectoryPath(asset BundleAsset) string {
	if store.mode == BundleAssetStoreModeEmbedded {
		return strings.TrimSuffix(asset.Prefix, "/")
	}

	return asset.Source
}

func (store *BundleAssetStore) assetDirectoryChildPath(asset BundleAsset, relativePath string) string {
	if store.mode == BundleAssetStoreModeEmbedded {
		return path.Join(asset.Prefix, relativePath)
	}

	return path.Join(asset.Source, relativePath)
}

func addProjectBundleAssets(
	zipWriter *zip.Writer,
	projectDir string,
	manifest *Manifest,
) error {
	if manifest == nil || len(manifest.Assets) == 0 {
		return nil
	}

	index := bundledAssetIndex{
		Assets: map[string]BundleAsset{},
	}

	names := make([]string, 0, len(manifest.Assets))
	for name := range manifest.Assets {
		names = append(names, name)
	}
	sort.Strings(names)

	fileIndex := 0
	directoryIndex := 0

	for _, name := range names {
		source := manifest.Assets[name]
		filename := filepath.Join(projectDir, filepath.FromSlash(source))

		info, err := os.Stat(filename)
		if err != nil {
			return fmt.Errorf("Could not inspect bundle asset %q at %q: %w", name, source, err)
		}

		if info.Mode().IsRegular() {
			bundlePath := path.Join(bundleAssetFileDir, fmt.Sprintf("%d", fileIndex))
			fileIndex++

			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("Could not read bundle asset %q at %q: %w", name, source, err)
			}
			if err := writeBundleArchiveFile(zipWriter, bundlePath, data); err != nil {
				return err
			}

			index.Assets[name] = BundleAsset{
				Kind:   BundleAssetKindFile,
				Source: source,
				Path:   bundlePath,
			}
			continue
		}

		if info.IsDir() {
			prefix := path.Join(bundleAssetDirectoryDir, fmt.Sprintf("%d", directoryIndex)) + "/"
			directoryIndex++

			if err := addProjectBundleAssetDirectory(zipWriter, filename, prefix); err != nil {
				return fmt.Errorf("Could not add bundle asset directory %q at %q: %w", name, source, err)
			}

			index.Assets[name] = BundleAsset{
				Kind:   BundleAssetKindDirectory,
				Source: source,
				Prefix: prefix,
			}
			continue
		}

		return fmt.Errorf("Bundle asset %q at %q is not a regular file or directory", name, source)
	}

	indexBytes, err := encodeBundledAssetIndex(index)
	if err != nil {
		return err
	}

	return writeBundleArchiveFile(zipWriter, bundleAssetsIndexFilename, indexBytes)
}

func addProjectBundleAssetDirectory(zipWriter *zip.Writer, directory string, bundlePrefix string) error {
	return filepath.WalkDir(directory, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported non-regular asset file %q", filename)
		}

		relativePath, err := filepath.Rel(directory, filename)
		if err != nil {
			return err
		}
		bundlePath := path.Join(bundlePrefix, cleanFSPath(filepath.ToSlash(relativePath)))

		data, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		return writeBundleArchiveFile(zipWriter, bundlePath, data)
	})
}

func encodeBundledAssetIndex(index bundledAssetIndex) ([]byte, error) {
	if index.Assets == nil {
		index.Assets = map[string]BundleAsset{}
	}

	return json.MarshalIndent(index, "", "\t")
}

func readBundledAssetIndex(files fs.FS) (bundledAssetIndex, error) {
	bytes, err := fs.ReadFile(files, bundleAssetsIndexFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return bundledAssetIndex{Assets: map[string]BundleAsset{}}, nil
		}

		return bundledAssetIndex{}, fmt.Errorf("Could not read bundled asset index: %w", err)
	}

	var index bundledAssetIndex
	if err := json.Unmarshal(bytes, &index); err != nil {
		return bundledAssetIndex{}, fmt.Errorf("Could not decode bundled asset index: %w", err)
	}
	if index.Assets == nil {
		index.Assets = map[string]BundleAsset{}
	}

	for name, asset := range index.Assets {
		if err := validateManifestAssetName(name); err != nil {
			return bundledAssetIndex{}, fmt.Errorf("invalid bundled asset name %q: %w", name, err)
		}
		switch asset.Kind {
		case BundleAssetKindFile:
			if asset.Path == "" {
				return bundledAssetIndex{}, fmt.Errorf("bundle asset %q is missing file path", name)
			}
		case BundleAssetKindDirectory:
			if asset.Prefix == "" {
				return bundledAssetIndex{}, fmt.Errorf("bundle asset %q is missing directory prefix", name)
			}
		default:
			return bundledAssetIndex{}, fmt.Errorf("bundle asset %q has unknown kind %q", name, asset.Kind)
		}
	}

	return index, nil
}

func cleanManifestAssetSourcePath(source string) (string, error) {
	if strings.TrimSpace(source) == "" {
		return "", fmt.Errorf("path must not be empty")
	}
	if source != strings.TrimSpace(source) {
		return "", fmt.Errorf("path must not have leading or trailing whitespace")
	}
	if filepath.IsAbs(source) || path.IsAbs(source) || looksLikeWindowsAbsolutePath(source) {
		return "", fmt.Errorf("path must be relative")
	}

	normalized := strings.ReplaceAll(source, "\\", "/")
	cleaned := path.Clean(normalized)
	cleaned = strings.TrimPrefix(cleaned, "./")

	if cleaned == "." || cleaned == "" {
		return "", fmt.Errorf("path must identify a file or directory")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("path must not escape the manifest directory")
	}

	return cleaned, nil
}

func validateManifestAssetName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name must not be empty")
	}
	if name != strings.TrimSpace(name) {
		return fmt.Errorf("name must not have leading or trailing whitespace")
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("name must not contain path separators")
	}
	if name == "." || name == ".." || strings.ContainsRune(name, '\x00') {
		return fmt.Errorf("name is invalid")
	}

	return nil
}

func cleanBundleAssetInnerPath(relativePath string, allowEmpty bool) (string, error) {
	if relativePath != strings.TrimSpace(relativePath) {
		return "", fmt.Errorf("path must not have leading or trailing whitespace")
	}
	if strings.TrimSpace(relativePath) == "" {
		if allowEmpty {
			return "", nil
		}
		return "", fmt.Errorf("path must not be empty")
	}
	if filepath.IsAbs(relativePath) || path.IsAbs(relativePath) || looksLikeWindowsAbsolutePath(relativePath) {
		return "", fmt.Errorf("path must be relative")
	}

	normalized := strings.ReplaceAll(relativePath, "\\", "/")
	cleaned := path.Clean(normalized)
	cleaned = strings.TrimPrefix(cleaned, "./")

	if cleaned == "." || cleaned == "" {
		if allowEmpty {
			return "", nil
		}
		return "", fmt.Errorf("path must identify a file")
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("path must not escape the asset directory")
	}

	return cleaned, nil
}

func looksLikeWindowsAbsolutePath(filename string) bool {
	if len(filename) < 3 {
		return false
	}
	first := filename[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return false
	}
	if filename[1] != ':' {
		return false
	}
	return filename[2] == '/' || filename[2] == '\\'
}
