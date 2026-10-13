//go:build linux

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	portalName = "org.freedesktop.portal.Desktop"
	portalPath = "/org/freedesktop/portal/desktop"
	settings   = "org.freedesktop.portal.Settings"
	namespace  = "org.freedesktop.appearance"
	key        = "color-scheme"
)

// observer reads the XDG Desktop Portal, which covers GNOME, KDE and other
// desktops, and falls back to GNOME settings when no portal is available.
type observer struct{ conn *dbus.Conn }

func (o observer) portal() (string, error) {
	if o.conn == nil {
		return "", errors.New("no session bus")
	}
	object := o.conn.Object(portalName, portalPath)
	var value dbus.Variant
	err := object.Call(settings+".ReadOne", 0, namespace, key).Store(&value)
	var dbusErr dbus.Error
	if errors.As(err, &dbusErr) && dbusErr.Name == "org.freedesktop.DBus.Error.UnknownMethod" {
		// Portal versions before ReadOne wrap the value in a second variant.
		if err = object.Call(settings+".Read", 0, namespace, key).Store(&value); err == nil {
			if inner, ok := value.Value().(dbus.Variant); ok {
				value = inner
			}
		}
	}
	if err != nil {
		return "", fmt.Errorf("portal: %w", err)
	}
	scheme, ok := value.Value().(uint32)
	if !ok {
		return "", fmt.Errorf("portal: unexpected color-scheme %s", value)
	}
	return portalMode(scheme), nil
}

func gsettings(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output()
	if err != nil {
		return "", fmt.Errorf("gsettings: %w", err)
	}
	return gsettingsMode(string(output))
}

func (o observer) read(ctx context.Context) (string, error) {
	mode, portalErr := o.portal()
	if portalErr == nil {
		return mode, nil
	}
	mode, err := gsettings(ctx)
	if err != nil {
		return "", errors.Join(portalErr, err)
	}
	return mode, nil
}

// watch publishes the initial mode, then changes from SettingChanged signals.
// Periodic reads recover from a restarted portal or a missed signal.
func (o observer) watch(ctx context.Context, out io.Writer, interval time.Duration) error {
	// Subscribe before the initial read so a change in between is not lost.
	signals := make(chan *dbus.Signal, 16)
	if o.conn != nil {
		if err := o.conn.AddMatchSignal(dbus.WithMatchObjectPath(portalPath), dbus.WithMatchInterface(settings), dbus.WithMatchMember("SettingChanged")); err != nil {
			fmt.Fprintln(os.Stderr, "portal signals unavailable, polling:", err)
		}
		o.conn.Signal(signals)
		defer o.conn.RemoveSignal(signals)
	}
	current, err := o.read(ctx)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, current); err != nil {
		return err
	}
	publish := func(mode string) error {
		if mode == current {
			return nil
		}
		current = mode
		_, err := fmt.Fprintln(out, mode)
		return err
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-signals:
			if !ok {
				return errors.New("session bus connection closed")
			}
			if len(event.Body) != 3 || event.Body[0] != namespace || event.Body[1] != key {
				continue
			}
			value, ok := event.Body[2].(dbus.Variant)
			if !ok {
				continue
			}
			if scheme, ok := value.Value().(uint32); ok {
				if err := publish(portalMode(scheme)); err != nil {
					return err
				}
			}
		case <-ticker.C:
			mode, err := o.read(ctx)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				continue
			}
			if err := publish(mode); err != nil {
				return err
			}
		}
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var o observer
	if conn, err := dbus.ConnectSessionBus(); err == nil {
		o.conn = conn
		defer conn.Close()
	} else {
		fmt.Fprintln(os.Stderr, "session bus unavailable:", err)
	}
	var err error
	switch command := os.Args[1:]; {
	case len(command) == 1 && command[0] == "read":
		var mode string
		if mode, err = o.read(ctx); err == nil {
			fmt.Println(mode)
		}
	case len(command) == 1 && command[0] == "watch":
		err = o.watch(ctx, os.Stdout, 2*time.Second)
	default:
		err = errors.New("usage: linux-backend {read|watch}")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
