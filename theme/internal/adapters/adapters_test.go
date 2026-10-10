package adapters

import (
	"context"
	"dotfiles/theme/internal/engine"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func sourceRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestAdapterOutputs(t *testing.T) {
	c := engine.Config{Root: sourceRoot(t)}
	for _, mode := range []engine.Mode{engine.Dark, engine.Light} {
		for _, adapter := range []engine.Adapter{Palette{"alacritty", "toml"}, Tmux{}, Ghostty{}, WezTerm{}} {
			artifacts, err := adapter.Render(c, mode)
			if err != nil || len(artifacts) != 1 || len(artifacts[0].Data) == 0 {
				t.Fatalf("%s %s: %v", adapter.Name(), mode, err)
			}
			if adapter.Name() == "alacritty" || adapter.Name() == "tmux" {
				expected := "#222436"
				if mode == engine.Light {
					expected = "#e1e2e7"
				}
				if !strings.Contains(string(artifacts[0].Data), expected) {
					t.Fatalf("%s wrong %s palette", adapter.Name(), mode)
				}
			}
			if adapter.Name() == "ghostty" && strings.Contains(string(artifacts[0].Data), "dark:") {
				t.Fatal("Ghostty independently selects the mode")
			}
		}
	}
}

func TestGhosttyReloadPolicy(t *testing.T) {
	calls := 0
	failure := false
	reload := func() (string, error) {
		calls++
		if failure {
			return "", errors.New("permission unavailable")
		}
		return "reload requested", nil
	}
	r := engine.ApplyRequest{Mode: engine.Dark}
	state, err := applyGhostty(r, "", reload)
	if err != nil || state.State != "not running" || calls != 0 {
		t.Fatalf("closed app launched: %+v %v", state, err)
	}
	state, err = applyGhostty(r, "123", reload)
	if err != nil || calls != 1 {
		t.Fatalf("initial reload: %+v %v", state, err)
	}
	r.Previous = state
	if _, err := applyGhostty(r, "123", reload); err != nil || calls != 1 {
		t.Fatal("healthy app reloaded")
	}
	r.Mode = engine.Light
	failure = true
	state, err = applyGhostty(r, "123", reload)
	if err == nil {
		t.Fatal("failed reload acknowledged")
	}
	state.State = "error"
	r.Previous = state
	failure = false
	state, err = applyGhostty(r, "123", reload)
	if err != nil || calls != 3 || state.Mode != engine.Light {
		t.Fatalf("reload not retried: %+v %v", state, err)
	}
	r.Previous = state
	if _, err := applyGhostty(r, "456", reload); err != nil || calls != 4 {
		t.Fatal("restart did not trigger reload")
	}
	r.Force = true
	if _, err := applyGhostty(r, "456", reload); err != nil || calls != 5 {
		t.Fatal("sync did not trigger reload")
	}
}

func TestGhosttyLateReplyCannotLeaveOrphan(t *testing.T) {
	c := engine.Config{Directory: t.TempDir()}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	var path string
	_, err := ghosttyRequest(ctx, c, "ghostty-reload", func(_ context.Context, reply string) error { path = reply; return nil })
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("request did not time out: %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"state":"reload requested"}`), 0600); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("late reply could recreate a file: %v", err)
	}
	files, err := os.ReadDir(c.Directory)
	if err != nil || len(files) != 0 {
		t.Fatalf("request files leaked: %v, %v", files, err)
	}
}

func TestGhosttyHelperReply(t *testing.T) {
	c := engine.Config{Directory: t.TempDir()}
	for _, body := range []string{`{"state":"reload requested"}`, `{"state":"permission required","error":"denied"}`} {
		state, err := ghosttyRequest(context.Background(), c, "ghostty-reload", func(_ context.Context, path string) error { return os.WriteFile(path, []byte(body), 0600) })
		if state == "" || ((err != nil) != strings.Contains(body, "denied")) {
			t.Fatalf("helper reply: %s, %v", state, err)
		}
	}
}

func TestTmuxIntegration(t *testing.T) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		t.Skip("tmux is not installed")
	}
	// Darwin sockets have a short path limit; keep the socket outside the
	// platform's deeply nested default test directory.
	socketDirectory, err := os.MkdirTemp("/tmp", "theme-tmux-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(socketDirectory)
	c := engine.Config{Root: sourceRoot(t), Directory: filepath.Join(t.TempDir(), "config with spaces", "theme"), Tmux: tmux, Socket: filepath.Join(socketDirectory, "socket")}
	ctx := context.Background()
	run := func(args ...string) string {
		t.Helper()
		output, err := engine.Run(ctx, tmux, append([]string{"-S", c.Socket}, args...)...)
		if err != nil {
			t.Fatalf("tmux %v: %v: %s", args, err, output)
		}
		return strings.TrimSpace(output)
	}
	start := func() { run("-f", "/dev/null", "new-session", "-d", "-s", "theme-test", "/bin/cat") }
	start()
	defer func() { _, _ = engine.Run(ctx, tmux, "-S", c.Socket, "kill-server") }()
	if err := os.MkdirAll(c.Directory, 0700); err != nil {
		t.Fatal(err)
	}
	adapter := Tmux{}
	var previous engine.AdapterState
	apply := func(mode engine.Mode) {
		t.Helper()
		artifacts, err := adapter.Render(c, mode)
		if err != nil {
			t.Fatal(err)
		}
		changed, err := engine.AtomicWrite(filepath.Join(c.Directory, artifacts[0].Name), artifacts[0].Data)
		if err != nil {
			t.Fatal(err)
		}
		previous, err = adapter.Apply(ctx, c, engine.ApplyRequest{Mode: mode, Previous: previous, Changed: changed})
		if err != nil || previous.State != "applied" {
			t.Fatalf("apply %s: %+v, %v", mode, previous, err)
		}
	}
	apply(engine.Dark)
	if run("show-options", "-gqv", "@bg") != "#222436" {
		t.Fatal("dark palette did not apply")
	}
	apply(engine.Light)
	if run("show-options", "-gqv", "@bg") != "#e1e2e7" {
		t.Fatal("light palette did not apply")
	}
	run("set-environment", "-g", "XDG_CONFIG_HOME", filepath.Dir(c.Directory))
	run("set-option", "-g", "@bg", "incorrect")
	run("source-file", "-q", "-F", "#{?XDG_CONFIG_HOME,#{XDG_CONFIG_HOME},#{HOME}/.config}/theme/tmux-theme.conf")
	if run("show-options", "-gqv", "@bg") != "#e1e2e7" {
		t.Fatal("tmux XDG startup import failed")
	}
	run("set-option", "-g", "@theme_mode", "dark")
	apply(engine.Light)
	if run("show-options", "-gqv", "@theme_mode") != "light" {
		t.Fatal("stale server marker not recovered")
	}
	run("kill-server")
	time.Sleep(100 * time.Millisecond)
	start()
	apply(engine.Light)
	if run("show-options", "-gqv", "@bg") != "#e1e2e7" {
		t.Fatal("new server not updated")
	}
}
