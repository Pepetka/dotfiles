package adapters

import (
	"context"
	"dotfiles/theme/internal/engine"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type Tmux struct{}

func (Tmux) Name() string { return "tmux" }
func (Tmux) Render(c engine.Config, r engine.RenderRequest) ([]engine.Artifact, error) {
	return palette(c, r.Mode, "tmux", "conf", "tmux-theme.conf")
}

func (Tmux) Apply(ctx context.Context, c engine.Config, r engine.ApplyRequest) (engine.AdapterState, error) {
	if c.Tmux == "" {
		return engine.AdapterState{State: "not installed"}, nil
	}
	info, err := os.Stat(c.Socket)
	if os.IsNotExist(err) {
		return engine.AdapterState{State: "not running"}, nil
	}
	if err != nil {
		return engine.AdapterState{}, err
	}
	identity := fmt.Sprint(info.Sys().(*syscall.Stat_t).Ino)
	run := func(args ...string) (string, error) {
		output, err := engine.Run(ctx, c.Tmux, append([]string{"-S", c.Socket}, args...)...)
		if err != nil {
			err = fmt.Errorf("tmux: %w: %s", err, strings.TrimSpace(output))
		}
		return output, err
	}
	marker, err := run("show-options", "-gqv", "@theme_mode")
	if err != nil {
		if strings.Contains(err.Error(), "no server running") || strings.Contains(err.Error(), "Connection refused") {
			return engine.AdapterState{State: "not running"}, nil
		}
		return engine.AdapterState{}, err
	}
	if r.Force || r.Changed || r.Previous.State == "error" || r.Previous.Identity != identity || r.Previous.Mode != r.Mode || strings.TrimSpace(marker) != string(r.Mode) {
		path := filepath.Join(c.Directory, "tmux-theme.conf")
		if _, err := run("source-file", path); err != nil {
			return engine.AdapterState{}, err
		}
		if _, err := run("set-option", "-g", "@theme_mode", string(r.Mode)); err != nil {
			return engine.AdapterState{}, err
		}
		clients, err := run("list-clients", "-F", "#{client_name}")
		if err != nil {
			return engine.AdapterState{}, err
		}
		for _, client := range strings.Split(strings.TrimSpace(clients), "\n") {
			if client != "" {
				if _, err := run("refresh-client", "-t", client); err != nil {
					return engine.AdapterState{}, err
				}
			}
		}
	}
	return engine.AdapterState{State: "applied", Mode: r.Mode, Identity: identity}, nil
}
