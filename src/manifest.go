package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const DefaultManifestFilename = "stult.json"

type Manifest struct {
	Path string
	Dir  string

	Alias      map[string]string
	AliasNames []string

	Run      []string
	RunFiles []string
}

type manifestFile struct {
	Alias map[string]string `json:"alias"`
	Run   manifestRunList   `json:"run"`
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
		Path:  absolutePath,
		Dir:   dir,
		Alias: normalizeManifestAliases(file.Alias),
		Run:   []string(file.Run),
	}

	manifest.AliasNames = sortedManifestAliasNames(manifest.Alias)

	if err := manifest.validate(); err != nil {
		return nil, err
	}

	manifest.RunFiles = resolveManifestRunFiles(manifest.Dir, manifest.Run)

	return manifest, nil
}

func normalizeManifestAliases(aliases map[string]string) map[string]string {
	if aliases == nil {
		return map[string]string{}
	}

	normalized := make(map[string]string, len(aliases))

	for name, expression := range aliases {
		normalized[name] = expression
	}

	return normalized
}

func sortedManifestAliasNames(aliases map[string]string) []string {
	names := make([]string, 0, len(aliases))

	for name := range aliases {
		names = append(names, name)
	}

	sort.Strings(names)

	return names
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

	for _, name := range manifest.AliasNames {
		expression := manifest.Alias[name]

		if !isValidManifestAliasName(name) {
			return fmt.Errorf("Manifest %q has invalid alias name %q", manifest.Path, name)
		}

		if strings.TrimSpace(expression) == "" {
			return fmt.Errorf("Manifest %q has an empty expression for alias %q", manifest.Path, name)
		}
	}

	return nil
}

func isValidManifestAliasName(name string) bool {
	if name == "" || name == "_" {
		return false
	}

	runes := []rune(name)

	if !isManifestIdentifierStart(runes[0]) {
		return false
	}

	for _, ch := range runes[1:] {
		if !isManifestIdentifierPart(ch) {
			return false
		}
	}

	return true
}

func isManifestIdentifierStart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isManifestIdentifierPart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
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
