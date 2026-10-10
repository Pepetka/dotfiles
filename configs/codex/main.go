package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

var (
	configFiles = []string{"config.toml", "work.config.toml", "deep.config.toml", "review.config.toml"}
	themeFiles  = []string{"Tokyo Night Moon.tmTheme", "Tokyo Night Day.tmTheme"}
)

// Version 1 remains compatible with snapshots written by the former Python installer.
type state struct {
	Version int                       `json:"version"`
	Files   map[string]map[string]any `json:"files"`
	Themes  map[string]string         `json:"themes,omitempty"`
}
type options struct {
	Source, Home string
	DryRun       bool
	Output       io.Writer
	Write        func(string, []byte) error
	BeforeCommit func()
}

func readOptional(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return b, err
}
func digest(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func atomicWrite(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".codex-sync-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err == nil {
		err = f.Sync()
	}
	err = errors.Join(err, f.Close())
	if err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func synchronize(o options) error {
	if o.Output == nil {
		o.Output = io.Discard
	}
	if o.Write == nil {
		o.Write = atomicWrite
	}
	stateDir := filepath.Join(o.Home, ".dotfiles-sync")
	stateFile := filepath.Join(stateDir, "state.json")
	stateBytes, err := readOptional(stateFile)
	if err != nil {
		return err
	}
	previous := state{Version: 1, Files: map[string]map[string]any{}, Themes: map[string]string{}}
	if stateBytes != nil {
		if err = json.Unmarshal(stateBytes, &previous); err != nil {
			return fmt.Errorf("invalid synchronization state: %w", err)
		}
		if previous.Version != 1 || previous.Files == nil {
			return fmt.Errorf("unsupported synchronization state")
		}
	}
	next := state{Version: 1, Files: map[string]map[string]any{}, Themes: map[string]string{}}
	originals, outputs := map[string][]byte{}, map[string][]byte{}
	conflicts := []string{}
	for _, name := range configFiles {
		target := filepath.Join(o.Home, name)
		info, statErr := os.Lstat(target)
		if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
			return statErr
		}
		if info != nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("expected a local file, found a symlink: %s", target)
		}
		data, err := os.ReadFile(filepath.Join(o.Source, name))
		if err != nil {
			return err
		}
		source, err := decode(data)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		originals[target], err = readOptional(target)
		if err != nil {
			return err
		}
		result, issues, err := merge(originals[target], source, previous.Files[name], name)
		if err != nil {
			return fmt.Errorf("%s: %w", target, err)
		}
		conflicts = append(conflicts, issues...)
		next.Files[name] = source
		if !bytes.Equal(result, originals[target]) {
			outputs[target] = result
		}
	}
	if len(conflicts) > 0 {
		return fmt.Errorf("both local and shared settings changed; nothing written:\n%s", strings.Join(conflicts, "\n"))
	}
	themes := filepath.Join(o.Source, "..", "..", "theme", "resources", "codex")
	for _, name := range themeFiles {
		target := filepath.Join(o.Home, "themes", name)
		data, err := os.ReadFile(filepath.Join(themes, name))
		if err != nil {
			return err
		}
		originals[target], err = readOptional(target)
		if err != nil {
			return err
		}
		next.Themes[name] = digest(data)
		if !bytes.Equal(data, originals[target]) {
			if originals[target] != nil && digest(originals[target]) != previous.Themes[name] {
				return fmt.Errorf("existing custom theme differs; nothing written: %s", target)
			}
			outputs[target] = data
		}
	}
	paths := []string{}
	for p := range outputs {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		action := "Update "
		if o.DryRun {
			action = "Would update "
		}
		fmt.Fprintln(o.Output, action+p)
	}
	if o.DryRun {
		fmt.Fprintln(o.Output, "Dry run: no configuration or synchronization state written.")
		return nil
	}
	if err = os.MkdirAll(stateDir, 0700); err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(stateDir, "lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if o.BeforeCommit != nil {
		o.BeforeCommit()
	}
	check := func(path string, old []byte) error {
		actual, err := readOptional(path)
		if err != nil {
			return err
		}
		if (actual == nil) != (old == nil) || !bytes.Equal(actual, old) {
			return fmt.Errorf("configuration changed during synchronization; retry")
		}
		return nil
	}
	if err = check(stateFile, stateBytes); err != nil {
		return err
	}
	for p, old := range originals {
		if err = check(p, old); err != nil {
			return err
		}
	}
	backup := filepath.Join(stateDir, "backups", time.Now().UTC().Format("20060102T150405.000000000Z"))
	written := []string{}
	rollback := func(cause error) error {
		for i := len(written) - 1; i >= 0; i-- {
			path := written[i]
			var restoreErr error
			if originals[path] == nil {
				restoreErr = os.Remove(path)
			} else {
				restoreErr = atomicWrite(path, originals[path])
			}
			if restoreErr != nil {
				cause = errors.Join(cause, fmt.Errorf("rollback %s: %w", path, restoreErr))
			}
		}
		return cause
	}
	for _, path := range paths {
		if originals[path] != nil {
			rel, err := filepath.Rel(o.Home, path)
			if err != nil {
				return rollback(err)
			}
			if err = o.Write(filepath.Join(backup, rel), originals[path]); err != nil {
				return rollback(err)
			}
		}
		if err = o.Write(path, outputs[path]); err != nil {
			return rollback(err)
		}
		written = append(written, path)
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return rollback(err)
	}
	if err = o.Write(stateFile, append(data, '\n')); err != nil {
		return rollback(err)
	}
	fmt.Fprintln(o.Output, "Codex settings synchronized; local values and manual theme selection preserved.")
	return nil
}

func run() error {
	source := flag.String("source", ".", "shared Codex configuration directory")
	dry := flag.Bool("dry-run", false, "preview without changing configuration or sync state")
	flag.Parse()
	if flag.NArg() > 0 {
		return fmt.Errorf("unexpected arguments")
	}
	home := os.Getenv("CODEX_HOME")
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		home = filepath.Join(userHome, ".codex")
	}
	return synchronize(options{Source: *source, Home: home, DryRun: *dry, Output: os.Stdout})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
