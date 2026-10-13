package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Mode string

const (
	Dark  Mode = "dark"
	Light Mode = "light"
	Auto  Mode = "auto"
)

func ParseMode(value string, allowAuto bool) (Mode, error) {
	mode := Mode(strings.TrimSpace(value))
	if mode == Dark || mode == Light || (allowAuto && mode == Auto) {
		return mode, nil
	}
	return "", fmt.Errorf("invalid theme mode %q", value)
}

// An empty Backend selects manual mode: dark and light work without a service.
type Config struct {
	Root           string `json:"root"`
	Directory      string `json:"directory"`
	Backend        string `json:"backend"`
	Service        string `json:"service,omitempty"`
	PlatformHelper string `json:"platform_helper,omitempty"`
	Tmux           string `json:"tmux,omitempty"`
	Socket         string `json:"socket,omitempty"`
}

func (c Config) Manual() bool { return c.Backend == "" }

// ServiceManager names the installed service; engines installed before the
// field existed always used a macOS LaunchAgent.
func (c Config) ServiceManager() string {
	if c.Service == "" && !c.Manual() && runtime.GOOS == "darwin" {
		return "launchd"
	}
	return c.Service
}

func (c Config) Resource(parts ...string) string {
	return filepath.Join(append([]string{c.Root, "theme", "resources"}, parts...)...)
}

type Artifact struct {
	Name string
	Data []byte
}

// State distinguishes publication from a confirmed reload request. Mode and
// Identity describe the last successful application, not just the desired mode.
type AdapterState struct {
	State    string `json:"state"`
	Mode     Mode   `json:"mode,omitempty"`
	Identity string `json:"identity,omitempty"`
	Error    string `json:"error,omitempty"`
}

type Status struct {
	Preference   Mode                    `json:"preference"`
	System       Mode                    `json:"system,omitempty"`
	Mode         Mode                    `json:"mode"`
	Backend      string                  `json:"backend,omitempty"`
	Service      string                  `json:"service,omitempty"`
	BackendError string                  `json:"backend_error,omitempty"`
	StateError   string                  `json:"state_error,omitempty"`
	Adapters     map[string]AdapterState `json:"adapters"`
}

type ApplyRequest struct {
	Mode     Mode
	Previous AdapterState
	Changed  bool
	Force    bool
}

// Render has no side effects. The engine owns locking and atomic publication;
// Apply owns application-specific reloads and reports errors independently.
type Adapter interface {
	Name() string
	Render(Config, Mode) ([]Artifact, error)
	Apply(context.Context, Config, ApplyRequest) (AdapterState, error)
}

type System struct {
	Mode Mode
	Err  error
}

type Engine struct {
	Config   Config
	Adapters []Adapter
	Backend  Backend
}

func DefaultDirectory() string {
	root := os.Getenv("XDG_CONFIG_HOME")
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".config")
	}
	return filepath.Join(root, "theme")
}

const CommandTimeout = 3 * time.Second

// Applications read the published files at startup, so a healthy daemon only
// rechecks reload targets periodically; failures are retried every second.
var (
	RetryInterval  = time.Second
	HealthInterval = 30 * time.Second
)
