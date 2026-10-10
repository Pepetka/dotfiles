package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func script(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "backend with spaces")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExecutableBackendContract(t *testing.T) {
	b := ExecutableBackend{Path: script(t, `case "$1" in
read) echo diagnostic >&2; echo light ;;
watch) echo dark; echo light; sleep 30 ;;
esac`)}
	mode, err := b.Read(context.Background())
	if err != nil || mode != Light {
		t.Fatalf("read/stderr: %s, %v", mode, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan System, 8)
	done := make(chan struct{})
	go func() { b.Watch(ctx, events); close(done) }()
	for _, expected := range []Mode{Dark, Light} {
		select {
		case event := <-events:
			if event.Err != nil || event.Mode != expected {
				t.Fatalf("event: %+v", event)
			}
		case <-time.After(time.Second):
			t.Fatal("missing backend event")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watch leaked process")
	}
	bad := ExecutableBackend{Path: script(t, "echo auto")}
	if _, err := bad.Read(context.Background()); err == nil {
		t.Fatal("backend returned policy instead of mode")
	}
}

func TestBackendWatchCrashAndMalformedOutputRestart(t *testing.T) {
	b := ExecutableBackend{Path: script(t, "echo light; echo malformed")}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan System, 8)
	go b.Watch(ctx, events)
	for _, failure := range []bool{false, true, false} {
		select {
		case event := <-events:
			if (event.Err != nil) != failure {
				t.Fatalf("restart event: %+v", event)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("backend was not restarted")
		}
	}
}

func TestBackendInitialTimeout(t *testing.T) {
	b := ExecutableBackend{Path: script(t, "sleep 30")}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := make(chan System, 1)
	go b.Watch(ctx, events)
	select {
	case event := <-events:
		if event.Err == nil || !strings.Contains(event.Err.Error(), "timed out") {
			t.Fatalf("initial timeout: %+v", event)
		}
	case <-time.After(CommandTimeout + time.Second):
		t.Fatal("backend initial read hangs")
	}
}

func TestSubprocessOutputAndTimeout(t *testing.T) {
	output, err := Run(context.Background(), "/bin/sh", "-c", "yes diagnostic | head -n 20000")
	if err != nil || len(output) != 65536 {
		t.Fatalf("output not drained/bounded: %d, %v", len(output), err)
	}
	start := time.Now()
	_, err = Run(context.Background(), "/bin/sh", "-c", "sleep 30 & wait")
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > CommandTimeout+time.Second {
		t.Fatalf("process tree timeout: %v in %s", err, time.Since(start))
	}
}
