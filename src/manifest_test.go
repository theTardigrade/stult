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
