// Package install owns engine and service installation, independently of theme policy.
package install

import (
	"context"
	"crypto/sha256"
	"dotfiles/theme/internal/engine"
	"dotfiles/theme/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const usage = "usage: theme {install [--manual | --backend PATH]|uninstall}"

// Records that manual installation replaced an auto policy, so a later
// automatic installation can follow the system again.
const manualMarker = ".auto-before-manual"

// Every artifact any platform or mode may install; switching modes removes the
// ones the new build does not provide.
var artifacts = []string{"theme-sync", "macos-backend", "linux-backend", "engine.json", ".platform-source.sha256", "wezterm-bridge.lua", "Theme Automation.app"}

type installer struct {
	source, root, directory string
	manual                  bool
	service                 service.Manager
}

func command(directory, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func exists(path string) bool { _, err := os.Lstat(path); return err == nil }
func copyPath(source, target string) error {
	// Native cp preserves signed bundle metadata and symlinks.
	return command("", "cp", "-pPR", source, target)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0600)
}

func fingerprint(source string) (string, error) {
	hash := sha256.New()
	hash.Write([]byte("macos-platform-helper-v1\x00"))
	for _, name := range []string{"main.swift", "Info.plist"} {
		data, err := os.ReadFile(filepath.Join(source, "platform/macos", name))
		if err != nil {
			return "", err
		}
		hash.Write([]byte(name + "\x00"))
		hash.Write(data)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func platformService() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "launchd", nil
	case "linux":
		return "systemd", nil
	}
	return "", fmt.Errorf("theme installation is not supported on %s", runtime.GOOS)
}

func platformBackend() string {
	return map[string]string{"darwin": "macos-backend", "linux": "linux-backend"}[runtime.GOOS]
}

func Execute(source string, args []string) error {
	if len(args) == 0 || (args[0] != "install" && args[0] != "uninstall") {
		return errors.New(usage)
	}
	backend, manual := "", false
	switch {
	case len(args) == 1:
	case args[0] == "install" && len(args) == 2 && args[1] == "--manual":
		manual = true
	case args[0] == "install" && len(args) == 3 && args[1] == "--backend":
		var err error
		backend, err = filepath.Abs(args[2])
		if err != nil {
			return err
		}
		info, err := os.Stat(backend)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			return fmt.Errorf("backend is not executable: %s", backend)
		}
	default:
		return errors.New(usage)
	}
	name, err := platformService()
	if err != nil {
		return err
	}
	source, err = filepath.Abs(source)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	root := os.Getenv("DOTFILES")
	if root == "" {
		root = filepath.Dir(source)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	directory, err := filepath.Abs(engine.DefaultDirectory())
	if err != nil {
		return err
	}
	manager, err := service.New(name, home)
	if err != nil {
		return err
	}
	i := installer{source: source, root: root, directory: directory, manual: manual, service: manager}
	if args[0] == "uninstall" {
		return i.uninstall()
	}
	return i.install(backend)
}

func (i installer) uninstall() error {
	for _, name := range artifacts {
		if name == "wezterm-bridge.lua" {
			continue
		}
		if err := os.RemoveAll(filepath.Join(i.directory, name)); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(i.service.Path()); err != nil {
		return err
	}
	i.service.Stop()
	fmt.Println("Theme engine removed; preference and last palettes retained.")
	return nil
}

func (i installer) buildMacOS(build string) error {
	path := func(name string) string { return filepath.Join(build, name) }
	swift := func(source, target string) error {
		return command("", "xcrun", "swiftc", "-module-cache-path", path("module-cache"), "-parse-as-library", "-O", filepath.Join(i.source, source), "-o", target)
	}
	if !i.manual {
		if err := swift("backends/macos/main.swift", path("macos-backend")); err != nil {
			return err
		}
		if err := os.Chmod(path("macos-backend"), 0700); err != nil {
			return err
		}
	}
	hash, err := fingerprint(i.source)
	if err != nil {
		return err
	}
	app := path("Theme Automation.app")
	installedApp := filepath.Join(i.directory, "Theme Automation.app")
	oldHash, _ := os.ReadFile(filepath.Join(i.directory, ".platform-source.sha256"))
	info, statErr := os.Stat(filepath.Join(installedApp, "Contents/MacOS/theme-macos"))
	if statErr == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 && strings.TrimSpace(string(oldHash)) == hash {
		if err := copyPath(installedApp, app); err != nil {
			return err
		}
	} else {
		if err := os.MkdirAll(filepath.Join(app, "Contents/MacOS"), 0700); err != nil {
			return err
		}
		if err := swift("platform/macos/main.swift", filepath.Join(app, "Contents/MacOS/theme-macos")); err != nil {
			return err
		}
		if err := copyPath(filepath.Join(i.source, "platform/macos/Info.plist"), filepath.Join(app, "Contents/Info.plist")); err != nil {
			return err
		}
		if err := command("", "codesign", "--force", "--sign", "-", "--identifier", "com.vlvkuznetsov.theme-automation", app); err != nil {
			return err
		}
	}
	if err := command("", "codesign", "--verify", "--strict", app); err != nil {
		return err
	}
	if err := command("", "plutil", "-lint", filepath.Join(app, "Contents/Info.plist")); err != nil {
		return err
	}
	return os.WriteFile(path(".platform-source.sha256"), []byte(hash+"\n"), 0600)
}

func (i installer) buildLinux(build string) error {
	if i.manual {
		return nil
	}
	target := filepath.Join(build, "linux-backend")
	if err := command(i.source, "go", "build", "-trimpath", "-o", target, "./backends/linux"); err != nil {
		return err
	}
	return os.Chmod(target, 0700)
}

// Manual mode has no system observer, so a persisted auto policy becomes the
// last effective mode instead of failing every later reconciliation.
func (i installer) manualPreference() (engine.Mode, error) {
	data, err := os.ReadFile(filepath.Join(i.directory, "preference"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err == nil && strings.TrimSpace(string(data)) != string(engine.Auto) {
		return "", nil
	}
	data, _ = os.ReadFile(filepath.Join(i.directory, "mode"))
	mode, err := engine.ParseMode(string(data), false)
	if err != nil {
		mode = engine.Dark
	}
	return mode, nil
}

func (i installer) install(backend string) error {
	if err := os.MkdirAll(i.directory, 0700); err != nil {
		return err
	}
	build, err := os.MkdirTemp(i.directory, ".install.")
	if err != nil {
		return err
	}
	defer os.RemoveAll(build)
	path := func(name string) string { return filepath.Join(build, name) }
	if err := command(i.source, "go", "build", "-trimpath", "-o", path("theme-sync"), "./cmd/theme"); err != nil {
		return err
	}
	if err := os.Chmod(path("theme-sync"), 0700); err != nil {
		return err
	}
	helper := ""
	if runtime.GOOS == "darwin" {
		helper = filepath.Join(i.directory, "Theme Automation.app")
		err = i.buildMacOS(build)
	} else {
		err = i.buildLinux(build)
	}
	if err != nil {
		return err
	}
	defaultBackend := filepath.Join(i.directory, platformBackend())
	if !i.manual && backend == "" {
		old, err := os.ReadFile(filepath.Join(i.directory, "engine.json"))
		if err == nil {
			var c engine.Config
			if err := json.Unmarshal(old, &c); err != nil {
				return err
			}
			// Preserve only an explicitly chosen backend, not another mode's default.
			if c.Backend != filepath.Join(i.directory, "macos-backend") && c.Backend != filepath.Join(i.directory, "linux-backend") {
				backend = c.Backend
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if backend == "" {
			backend = defaultBackend
		}
	}
	tmux, _ := exec.LookPath("tmux")
	socketRoot := os.Getenv("TMUX_TMPDIR")
	if socketRoot == "" {
		socketRoot = "/tmp"
	}
	c := engine.Config{Root: i.root, Directory: i.directory, Backend: backend, PlatformHelper: helper, Tmux: tmux, Socket: filepath.Join(socketRoot, fmt.Sprintf("tmux-%d", os.Getuid()), "default")}
	if !i.manual {
		c.Service = i.service.Name()
	}
	if err := writeJSON(path("engine.json"), c); err != nil {
		return err
	}
	if err := copyPath(filepath.Join(i.source, "wezterm.lua"), path("wezterm-bridge.lua")); err != nil {
		return err
	}
	if !i.manual {
		if err := os.WriteFile(path("service"), i.service.Definition(i.directory), 0600); err != nil {
			return err
		}
		if err := i.service.Validate(path("service")); err != nil {
			return err
		}
	}
	preference := engine.Mode("")
	restoreAuto := !i.manual && exists(filepath.Join(i.directory, manualMarker))
	if i.manual {
		if preference, err = i.manualPreference(); err != nil {
			return err
		}
	} else if restoreAuto {
		preference = engine.Auto
	}
	preflight := path("preflight")
	if err := os.Mkdir(preflight, 0700); err != nil {
		return err
	}
	for _, name := range []string{"preference", "mode"} {
		from := filepath.Join(i.directory, name)
		if exists(from) {
			if err := copyPath(from, filepath.Join(preflight, name)); err != nil {
				return err
			}
		}
	}
	if preference != "" {
		if err := os.WriteFile(filepath.Join(preflight, "preference"), []byte(preference+"\n"), 0600); err != nil {
			return err
		}
	}
	c.Directory, c.Tmux, c.PlatformHelper = preflight, "", ""
	if backend == defaultBackend {
		c.Backend = path(platformBackend())
	}
	if !i.manual {
		if _, err := (engine.ExecutableBackend{Path: c.Backend}).Read(context.Background()); err != nil {
			return fmt.Errorf("the system backend cannot read the appearance; fix the desktop session or use: theme install --manual\n%w", err)
		}
	}
	if err := writeJSON(path("preflight.json"), c); err != nil {
		return err
	}
	if err := command("", path("theme-sync"), "sync", "--config", path("preflight.json")); err != nil {
		return err
	}
	if err := i.commit(build); err != nil {
		return err
	}
	if restoreAuto {
		if _, err := engine.AtomicWrite(filepath.Join(i.directory, "preference"), []byte("auto\n")); err != nil {
			return err
		}
		if err := os.Remove(filepath.Join(i.directory, manualMarker)); err != nil {
			return err
		}
		fmt.Println("Restored the auto preference replaced by manual installation.")
	}
	if !i.manual {
		fmt.Println("Theme engine installed with the " + i.service.Name() + " service. Reload existing terminal configs once to adopt the engine.")
		if runtime.GOOS == "darwin" {
			fmt.Println("For Ghostty live reload, run theme authorize with Ghostty open and allow Automation.")
		}
		return nil
	}
	if preference != "" {
		if _, err := engine.AtomicWrite(filepath.Join(i.directory, "preference"), []byte(preference+"\n")); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(i.directory, manualMarker), nil, 0600); err != nil {
			return err
		}
		fmt.Printf("Manual mode cannot follow the system; preference auto replaced with %s until the next theme install.\n", preference)
	}
	// No service publishes the palettes, so apply the current selection now.
	if err := command("", filepath.Join(i.directory, "theme-sync"), "sync", "--theme-dir", i.directory); err != nil {
		fmt.Fprintln(os.Stderr, "Initial synchronization reported errors; see: theme status")
	}
	fmt.Println("Theme engine installed in manual mode. Switch with theme dark or theme light.")
	return nil
}

func (i installer) commit(build string) error {
	previous := filepath.Join(build, "previous")
	if err := os.Mkdir(previous, 0700); err != nil {
		return err
	}
	for _, name := range artifacts {
		from := filepath.Join(i.directory, name)
		if exists(from) {
			if err := copyPath(from, filepath.Join(previous, name)); err != nil {
				return err
			}
		}
	}
	definition := i.service.Path()
	hadDefinition := exists(definition)
	if hadDefinition {
		if err := copyPath(definition, filepath.Join(previous, "service")); err != nil {
			return err
		}
	}
	wasLoaded := hadDefinition && i.service.Loaded()
	if hadDefinition || wasLoaded {
		i.service.Stop()
	}
	rollback := func(cause error) error {
		i.service.Stop()
		failures := []error{cause}
		for _, name := range artifacts {
			target := filepath.Join(i.directory, name)
			if err := os.RemoveAll(target); err != nil {
				failures = append(failures, err)
				continue
			}
			if exists(filepath.Join(previous, name)) {
				if err := copyPath(filepath.Join(previous, name), target); err != nil {
					failures = append(failures, err)
				}
			}
		}
		if err := os.RemoveAll(definition); err != nil {
			failures = append(failures, err)
		}
		if hadDefinition {
			if err := copyPath(filepath.Join(previous, "service"), definition); err != nil {
				failures = append(failures, err)
			} else if wasLoaded {
				if err := i.service.Start(); err != nil {
					failures = append(failures, err)
				}
			}
		}
		return fmt.Errorf("installation failed; attempted restoration of the previous installation: %w", errors.Join(failures...))
	}
	for _, name := range artifacts {
		target := filepath.Join(i.directory, name)
		if err := os.RemoveAll(target); err != nil {
			return rollback(err)
		}
		if exists(filepath.Join(build, name)) {
			if err := os.Rename(filepath.Join(build, name), target); err != nil {
				return rollback(err)
			}
		}
	}
	if i.manual {
		if err := os.RemoveAll(definition); err != nil {
			return rollback(err)
		}
		return nil
	}
	// Service definitions may reside on another filesystem from XDG_CONFIG_HOME.
	data, err := os.ReadFile(filepath.Join(build, "service"))
	if err != nil {
		return rollback(err)
	}
	if err := os.MkdirAll(filepath.Dir(definition), 0700); err != nil {
		return rollback(err)
	}
	if _, err := engine.AtomicWrite(definition, data); err != nil {
		return rollback(err)
	}
	if err := i.service.Start(); err != nil {
		return rollback(err)
	}
	return nil
}
