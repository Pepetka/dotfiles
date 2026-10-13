package main

import (
	"context"
	"dotfiles/theme/internal/adapters"
	"dotfiles/theme/internal/engine"
	"dotfiles/theme/internal/service"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func execute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: theme {auto|dark|light|sync|status|watch|authorize}")
	}
	command := args[0]
	switch command {
	case "auto", "dark", "light", "sync", "status", "watch", "authorize":
	default:
		return fmt.Errorf("unknown theme command %q", command)
	}
	options := flag.NewFlagSet("theme", flag.ContinueOnError)
	directory := options.String("theme-dir", engine.DefaultDirectory(), "state directory")
	configPath := options.String("config", "", "engine configuration file")
	if err := options.Parse(args[1:]); err != nil {
		return err
	}
	if options.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", options.Args())
	}
	if *configPath == "" {
		*configPath = filepath.Join(*directory, "engine.json")
	}
	data, err := os.ReadFile(*configPath)
	if err != nil {
		return fmt.Errorf("theme service is not installed; run theme install: %w", err)
	}
	var config engine.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}
	if config.Directory == "" || config.Root == "" {
		return fmt.Errorf("incomplete engine configuration")
	}
	// Reject before touching state, so the persisted preference stays manual.
	if config.Manual() && (command == "auto" || command == "watch") {
		return fmt.Errorf("theme %s: %w (installed in manual mode)", command, engine.ErrNoBackend)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	e := engine.Engine{Config: config, Adapters: adapters.All(), Backend: engine.NewBackend(config)}
	switch command {
	case "watch":
		e.Watch(ctx)
		return nil
	case "status":
		status, stateErr := engine.CurrentStatus(ctx, config.Directory)
		status.Backend, status.Service = describe(config)
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(status); err != nil {
			return err
		}
		return stateErr
	case "authorize":
		state, err := adapters.GhosttyRequest(ctx, config, "ghostty-authorize")
		fmt.Println(state)
		if err == nil && state == "authorized" {
			_, err = e.Reconcile(ctx, "", nil, true)
		}
		return err
	default:
		var preference engine.Mode
		if command != "sync" {
			preference = engine.Mode(command)
		}
		status, err := e.Reconcile(ctx, preference, nil, command == "sync")
		if status.Mode != "" {
			fmt.Printf("Theme: %s (preference: %s)\n", status.Mode, status.Preference)
		}
		return err
	}
}

func describe(c engine.Config) (string, string) {
	if c.Manual() {
		return "manual", "none"
	}
	name := c.ServiceManager()
	if name == "" {
		return c.Backend, "none"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return c.Backend, name + ": " + err.Error()
	}
	manager, err := service.New(name, home)
	if err != nil {
		return c.Backend, err.Error()
	}
	return c.Backend, name + ": " + manager.State()
}

func main() {
	if err := execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
