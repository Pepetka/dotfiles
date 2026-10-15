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
func (Ghostty) Render(c engine.Config, r engine.RenderRequest) ([]engine.Artifact, error) {
	mode := r.Mode
	if r.Preference == engine.Auto {
		mode = engine.Auto
	}
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
	if runtime.GOOS == "linux" {
		return applyLinuxGhostty(r, func(args ...string) (string, error) {
			return engine.Run(ctx, "systemctl", append([]string{"--user"}, args...)...)
		})
	}
	if runtime.GOOS != "darwin" {
		return engine.AdapterState{State: "published", Mode: ghosttyMode(r)}, nil
	}
	if c.PlatformHelper == "" {
		return engine.AdapterState{State: "published", Mode: ghosttyMode(r)}, nil
	}
	executable := filepath.Join(c.PlatformHelper, "Contents", "MacOS", "theme-macos")
	output, err := engine.Run(ctx, executable, "ghostty-state")
	if err != nil {
		return engine.AdapterState{}, fmt.Errorf("Ghostty discovery: %w: %s", err, output)
	}
	identity := strings.TrimSpace(output)
	return applyGhostty(r, identity, func() (string, error) { return GhosttyRequest(ctx, c, "ghostty-reload") })
}

func ghosttyMode(r engine.ApplyRequest) engine.Mode {
	if r.Preference == engine.Auto {
		return engine.Auto
	}
	return r.Mode
}

func applyGhostty(r engine.ApplyRequest, identity string, reload func() (string, error)) (engine.AdapterState, error) {
	mode := ghosttyMode(r)
	if identity == "" {
		return engine.AdapterState{State: "not running"}, nil
	}
	if r.Force || r.Changed || r.Previous.State != "reload requested" || r.Previous.Mode != mode || r.Previous.Identity != identity {
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
	return engine.AdapterState{State: "reload requested", Mode: mode, Identity: identity}, nil
}

const ghosttyUnit = "app-com.mitchellh.ghostty.service"

// On Linux, Ghostty's own systemd unit maps reload to SIGUSR2 for its main
// process. Signalling Ghostty directly is unsafe: versions without a handler
// terminate on SIGUSR2. Instances outside the unit keep the published file.
func applyLinuxGhostty(r engine.ApplyRequest, systemctl func(...string) (string, error)) (engine.AdapterState, error) {
	output, err := systemctl("show", "--property=ActiveState,MainPID,CanReload", ghosttyUnit)
	properties := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		if name, value, ok := strings.Cut(strings.TrimSpace(line), "="); ok {
			properties[name] = value
		}
	}
	if err != nil || properties["ActiveState"] != "active" || properties["CanReload"] != "yes" || properties["MainPID"] == "0" {
		return engine.AdapterState{State: "published", Mode: ghosttyMode(r)}, nil
	}
	return applyGhostty(r, properties["MainPID"], func() (string, error) {
		if output, err := systemctl("reload", ghosttyUnit); err != nil {
			return "", fmt.Errorf("Ghostty reload: %w: %s", err, strings.TrimSpace(output))
		}
		return "reload requested", nil
	})
}
