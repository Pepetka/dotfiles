package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"dotfiles/theme/internal/engine"
)

type Ghostty struct{}

func (Ghostty) Name() string { return "ghostty" }
func (Ghostty) Render(c engine.Config, mode engine.Mode) ([]engine.Artifact, error) {
	data, err := os.ReadFile(c.Resource("ghostty", "theme-"+string(mode)+".conf"))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("empty Ghostty theme")
	}
	return []engine.Artifact{{Name: "ghostty-theme.conf", Data: data}}, nil
}

type ghosttyReply struct {
	State string `json:"state"`
	Error string `json:"error"`
}

// Launch through LaunchServices so Automation belongs to a consistent helper
// app, both during interactive setup and when called by the LaunchAgent.
func GhosttyRequest(ctx context.Context, c engine.Config, command string) (string, error) {
	return ghosttyRequest(ctx, c, command, func(ctx context.Context, path string) error {
		output, err := engine.Run(ctx, "/usr/bin/open", "-g", "-j", "-n", "-a", c.PlatformHelper, "--args", command, path)
		if err != nil {
			return fmt.Errorf("Ghostty helper: %w: %s", err, output)
		}
		return nil
	})
}

func ghosttyRequest(ctx context.Context, c engine.Config, command string, launch func(context.Context, string) error) (string, error) {
	timeout := engine.CommandTimeout
	if command == "ghostty-authorize" {
		timeout = 90 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	directory, err := os.MkdirTemp(c.Directory, ".ghostty-request-*")
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "reply.json")
	// Removing the private directory also prevents a late helper from creating
	// an orphan reply after this request timed out.
	defer os.RemoveAll(directory)
	if err := launch(ctx, path); err != nil {
		return "", err
	}
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		data, _ := os.ReadFile(path)
		var reply ghosttyReply
		if json.Unmarshal(data, &reply) == nil && reply.State != "" {
			if reply.Error != "" {
				return reply.State, fmt.Errorf("%s", reply.Error)
			}
			return reply.State, nil
		}
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("Ghostty helper: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func (Ghostty) Apply(ctx context.Context, c engine.Config, r engine.ApplyRequest) (engine.AdapterState, error) {
	if runtime.GOOS != "darwin" {
		return engine.AdapterState{State: "published", Mode: r.Mode}, nil
	}
	if c.PlatformHelper == "" {
		return engine.AdapterState{State: "published", Mode: r.Mode}, nil
	}
	executable := filepath.Join(c.PlatformHelper, "Contents", "MacOS", "theme-macos")
	output, err := engine.Run(ctx, executable, "ghostty-state")
	if err != nil {
		return engine.AdapterState{}, fmt.Errorf("Ghostty discovery: %w: %s", err, output)
	}
	identity := strings.TrimSpace(output)
	return applyGhostty(r, identity, func() (string, error) { return GhosttyRequest(ctx, c, "ghostty-reload") })
}

func applyGhostty(r engine.ApplyRequest, identity string, reload func() (string, error)) (engine.AdapterState, error) {
	if identity == "" {
		return engine.AdapterState{State: "not running"}, nil
	}
	if r.Force || r.Changed || r.Previous.State != "reload requested" || r.Previous.Mode != r.Mode || r.Previous.Identity != identity {
		state, err := reload()
		if err != nil {
			return engine.AdapterState{}, err
		}
		if state == "not running" {
			return engine.AdapterState{State: state}, nil
		}
		if state != "reload requested" {
			return engine.AdapterState{}, fmt.Errorf("unexpected Ghostty reply %q", state)
		}
	}
	return engine.AdapterState{State: "reload requested", Mode: r.Mode, Identity: identity}, nil
}
