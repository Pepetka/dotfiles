package install

import (
	"os"
	"path/filepath"
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
