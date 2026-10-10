package integration

import (
	"context"
	"dotfiles/theme/internal/engine"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Opt-in because this test launches installed terminal executables.
func TestApplications(t *testing.T) {
	if os.Getenv("THEME_APPLICATION_TESTS") != "1" {
		t.Skip("run theme/tests/applications.sh for native terminal checks")
	}
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(source)
	temporary := t.TempDir()
	configRoot := filepath.Join(temporary, "config with spaces")
	directory := filepath.Join(configRoot, "theme")
	write := func(path string, data []byte, mode os.FileMode) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, mode); err != nil {
			t.Fatal(err)
		}
	}
	for app, name := range map[string]string{"ghostty": "config", "alacritty": "alacritty.toml"} {
		target := filepath.Join(configRoot, app, name)
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		resource := filepath.Join(root, "configs", name)
		if app == "ghostty" {
			resource = filepath.Join(root, "configs/ghostty/config")
		}
		if err := os.Symlink(resource, target); err != nil {
			t.Fatal(err)
		}
	}
	bridge, err := os.ReadFile(filepath.Join(source, "wezterm.lua"))
	if err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(directory, "wezterm-bridge.lua"), bridge, 0600)
	backend := filepath.Join(temporary, "backend")
	write(backend, []byte("#!/bin/sh\necho light\n"), 0700)
	settings := filepath.Join(directory, "engine.json")
	data, err := json.Marshal(engine.Config{Root: root, Directory: directory, Backend: backend})
	if err != nil {
		t.Fatal(err)
	}
	write(settings, data, 0600)
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	run := func(name string, args ...string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, name, args...)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}
	requireRun := func(name string, args ...string) string {
		t.Helper()
		output, err := run(name, args...)
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, output)
		}
		return output
	}
	binary := filepath.Join(temporary, "engine")
	requireRun("go", "build", "-o", binary, "../../cmd/theme")
	for _, variant := range []struct{ mode, style, background string }{{"dark", "moon", "#222436"}, {"light", "day", "#e1e2e7"}} {
		requireRun(binary, variant.mode, "--config", settings)
		if ghostty, err := exec.LookPath("ghostty"); err == nil {
			content := requireRun(ghostty, "+show-config")
			label := map[string]string{"moon": "Moon", "day": "Day"}[variant.style]
			for _, expected := range []string{"background = " + variant.background + "\n", "theme = TokyoNight " + label + "\n"} {
				if !strings.Contains(content, expected) {
					t.Fatalf("Ghostty missing %q", expected)
				}
			}
			requireRun(ghostty, "+validate-config", "--config-file="+filepath.Join(configRoot, "ghostty/config"))
		}
		if wezterm, err := exec.LookPath("wezterm"); err == nil {
			check := filepath.Join(temporary, "wezterm-check.lua")
			quoted, _ := json.Marshal(filepath.Join(root, "configs/wezterm.lua"))
			write(check, []byte(fmt.Sprintf("local config = dofile(%s)\nassert(config.color_scheme == %q)\nreturn config\n", quoted, "tokyonight_"+variant.style)), 0600)
			requireRun(wezterm, "--config-file", check, "show-keys")
		}
		alacritty, err := exec.LookPath("alacritty")
		if err != nil {
			candidate := "/Applications/Alacritty.app/Contents/MacOS/alacritty"
			if _, err := os.Stat(candidate); err == nil {
				alacritty = candidate
			}
		}
		if alacritty != "" {
			content := requireRun(alacritty, "migrate", "--dry-run", "--config-file", filepath.Join(configRoot, "alacritty/alacritty.toml"))
			if !strings.Contains(content, variant.background) {
				t.Fatalf("Alacritty missing %s", variant.background)
			}
		}
	}
	write(filepath.Join(directory, "preference"), []byte("broken\n"), 0600)
	output, err := run(binary, "status", "--config", settings)
	if err == nil {
		t.Fatal("corrupt preference incorrectly reported success")
	}
	// stderr includes the diagnostic after the JSON, so decode only the first object.
	var reported map[string]any
	if err := json.NewDecoder(strings.NewReader(output)).Decode(&reported); err != nil {
		t.Fatal(err)
	}
	if reported["preference"] != "broken" || !strings.Contains(fmt.Sprint(reported["state_error"]), "preference") {
		t.Fatalf("incorrect status: %v", reported)
	}
	retained, err := os.ReadFile(filepath.Join(directory, "preference"))
	if err != nil || string(retained) != "broken\n" {
		t.Fatal("corrupt preference was overwritten")
	}
}
