package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func put(t *testing.T, path, value string) {
	t.Helper()
	if err := atomicWrite(path, []byte(value)); err != nil {
		t.Fatal(err)
	}
}

func get(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func fixture(t *testing.T) options {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "repo", "configs", "codex")
	for _, name := range configFiles {
		put(t, filepath.Join(source, name), "model_reasoning_effort = 'medium'\n")
	}
	for _, name := range themeFiles {
		put(t, filepath.Join(root, "repo", "theme", "resources", "codex", name), "theme fixture")
	}
	return options{Source: source, Home: filepath.Join(root, "home")}
}

func install(t *testing.T, o options) {
	t.Helper()
	if err := synchronize(o); err != nil {
		t.Fatal(err)
	}
}

func TestDryRun(t *testing.T) {
	o := fixture(t)
	o.DryRun = true
	install(t, o)
	if _, err := os.Stat(o.Home); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("dry run created home")
	}
}

func TestRepeatAndConflict(t *testing.T) {
	o := fixture(t)
	install(t, o)
	target := filepath.Join(o.Home, "config.toml")
	before := get(t, target)
	install(t, o)
	if !bytes.Equal(before, get(t, target)) {
		t.Fatal("repeat changed config")
	}
	put(t, filepath.Join(o.Home, "review.config.toml"), "model_reasoning_effort = 'high'\n")
	put(t, filepath.Join(o.Source, "review.config.toml"), "model_reasoning_effort = 'xhigh'\n")
	put(t, filepath.Join(o.Source, "config.toml"), "model_reasoning_effort = 'low'\n")
	statePath := filepath.Join(o.Home, ".dotfiles-sync", "state.json")
	oldState := get(t, statePath)
	if err := synchronize(o); err == nil || !strings.Contains(err.Error(), "nothing written") {
		t.Fatalf("expected conflict: %v", err)
	}
	if !bytes.Equal(before, get(t, target)) || !bytes.Equal(oldState, get(t, statePath)) {
		t.Fatal("conflict modified files")
	}
}

func TestRollback(t *testing.T) {
	o := fixture(t)
	install(t, o)
	target := filepath.Join(o.Home, "config.toml")
	before := get(t, target)
	put(t, filepath.Join(o.Source, "config.toml"), "model_reasoning_effort = 'high'\n")
	o.Write = func(path string, data []byte) error {
		if filepath.Base(path) == "state.json" {
			return errors.New("injected failure")
		}
		return atomicWrite(path, data)
	}
	if err := synchronize(o); err == nil {
		t.Fatal("expected failure")
	}
	if !bytes.Equal(before, get(t, target)) {
		t.Fatal("rollback failed")
	}
	backups, _ := filepath.Glob(filepath.Join(o.Home, ".dotfiles-sync", "backups", "*", "config.toml"))
	if len(backups) != 1 {
		t.Fatal("backup missing")
	}
}

func TestThemeUpdateAndLocalConflict(t *testing.T) {
	o := fixture(t)
	install(t, o)
	source := filepath.Join(o.Source, "..", "..", "theme", "resources", "codex", themeFiles[0])
	target := filepath.Join(o.Home, "themes", themeFiles[0])
	put(t, source, "updated")
	install(t, o)
	if string(get(t, target)) != "updated" {
		t.Fatal("theme not updated")
	}
	put(t, target, "local")
	put(t, source, "another update")
	if err := synchronize(o); err == nil || !strings.Contains(err.Error(), "nothing written") {
		t.Fatalf("expected conflict: %v", err)
	}
	if string(get(t, target)) != "local" {
		t.Fatal("local theme overwritten")
	}
}

func TestConcurrentEdit(t *testing.T) {
	o := fixture(t)
	install(t, o)
	target := filepath.Join(o.Home, "config.toml")
	put(t, filepath.Join(o.Source, "config.toml"), "model_reasoning_effort = 'high'\n")
	o.BeforeCommit = func() { put(t, target, "model_reasoning_effort = 'local'\n") }
	if err := synchronize(o); err == nil || !strings.Contains(err.Error(), "changed during") {
		t.Fatalf("expected concurrent edit: %v", err)
	}
	if !strings.Contains(string(get(t, target)), "local") {
		t.Fatal("external edit overwritten")
	}
}

func TestPythonSnapshotCompatibility(t *testing.T) {
	o := fixture(t)
	install(t, o)
	path := filepath.Join(o.Home, ".dotfiles-sync", "state.json")
	var snapshot map[string]any
	if err := json.Unmarshal(get(t, path), &snapshot); err != nil {
		t.Fatal(err)
	}
	// Python stores numbers as JSON numbers too; missing theme hashes were valid in v1.
	delete(snapshot, "themes")
	data, _ := json.Marshal(snapshot)
	put(t, path, string(data))
	install(t, o)
	if !equal(int64(1), float64(1)) {
		t.Fatal("JSON/TOML numbers differ")
	}
}
