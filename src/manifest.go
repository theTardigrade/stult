package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	ManifestStultonFilename = "manifest.stulton"
	ManifestJSONFilename    = "manifest.json"

	DefaultManifestFilename = ManifestStultonFilename
)

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

	file, err := parseManifestFile(absolutePath, bytes)
	if err != nil {
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

func parseManifestFile(filename string, bytes []byte) (manifestFile, error) {
	switch filepath.Base(filename) {
	case ManifestStultonFilename:
		return parseStultonManifest(bytes)

	case ManifestJSONFilename:
		return parseJSONManifest(bytes)

	default:
		return manifestFile{}, fmt.Errorf(
			"unsupported manifest filename %q; expected %s or %s",
			filepath.Base(filename),
			ManifestStultonFilename,
			ManifestJSONFilename,
		)
	}
}

func parseJSONManifest(bytes []byte) (manifestFile, error) {
	var file manifestFile

	if err := json.Unmarshal(bytes, &file); err != nil {
		return manifestFile{}, err
	}

	return file, nil
}

func parseStultonManifest(bytes []byte) (manifestFile, error) {
	value, err := stdDataStultonParseText(string(bytes))
	if err != nil {
		return manifestFile{}, err
	}

	return manifestFileFromStultonValue(value)
}

func manifestFileFromStultonValue(value Value) (manifestFile, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueMap {
		return manifestFile{}, fmt.Errorf("manifest root must be a map")
	}

	if value.Map == nil {
		return manifestFile{}, fmt.Errorf("manifest root map is invalid")
	}

	upperRunBinding, hasUpperRun := value.Map.Entries["RUN"]
	lowerRunBinding, hasLowerRun := value.Map.Entries["run"]

	if hasUpperRun && hasLowerRun {
		return manifestFile{}, fmt.Errorf(`manifest.stulton cannot contain both "RUN" and "run"`)
	}

	var runBinding Binding
	hasRun := false

	if hasUpperRun {
		runBinding = upperRunBinding
		hasRun = true
	} else if hasLowerRun {
		runBinding = lowerRunBinding
		hasRun = true
	}

	if !hasRun {
		return manifestFile{}, nil
	}

	run, err := manifestRunListFromValue(runBinding.Value)
	if err != nil {
		fieldName := "RUN"
		if hasLowerRun {
			fieldName = "run"
		}

		return manifestFile{}, fmt.Errorf("invalid manifest field %q: %w", fieldName, err)
	}

	return manifestFile{
		Run: run,
	}, nil
}

func manifestRunListFromValue(value Value) (manifestRunList, error) {
	value = resolveSpecializedValue(value)

	switch value.Kind {
	case ValueString:
		text, err := manifestStringFromValue(value)
		if err != nil {
			return nil, err
		}

		return manifestRunList{text}, nil

	case ValueArray:
		if value.Array == nil {
			return nil, fmt.Errorf("run array is invalid")
		}

		runFiles := make(manifestRunList, 0, len(value.Array.Elements))

		for index, element := range value.Array.Elements {
			text, err := manifestStringFromValue(element)
			if err != nil {
				return nil, fmt.Errorf("run array item at index %d must be a string", index)
			}

			runFiles = append(runFiles, text)
		}

		return runFiles, nil

	default:
		return nil, fmt.Errorf("run must be a string or an array of strings")
	}
}

func manifestStringFromValue(value Value) (string, error) {
	value = resolveSpecializedValue(value)

	if value.Kind != ValueString {
		return "", fmt.Errorf("value must be a string")
	}

	if value.Text == nil {
		return "", fmt.Errorf("string value is invalid")
	}

	return value.Text.String(), nil
}

func (manifest *Manifest) validate() error {
	if len(manifest.Run) == 0 {
		return fmt.Errorf("Manifest %q must contain a non-empty \"run\" or \"RUN\" field", manifest.Path)
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
