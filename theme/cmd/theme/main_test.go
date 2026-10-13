package main

import (
	"dotfiles/theme/internal/engine"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func install(t *testing.T, c engine.Config) string {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	c.Root, c.Directory = root, filepath.Join(t.TempDir(), "config with spaces", "theme")
	if err := os.MkdirAll(c.Directory, 0700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(c)
	if err := os.WriteFile(filepath.Join(c.Directory, "engine.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	return c.Directory
}

func TestManualInstallationRejectsAutomaticModes(t *testing.T) {
	directory := install(t, engine.Config{})
	if err := execute([]string{"light", "--theme-dir", directory}); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"auto", "watch"} {
		if err := execute([]string{command, "--theme-dir", directory}); !errors.Is(err, engine.ErrNoBackend) {
			t.Fatalf("%s accepted without a backend: %v", command, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(directory, "preference"))
	if err != nil || string(data) != "light\n" {
		t.Fatalf("rejected command changed the preference: %q, %v", data, err)
	}
	if backend, service := describe(engine.Config{}); backend != "manual" || service != "none" {
		t.Fatalf("manual status: %s, %s", backend, service)
	}
}

func TestLegacyConfigurationUsesLaunchAgent(t *testing.T) {
	c := engine.Config{Backend: "/backend"}
	want := ""
	if runtime.GOOS == "darwin" {
		want = "launchd"
	}
	if c.ServiceManager() != want {
		t.Fatalf("legacy service: %q", c.ServiceManager())
	}
}
