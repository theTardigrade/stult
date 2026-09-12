package main

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestStultonManifestRequiresUppercaseRun(t *testing.T) {
	files := fstest.MapFS{
		ManifestStultonFilename: {Data: []byte(`{
	"RUN": {
		"main.stult"
	}
}`)},
	}

	manifest, err := LoadManifestFromFS(files, ManifestStultonFilename)
	if err != nil {
		t.Fatalf("LoadManifestFromFS returned error: %v", err)
	}

	if len(manifest.RunFiles) != 1 || manifest.RunFiles[0] != "main.stult" {
		t.Fatalf("unexpected run files: %#v", manifest.RunFiles)
	}
}

func TestStultonManifestRejectsLowercaseRun(t *testing.T) {
	files := fstest.MapFS{
		ManifestStultonFilename: {Data: []byte(`{
	"run": {
		"main.stult"
	}
}`)},
	}

	_, err := LoadManifestFromFS(files, ManifestStultonFilename)
	if err == nil {
		t.Fatal("expected lowercase run field to be rejected")
	}

	if !strings.Contains(err.Error(), `manifest.stulton uses uppercase "RUN"; found "run"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJSONManifestRequiresLowercaseRun(t *testing.T) {
	files := fstest.MapFS{
		ManifestJSONFilename: {Data: []byte(`{
	"run": [
		"main.stult"
	]
}`)},
	}

	manifest, err := LoadManifestFromFS(files, ManifestJSONFilename)
	if err != nil {
		t.Fatalf("LoadManifestFromFS returned error: %v", err)
	}

	if len(manifest.RunFiles) != 1 || manifest.RunFiles[0] != "main.stult" {
		t.Fatalf("unexpected run files: %#v", manifest.RunFiles)
	}
}

func TestJSONManifestRejectsUppercaseRun(t *testing.T) {
	files := fstest.MapFS{
		ManifestJSONFilename: {Data: []byte(`{
	"RUN": [
		"main.stult"
	]
}`)},
	}

	_, err := LoadManifestFromFS(files, ManifestJSONFilename)
	if err == nil {
		t.Fatal("expected uppercase RUN field to be rejected")
	}

	if !strings.Contains(err.Error(), `manifest.json uses lowercase "run"; found "RUN"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStultonManifestParsesAssets(t *testing.T) {
	files := fstest.MapFS{
		ManifestStultonFilename: {Data: []byte(`{
	"RUN": "main.stult"
	"ASSETS": {
		"CONFIG": "./data/config.stulton"
		"TEMPLATES": "templates"
	}
}`)},
	}

	manifest, err := LoadManifestFromFS(files, ManifestStultonFilename)
	if err != nil {
		t.Fatalf("LoadManifestFromFS returned error: %v", err)
	}

	if manifest.Assets["CONFIG"] != "data/config.stulton" {
		t.Fatalf("unexpected CONFIG asset path: %#v", manifest.Assets["CONFIG"])
	}
	if manifest.Assets["TEMPLATES"] != "templates" {
		t.Fatalf("unexpected TEMPLATES asset path: %#v", manifest.Assets["TEMPLATES"])
	}
}

func TestStultonManifestRejectsLowercaseAssets(t *testing.T) {
	files := fstest.MapFS{
		ManifestStultonFilename: {Data: []byte(`{
	"RUN": "main.stult"
	"assets": {
		"CONFIG": "data/config.stulton"
	}
}`)},
	}

	_, err := LoadManifestFromFS(files, ManifestStultonFilename)
	if err == nil {
		t.Fatal("expected lowercase assets field to be rejected")
	}

	if !strings.Contains(err.Error(), `manifest.stulton uses uppercase "ASSETS"; found "assets"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJSONManifestParsesAssets(t *testing.T) {
	files := fstest.MapFS{
		ManifestJSONFilename: {Data: []byte(`{
	"run": "main.stult",
	"assets": {
		"CONFIG": "./data/config.stulton",
		"TEMPLATES": "templates"
	}
}`)},
	}

	manifest, err := LoadManifestFromFS(files, ManifestJSONFilename)
	if err != nil {
		t.Fatalf("LoadManifestFromFS returned error: %v", err)
	}

	if manifest.Assets["CONFIG"] != "data/config.stulton" {
		t.Fatalf("unexpected CONFIG asset path: %#v", manifest.Assets["CONFIG"])
	}
	if manifest.Assets["TEMPLATES"] != "templates" {
		t.Fatalf("unexpected TEMPLATES asset path: %#v", manifest.Assets["TEMPLATES"])
	}
}

func TestManifestRejectsInvalidAssetPath(t *testing.T) {
	files := fstest.MapFS{
		ManifestJSONFilename: {Data: []byte(`{
	"run": "main.stult",
	"assets": {
		"CONFIG": "../config.stulton"
	}
}`)},
	}

	_, err := LoadManifestFromFS(files, ManifestJSONFilename)
	if err == nil {
		t.Fatal("expected escaping asset path to be rejected")
	}

	if !strings.Contains(err.Error(), "path must not escape the manifest directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}
