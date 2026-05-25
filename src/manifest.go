package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultManifestFilename = "stult.json"

type Manifest struct {
	Path string
	Dir  string

	Run      []string
	RunFiles []string
}

type manifestFile struct {
	Run manifestRunList `json:"run"`
}

type manifestRunList []string

func LoadDefaultManifest() (*Manifest, error) {
	return LoadManifest(DefaultManifestFilename)
}

func LoadManifest(filename string) (*Manifest, error) {
	if strings.TrimSpace(filename) == "" {
		filename = DefaultManifestFilename
	}

	absolutePath, err := filepath.Abs(filename)
	if err != nil {
		return nil, fmt.Errorf("Could not resolve manifest path %q: %w", filename, err)
	}

	bytes, err := os.ReadFile(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("Could not read manifest %q: %w", filename, err)
	}

	var file manifestFile
	if err := json.Unmarshal(bytes, &file); err != nil {
		return nil, fmt.Errorf("Could not parse manifest %q: %w", filename, err)
	}

	dir := filepath.Dir(absolutePath)

	manifest := &Manifest{
		Path: absolutePath,
		Dir:  dir,
		Run:  []string(file.Run),
	}

	if err := manifest.validate(); err != nil {
		return nil, err
	}

	manifest.RunFiles = resolveManifestRunFiles(manifest.Dir, manifest.Run)

	return manifest, nil
}

func (manifest *Manifest) validate() error {
	if len(manifest.Run) == 0 {
		return fmt.Errorf("Manifest %q must contain a non-empty \"run\" field", manifest.Path)
	}

	for index, runFile := range manifest.Run {
		if strings.TrimSpace(runFile) == "" {
			return fmt.Errorf("Manifest %q has an empty run file at index %d", manifest.Path, index)
		}
	}

	return nil
}

func resolveManifestRunFiles(baseDir string, runFiles []string) []string {
	resolved := make([]string, 0, len(runFiles))

	for _, runFile := range runFiles {
		if filepath.IsAbs(runFile) {
			resolved = append(resolved, filepath.Clean(runFile))
			continue
		}

		resolved = append(resolved, filepath.Clean(filepath.Join(baseDir, runFile)))
	}

	return resolved
}

func (runList *manifestRunList) UnmarshalJSON(bytes []byte) error {
	var single string
	if err := json.Unmarshal(bytes, &single); err == nil {
		*runList = manifestRunList{single}
		return nil
	}

	var many []string
	if err := json.Unmarshal(bytes, &many); err == nil {
		*runList = manifestRunList(many)
		return nil
	}

	return fmt.Errorf("\"run\" must be a string or an array of strings")
}
