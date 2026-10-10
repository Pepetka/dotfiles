// Package install owns macOS service installation, independently of theme policy.
package install

import (
	"bytes"
	"crypto/sha256"
	"dotfiles/theme/internal/engine"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const label = "com.vlvkuznetsov.theme-sync"

var artifacts = []string{"theme-sync", "macos-backend", "engine.json", ".platform-source.sha256", "wezterm-bridge.lua", "Theme Automation.app"}

type installer struct{ source, root, directory, home, domain, plist string }

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
	// Native cp preserves signed bundle metadata and symlinks, including dangling prompt links.
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

func servicePlist(directory string) []byte {
	escape := func(s string) string { var b bytes.Buffer; _ = xml.EscapeText(&b, []byte(s)); return b.String() }
	var args strings.Builder
	for _, value := range []string{filepath.Join(directory, "theme-sync"), "watch", "--theme-dir", directory} {
		fmt.Fprintf(&args, "<string>%s</string>", escape(value))
	}
	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>%s</string>
<key>ProgramArguments</key><array>%s</array>
<key>RunAtLoad</key><true/><key>KeepAlive</key><true/>
<key>ThrottleInterval</key><integer>5</integer>
<key>LimitLoadToSessionType</key><string>Aqua</string>
<key>ProcessType</key><string>Background</string>
<key>StandardErrorPath</key><string>%s</string>
</dict></plist>
`, label, args.String(), escape(filepath.Join(directory, "service.log"))))
}

func Execute(source string, args []string) error {
	if runtime.GOOS != "darwin" {
		return errors.New("this service installer requires macOS")
	}
	if len(args) == 0 || (args[0] != "install" && args[0] != "uninstall") {
		return errors.New("usage: theme {install [--backend PATH]|uninstall}")
	}
	backend := ""
	if len(args) != 1 {
		if args[0] != "install" || len(args) != 3 || args[1] != "--backend" {
			return errors.New("usage: theme install [--backend PATH]")
		}
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
	}
	source, err := filepath.Abs(source)
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
	i := installer{
		source: source, root: root, directory: directory, home: home,
		domain: fmt.Sprintf("gui/%d", os.Getuid()), plist: filepath.Join(home, "Library/LaunchAgents", label+".plist"),
	}
	if args[0] == "uninstall" {
		_ = i.launch("bootout", i.domain+"/"+label)
		for _, name := range artifacts {
			if name == "wezterm-bridge.lua" {
				continue
			}
			if err := os.RemoveAll(filepath.Join(directory, name)); err != nil {
				return err
			}
		}
		if err := os.RemoveAll(i.plist); err != nil {
			return err
		}
		fmt.Println("Theme service removed; preference, last palettes and prompt retained.")
		return nil
	}
	return i.install(backend)
}

func (i installer) launch(args ...string) error {
	cmd := exec.Command("launchctl", args...)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl %s: %w: %s", strings.Join(args, " "), err, data)
	}
	return nil
}

func (i installer) install(backend string) error {
	for _, path := range []string{i.directory, filepath.Dir(i.plist)} {
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
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
	swift := func(source, target string) error {
		return command("", "xcrun", "swiftc", "-module-cache-path", path("module-cache"), "-parse-as-library", "-O", filepath.Join(i.source, source), "-o", target)
	}
	if err := swift("backends/macos/main.swift", path("macos-backend")); err != nil {
		return err
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
	if err := os.WriteFile(path(".platform-source.sha256"), []byte(hash+"\n"), 0600); err != nil {
		return err
	}
	for _, name := range []string{"theme-sync", "macos-backend"} {
		if err := os.Chmod(path(name), 0700); err != nil {
			return err
		}
	}
	if backend == "" {
		old, err := os.ReadFile(filepath.Join(i.directory, "engine.json"))
		if err == nil {
			var c engine.Config
			if err := json.Unmarshal(old, &c); err != nil {
				return err
			}
			backend = c.Backend
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	if backend == "" {
		backend = filepath.Join(i.directory, "macos-backend")
	}
	tmux, _ := exec.LookPath("tmux")
	socketRoot := os.Getenv("TMUX_TMPDIR")
	if socketRoot == "" {
		socketRoot = "/tmp"
	}
	c := engine.Config{Root: i.root, Directory: i.directory, Backend: backend, PlatformHelper: filepath.Join(i.directory, "Theme Automation.app"), Tmux: tmux, Socket: filepath.Join(socketRoot, fmt.Sprintf("tmux-%d", os.Getuid()), "default")}
	if err := writeJSON(path("engine.json"), c); err != nil {
		return err
	}
	if err := os.WriteFile(path("service.plist"), servicePlist(i.directory), 0600); err != nil {
		return err
	}
	if err := copyPath(filepath.Join(i.source, "wezterm.lua"), path("wezterm-bridge.lua")); err != nil {
		return err
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
	c.Directory, c.Tmux, c.PlatformHelper = preflight, "", ""
	if backend == filepath.Join(i.directory, "macos-backend") {
		c.Backend = path("macos-backend")
	}
	if err := writeJSON(path("preflight.json"), c); err != nil {
		return err
	}
	for _, plist := range []string{path("service.plist"), filepath.Join(app, "Contents/Info.plist")} {
		if err := command("", "plutil", "-lint", plist); err != nil {
			return err
		}
	}
	if err := command("", path("theme-sync"), "sync", "--config", path("preflight.json")); err != nil {
		return err
	}
	return i.commit(build)
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
	if exists(i.plist) {
		if err := copyPath(i.plist, filepath.Join(previous, "service.plist")); err != nil {
			return err
		}
	}
	wasLoaded := i.launch("print", i.domain+"/"+label) == nil
	_ = i.launch("bootout", i.domain+"/"+label)
	rollback := func(cause error) error {
		_ = i.launch("bootout", i.domain+"/"+label)
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
		if err := os.RemoveAll(i.plist); err != nil {
			failures = append(failures, err)
		}
		if exists(filepath.Join(previous, "service.plist")) {
			if err := copyPath(filepath.Join(previous, "service.plist"), i.plist); err != nil {
				failures = append(failures, err)
			} else if wasLoaded {
				if err := i.launch("bootstrap", i.domain, i.plist); err != nil {
					failures = append(failures, err)
				}
			}
		}
		return fmt.Errorf("installation failed; attempted restoration of previous service: %w", errors.Join(failures...))
	}
	for _, name := range artifacts {
		target := filepath.Join(i.directory, name)
		if err := os.RemoveAll(target); err != nil {
			return rollback(err)
		}
		if err := os.Rename(filepath.Join(build, name), target); err != nil {
			return rollback(err)
		}
	}
	// LaunchAgents may reside on another filesystem from XDG_CONFIG_HOME.
	plistData, err := os.ReadFile(filepath.Join(build, "service.plist"))
	if err != nil {
		return rollback(err)
	}
	if _, err := engine.AtomicWrite(i.plist, plistData); err != nil {
		return rollback(err)
	}
	if err := i.launch("bootstrap", i.domain, i.plist); err != nil {
		return rollback(err)
	}
	prompt := filepath.Join(i.home, ".p10k.zsh")
	backup := filepath.Join(i.directory, "p10k-before-install.zsh")
	if exists(prompt) && !exists(backup) {
		if err := copyPath(prompt, backup); err != nil {
			return err
		}
	}
	tmp, err := os.MkdirTemp(i.home, ".theme-prompt.")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	link := filepath.Join(tmp, "p10k.zsh")
	if err := os.Symlink(filepath.Join(i.source, "resources/p10k/p10k.zsh"), link); err != nil {
		return err
	}
	if err := os.Rename(link, prompt); err != nil {
		return err
	}
	fmt.Println("Theme engine installed. Reload existing terminal configs once to adopt the engine.")
	fmt.Println("For Ghostty live reload, run theme authorize with Ghostty open and allow Automation.")
	return nil
}
