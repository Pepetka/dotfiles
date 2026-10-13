// Package service manages the per-user background service that runs the theme
// watcher: a LaunchAgent on macOS and a systemd user unit on Linux.
package service

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const Label = "com.vlvkuznetsov.theme-sync"

// The installer owns definition files, backups and rollback; a Manager only
// renders, validates and controls the loaded service.
type Manager interface {
	Name() string
	Path() string
	Definition(directory string) []byte
	Validate(path string) error
	Start() error
	Stop()
	Loaded() bool
	State() string
}

func New(name, home string) (Manager, error) {
	switch name {
	case "launchd":
		return Launchd{Home: home, Domain: fmt.Sprintf("gui/%d", os.Getuid())}, nil
	case "systemd":
		return Systemd{Home: home}, nil
	}
	return nil, fmt.Errorf("unknown service manager %q", name)
}

func run(name string, args ...string) (string, error) {
	data, err := exec.Command(name, args...).CombinedOutput()
	output := strings.TrimSpace(string(data))
	if err != nil {
		return output, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, output)
	}
	return output, nil
}
