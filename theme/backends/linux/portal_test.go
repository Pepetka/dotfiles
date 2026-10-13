//go:build linux

package main

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

type fakePortal struct {
	mu    sync.Mutex
	value uint32
}

func (p *fakePortal) ReadOne(ns, name string) (dbus.Variant, *dbus.Error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if ns != namespace || name != key {
		return dbus.Variant{}, dbus.NewError("org.freedesktop.portal.Error.NotFound", nil)
	}
	return dbus.MakeVariant(p.value), nil
}

// legacyPortal predates ReadOne and double-wraps Read results.
type legacyPortal struct{ value uint32 }

func (p legacyPortal) Read(string, string) (dbus.Variant, *dbus.Error) {
	return dbus.MakeVariant(dbus.MakeVariant(p.value)), nil
}

// privateBus starts an isolated session bus; tests never touch the desktop's.
func privateBus(t *testing.T) string {
	t.Helper()
	daemon, err := exec.LookPath("dbus-daemon")
	if err != nil {
		t.Skip("dbus-daemon is not installed")
	}
	cmd := exec.Command(daemon, "--session", "--nofork", "--print-address")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	address, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(address)
}

func connect(t *testing.T, address string) *dbus.Conn {
	t.Helper()
	conn, err := dbus.Connect(address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func servePortal(t *testing.T, address string, portal any) *dbus.Conn {
	t.Helper()
	conn := connect(t, address)
	if err := conn.Export(portal, portalPath, settings); err != nil {
		t.Fatal(err)
	}
	if reply, err := conn.RequestName(portalName, dbus.NameFlagDoNotQueue); err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		t.Fatalf("portal name: %v, %v", reply, err)
	}
	return conn
}

func TestPortalReadAndWatch(t *testing.T) {
	address := privateBus(t)
	portal := &fakePortal{value: 2}
	server := servePortal(t, address, portal)
	o := observer{conn: connect(t, address)}
	if mode, err := o.read(context.Background()); err != nil || mode != "light" {
		t.Fatalf("read: %s, %v", mode, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reader, writer := io.Pipe()
	done := make(chan error, 1)
	go func() { done <- o.watch(ctx, writer, time.Hour); _ = writer.Close() }()
	lines := bufio.NewScanner(reader)
	next := func() string {
		t.Helper()
		line := make(chan string, 1)
		go func() { lines.Scan(); line <- lines.Text() }()
		select {
		case value := <-line:
			return value
		case <-time.After(3 * time.Second):
			t.Fatal("missing watch output")
		}
		return ""
	}
	if initial := next(); initial != "light" {
		t.Fatalf("initial: %s", initial)
	}
	emit := func(ns, name string, value uint32) {
		t.Helper()
		if err := server.Emit(portalPath, settings+".SettingChanged", ns, name, dbus.MakeVariant(value)); err != nil {
			t.Fatal(err)
		}
	}
	// Unrelated settings and repeated values must not produce output.
	emit("org.gnome.desktop.interface", "gtk-theme", 1)
	emit(namespace, key, 2)
	emit(namespace, key, 1)
	if changed := next(); changed != "dark" {
		t.Fatalf("signal: %s", changed)
	}
	emit(namespace, key, 0)
	if changed := next(); changed != "light" {
		t.Fatalf("no preference: %s", changed)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestLegacyPortalRead(t *testing.T) {
	address := privateBus(t)
	servePortal(t, address, legacyPortal{value: 1})
	if mode, err := (observer{conn: connect(t, address)}).read(context.Background()); err != nil || mode != "dark" {
		t.Fatalf("legacy read: %s, %v", mode, err)
	}
}

func TestGSettingsFallbackWithoutPortal(t *testing.T) {
	address := privateBus(t)
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "gsettings"), []byte("#!/bin/sh\necho \"'prefer-dark'\"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	if mode, err := (observer{conn: connect(t, address)}).read(context.Background()); err != nil || mode != "dark" {
		t.Fatalf("fallback: %s, %v", mode, err)
	}
	if mode, err := (observer{}).read(context.Background()); err != nil || mode != "dark" {
		t.Fatalf("fallback without a session bus: %s, %v", mode, err)
	}
	t.Setenv("PATH", t.TempDir())
	if _, err := (observer{conn: connect(t, address)}).read(context.Background()); err == nil {
		t.Fatal("missing portal and gsettings reported a mode")
	}
}
