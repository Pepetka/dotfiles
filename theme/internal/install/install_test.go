package install

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFingerprintCompatibility(t *testing.T) {
	source := t.TempDir()
	directory := filepath.Join(source, "platform/macos")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{"main.swift": "alpha", "Info.plist": "beta"} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	hash, err := fingerprint(source)
	if err != nil {
		t.Fatal(err)
	}
	// SHA-256 of the exact byte stream used by the former installer, including NUL separators.
	if hash != "85d812cb27e6da86beb6e7a4e8a88309767205ea3d73984a0c330c2cf7fde358" {
		t.Fatalf("incompatible helper fingerprint: %s", hash)
	}
	if err := os.WriteFile(filepath.Join(directory, "Info.plist"), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	changed, err := fingerprint(source)
	if err != nil || changed == hash {
		t.Fatalf("metadata change did not invalidate fingerprint: %s, %v", changed, err)
	}
}

func TestServicePlistEscapesPaths(t *testing.T) {
	directory := "/config with spaces/кириллица & <tag> \"quotes\"/theme"
	data := servicePlist(directory)
	var parsed struct {
		Dict struct {
			Keys    []string `xml:"key"`
			Strings []string `xml:"string"`
			Array   struct {
				Strings []string `xml:"string"`
			} `xml:"array"`
			Integer int        `xml:"integer"`
			True    []struct{} `xml:"true"`
		} `xml:"dict"`
	}
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	d := parsed.Dict
	if len(d.Keys) != 8 || len(d.True) != 2 || d.Integer != 5 {
		t.Fatalf("incorrect launch policy: %+v", d)
	}
	want := []string{filepath.Join(directory, "theme-sync"), "watch", "--theme-dir", directory}
	if strings.Join(d.Array.Strings, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("incorrect arguments: %v", d.Array.Strings)
	}
	if d.Strings[0] != label || d.Strings[1] != "Aqua" || d.Strings[2] != "Background" || d.Strings[3] != filepath.Join(directory, "service.log") {
		t.Fatalf("incorrect service metadata: %v", d.Strings)
	}
}
