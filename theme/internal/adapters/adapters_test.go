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
			artifacts, err := adapter.Render(c, engine.RenderRequest{Mode: mode, Preference: mode})
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

type ghosttyPolicyAdapter struct {
	Ghostty
	reloads int
}

func (a *ghosttyPolicyAdapter) Apply(_ context.Context, _ engine.Config, r engine.ApplyRequest) (engine.AdapterState, error) {
	return applyGhostty(r, "42", func() (string, error) {
		a.reloads++
		return "reload requested", nil
	})
}

func TestGhosttyNativeAutoAndManualOverrides(t *testing.T) {
	c := engine.Config{Root: sourceRoot(t), Directory: filepath.Join(t.TempDir(), "config with spaces")}
	a := &ghosttyPolicyAdapter{}
	e := engine.Engine{Config: c, Adapters: []engine.Adapter{a}}
	var autoFile os.FileInfo
	for _, step := range []struct {
		preference, system, applied engine.Mode
		force                       bool
		reloads                     int
	}{
		{engine.Auto, engine.Light, engine.Auto, false, 1},
		{"", engine.Dark, engine.Auto, false, 1},
		{engine.Light, engine.Dark, engine.Light, false, 2},
		{"", engine.Light, engine.Light, false, 2},
		{"", engine.Dark, engine.Light, false, 2},
		{engine.Dark, engine.Light, engine.Dark, false, 3},
		{"", engine.Dark, engine.Dark, false, 3},
		{engine.Auto, engine.Dark, engine.Auto, false, 4},
		{"", engine.Light, engine.Auto, false, 4},
		{"", engine.Light, engine.Auto, true, 5},
	} {
		status, err := e.Reconcile(context.Background(), step.preference, &engine.System{Mode: step.system}, step.force)
		if err != nil || a.reloads != step.reloads || status.Adapters["ghostty"].Mode != step.applied {
			t.Fatalf("%+v: reloads=%d, status=%+v, err=%v", step, a.reloads, status, err)
		}
		path := filepath.Join(c.Directory, "ghostty-theme.conf")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if step.applied == engine.Auto {
			if !strings.Contains(string(data), "light:TokyoNight Day") || !strings.Contains(string(data), "dark:TokyoNight Moon") {
				t.Fatalf("missing native themes: %s", data)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if autoFile != nil && !os.SameFile(autoFile, info) {
				t.Fatal("system change rewrote the native selection")
			}
			autoFile = info
		} else {
			autoFile = nil
			expected := `theme = "TokyoNight Moon"`
			if step.applied == engine.Light {
				expected = `theme = "TokyoNight Day"`
			}
			if !strings.Contains(string(data), expected) || strings.Contains(string(data), "cursor-color") {
				t.Fatalf("manual palette: %s", data)
			}
		}
	}
	for _, name := range []string{"ghostty-light.conf", "ghostty-dark.conf"} {
		if _, err := os.Stat(filepath.Join(c.Directory, name)); !os.IsNotExist(err) {
			t.Fatalf("unnecessary palette file %s: %v", name, err)
		}
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
		artifacts, err := adapter.Render(c, engine.RenderRequest{Mode: mode, Preference: mode})
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

func TestLinuxGhosttyReloadThroughSystemd(t *testing.T) {
	show, reloads := "", 0
	systemctl := func(args ...string) (string, error) {
		if args[0] == "reload" {
			reloads++
			return "", nil
		}
		if show == "" {
			return "", errors.New("systemctl unavailable")
		}
		return show, nil
	}
	for _, preference := range []engine.Mode{engine.Dark, engine.Auto} {
		r := engine.ApplyRequest{Mode: engine.Dark, Preference: preference}
		for _, unit := range []string{"", "ActiveState=inactive\nMainPID=0\nCanReload=yes\n", "ActiveState=active\nMainPID=42\nCanReload=no\n"} {
			show = unit
			state, err := applyLinuxGhostty(r, systemctl)
			if err != nil || state.State != "published" || state.Mode != preference || reloads != 0 {
				t.Fatalf("%s without a reloadable unit %q: %+v, reloads=%d, err=%v", preference, unit, state, reloads, err)
			}
		}
	}
	r := engine.ApplyRequest{Mode: engine.Dark, Preference: engine.Dark}
	show = "ActiveState=active\nMainPID=42\nCanReload=yes\n"
	state, err := applyLinuxGhostty(r, systemctl)
	if err != nil || state.State != "reload requested" || state.Identity != "42" || reloads != 1 {
		t.Fatalf("reload: %+v, %v", state, err)
	}
	r.Previous = state
	if _, err := applyLinuxGhostty(r, systemctl); err != nil || reloads != 1 {
		t.Fatal("healthy instance reloaded again")
	}
	show = "ActiveState=active\nMainPID=43\nCanReload=yes\n"
	state, err = applyLinuxGhostty(r, systemctl)
	if err != nil || reloads != 2 {
		t.Fatal("restarted instance not reloaded")
	}
	r.Previous = state
	for _, step := range []struct {
		mode, preference engine.Mode
		reloads          int
	}{
		{engine.Light, engine.Auto, 3},
		{engine.Dark, engine.Auto, 3},
		{engine.Light, engine.Auto, 3},
		{engine.Light, engine.Light, 4},
		{engine.Light, engine.Auto, 5},
		{engine.Dark, engine.Auto, 5},
		{engine.Dark, engine.Dark, 6},
		{engine.Dark, engine.Dark, 6},
	} {
		r.Mode, r.Preference = step.mode, step.preference
		state, err := applyLinuxGhostty(r, systemctl)
		if err != nil || state.State != "reload requested" || state.Mode != step.preference || state.Identity != "43" || reloads != step.reloads {
			t.Fatalf("%+v: state=%+v, reloads=%d, err=%v", step, state, reloads, err)
		}
		r.Previous = state
	}
}
