package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fakeBackend struct {
	mode Mode
	err  error
}

func (b fakeBackend) Read(context.Context) (Mode, error) { return b.mode, b.err }
func (b fakeBackend) Watch(ctx context.Context, events chan<- System) {
	sendSystem(ctx, events, System{Mode: b.mode, Err: b.err})
	<-ctx.Done()
}

type fakeAdapter struct {
	name        string
	renderErr   error
	applyErr    error
	applyCount  int
	reloadCount int
}

func (a *fakeAdapter) Name() string { return a.name }
func (a *fakeAdapter) Render(_ Config, r RenderRequest) ([]Artifact, error) {
	return []Artifact{{Name: a.name + ".conf", Data: []byte(r.Mode)}}, a.renderErr
}

func (a *fakeAdapter) Apply(_ context.Context, _ Config, r ApplyRequest) (AdapterState, error) {
	a.applyCount++
	if r.Changed || r.Force || r.Previous.Mode != r.Mode || r.Previous.State == "error" {
		a.reloadCount++
	}
	if a.applyErr != nil {
		return AdapterState{}, a.applyErr
	}
	return AdapterState{State: "applied", Mode: r.Mode}, nil
}

func newEngine(t *testing.T) *Engine {
	t.Helper()
	return &Engine{Config: Config{Directory: filepath.Join(t.TempDir(), "custom config", "theme")}, Backend: fakeBackend{mode: Light}}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestPolicyPersistenceAndBackendFailure(t *testing.T) {
	e := newEngine(t)
	ctx := context.Background()
	status, err := e.Reconcile(ctx, "", nil, false)
	if err != nil || status.Preference != Auto || status.Mode != Light {
		t.Fatalf("initial: %+v, %v", status, err)
	}
	status, err = e.Reconcile(ctx, Dark, nil, false)
	if err != nil || status.Mode != Dark {
		t.Fatalf("manual: %+v, %v", status, err)
	}
	restarted := &Engine{Config: e.Config, Backend: fakeBackend{mode: Light}}
	status, err = restarted.Reconcile(ctx, "", &System{Mode: Light}, false)
	if err != nil || status.Preference != Dark || status.Mode != Dark {
		t.Fatalf("restart/system event overwrote manual: %+v, %v", status, err)
	}
	status, err = restarted.Reconcile(ctx, Auto, nil, false)
	if err != nil || status.Mode != Light {
		t.Fatalf("return to auto: %+v, %v", status, err)
	}
	restarted.Backend = fakeBackend{err: errors.New("unavailable")}
	status, err = restarted.Reconcile(ctx, "", nil, false)
	if err == nil || status.Mode != Light || status.BackendError == "" {
		t.Fatalf("failed backend did not retain mode: %+v, %v", status, err)
	}
	status, err = restarted.Reconcile(ctx, Dark, nil, false)
	if err != nil || status.Mode != Dark {
		t.Fatalf("manual depends on backend: %+v, %v", status, err)
	}
	if read(t, filepath.Join(e.Config.Directory, "preference")) != "dark\n" {
		t.Fatal("preference not persisted")
	}
}

func TestColdFailureInvalidPreferenceAndInvalidArguments(t *testing.T) {
	e := newEngine(t)
	e.Backend = fakeBackend{err: errors.New("broken")}
	status, err := e.Reconcile(context.Background(), "", nil, false)
	if err == nil || status.Mode != Dark {
		t.Fatalf("cold fallback: %+v, %v", status, err)
	}
	before := read(t, filepath.Join(e.Config.Directory, "preference"))
	if _, err := e.Reconcile(context.Background(), Mode("invalid"), nil, false); err == nil {
		t.Fatal("invalid preference accepted")
	}
	if read(t, filepath.Join(e.Config.Directory, "preference")) != before {
		t.Fatal("invalid request changed preference")
	}
	if err := os.WriteFile(filepath.Join(e.Config.Directory, "preference"), []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Reconcile(context.Background(), "", nil, false); err == nil {
		t.Fatal("corrupt preference silently reset")
	}
	if read(t, filepath.Join(e.Config.Directory, "mode")) != "dark\n" {
		t.Fatal("corrupt preference changed last mode")
	}
}

func TestCurrentStatusDetectsCorruptStateWithoutRepair(t *testing.T) {
	e := newEngine(t)
	ctx := context.Background()
	if _, err := e.Reconcile(ctx, "", nil, false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(e.Config.Directory, "preference"), []byte("broken\n"), 0600); err != nil {
		t.Fatal(err)
	}
	status, err := CurrentStatus(ctx, e.Config.Directory)
	if err == nil || status.Preference != "broken" || status.StateError == "" || status.Mode != Light {
		t.Fatalf("corrupt preference hidden: %+v, %v", status, err)
	}
	if read(t, filepath.Join(e.Config.Directory, "preference")) != "broken\n" {
		t.Fatal("status repaired state")
	}
	if _, err := e.Reconcile(ctx, Dark, nil, false); err != nil {
		t.Fatal(err)
	}
	status, err = CurrentStatus(ctx, e.Config.Directory)
	if err != nil || status.StateError != "" || status.Preference != Dark || status.Mode != Dark {
		t.Fatalf("recovery: %+v, %v", status, err)
	}
	if err := os.WriteFile(filepath.Join(e.Config.Directory, "mode"), []byte("auto\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if status, err := CurrentStatus(ctx, e.Config.Directory); err == nil || status.StateError == "" {
		t.Fatalf("corrupt mode hidden: %+v, %v", status, err)
	}
}

func TestCanceledRequestDoesNotPublish(t *testing.T) {
	e := newEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.Reconcile(ctx, Dark, nil, false); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled request accepted: %v", err)
	}
	if _, err := os.Stat(e.Config.Directory); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canceled request created state: %v", err)
	}
}

func TestAtomicNoopAndSymlinkMigration(t *testing.T) {
	directory := t.TempDir()
	source := filepath.Join(directory, "source")
	path := filepath.Join(directory, "generated")
	if err := os.WriteFile(source, []byte("dark\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, path); err != nil {
		t.Fatal(err)
	}
	changed, err := AtomicWrite(path, []byte("dark\n"))
	if err != nil || !changed {
		t.Fatalf("symlink not replaced: %v, %v", changed, err)
	}
	info, _ := os.Lstat(path)
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
		t.Fatalf("incorrect generated file: %v", info.Mode())
	}
	before := info.ModTime()
	changed, err = AtomicWrite(path, []byte("dark\n"))
	if err != nil || changed {
		t.Fatalf("noop: %v, %v", changed, err)
	}
	info, _ = os.Stat(path)
	if !info.ModTime().Equal(before) {
		t.Fatal("noop changed modification time")
	}
	if _, err := AtomicWrite(path, []byte("light\n")); err != nil {
		t.Fatal(err)
	}
	if read(t, source) != "dark\n" {
		t.Fatal("source changed through symlink")
	}
}

func TestIndependentAdaptersAndReloadRecovery(t *testing.T) {
	e := newEngine(t)
	broken := &fakeAdapter{name: "broken", renderErr: errors.New("missing palette")}
	retry := &fakeAdapter{name: "retry", applyErr: errors.New("reload failed")}
	healthy := &fakeAdapter{name: "healthy"}
	e.Adapters = []Adapter{broken, retry, healthy}
	status, err := e.Reconcile(context.Background(), "", nil, false)
	if err == nil || status.Adapters["broken"].State != "error" || status.Adapters["retry"].State != "error" || status.Adapters["healthy"].Mode != Light {
		t.Fatalf("adapter isolation: %+v, %v", status, err)
	}
	if read(t, filepath.Join(e.Config.Directory, "mode")) != "light\n" {
		t.Fatal("adapter failure blocked mode publication")
	}
	retry.applyErr = nil
	broken.renderErr = nil
	if _, err := e.Reconcile(context.Background(), "", nil, false); err != nil {
		t.Fatal(err)
	}
	if retry.reloadCount != 2 || healthy.reloadCount != 1 {
		t.Fatalf("retry without healthy reload: retry=%d healthy=%d", retry.reloadCount, healthy.reloadCount)
	}
	before, _ := os.Stat(filepath.Join(e.Config.Directory, "status.json"))
	if _, err := e.Reconcile(context.Background(), "", nil, false); err != nil {
		t.Fatal(err)
	}
	after, _ := os.Stat(filepath.Join(e.Config.Directory, "status.json"))
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("stable status rewritten")
	}
	if _, err := e.Reconcile(context.Background(), "", nil, true); err != nil {
		t.Fatal(err)
	}
	if healthy.reloadCount != 2 {
		t.Fatal("sync did not force reload")
	}
}

func TestConcurrentDaemonAndCLIKeepLatestPreference(t *testing.T) {
	e := newEngine(t)
	ctx := context.Background()
	if _, err := e.Reconcile(ctx, Auto, nil, false); err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	for i := 0; i < 20; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			if _, err := e.Reconcile(ctx, "", &System{Mode: Light}, false); err != nil {
				t.Error(err)
			}
		}()
	}
	if _, err := e.Reconcile(ctx, Dark, nil, false); err != nil {
		t.Fatal(err)
	}
	workers.Wait()
	if read(t, filepath.Join(e.Config.Directory, "mode")) != "dark\n" {
		t.Fatal("daemon overwrote manual selection")
	}
	if _, err := e.Reconcile(ctx, Light, nil, false); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Reconcile(ctx, "", &System{Mode: Dark}, false); err != nil {
		t.Fatal(err)
	}
	if read(t, filepath.Join(e.Config.Directory, "mode")) != "light\n" {
		t.Fatal("later preference was lost")
	}
}

func TestWatchRecoversAndReadsExternalPreference(t *testing.T) {
	e := newEngine(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { e.Watch(ctx); close(done) }()
	defer func() {
		cancel()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("daemon did not stop")
		}
	}()
	wait := func(mode Mode) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			data, _ := os.ReadFile(filepath.Join(e.Config.Directory, "mode"))
			if string(data) == string(mode)+"\n" {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatalf("watch did not publish %s", mode)
	}
	wait(Light)
	if _, err := e.Reconcile(context.Background(), Dark, nil, false); err != nil {
		t.Fatal(err)
	}
	time.Sleep(1100 * time.Millisecond)
	wait(Dark)
	if _, err := e.Reconcile(context.Background(), Auto, nil, false); err != nil {
		t.Fatal(err)
	}
	wait(Light)
}

func TestManualModeWithoutBackend(t *testing.T) {
	e := newEngine(t)
	e.Backend = NewBackend(e.Config)
	if _, ok := e.Backend.(NoBackend); !ok {
		t.Fatalf("empty backend did not select manual mode: %T", e.Backend)
	}
	ctx := context.Background()
	for _, mode := range []Mode{Light, Dark} {
		status, err := e.Reconcile(ctx, mode, nil, false)
		if err != nil || status.Mode != mode || status.BackendError != "" {
			t.Fatalf("manual %s: %+v, %v", mode, status, err)
		}
	}
	status, err := e.Reconcile(ctx, Auto, nil, false)
	if !errors.Is(err, ErrNoBackend) || status.Mode != Dark {
		t.Fatalf("auto without backend did not retain the last mode: %+v, %v", status, err)
	}
	events := make(chan System, 1)
	NoBackend{}.Watch(ctx, events)
	if event := <-events; !errors.Is(event.Err, ErrNoBackend) {
		t.Fatalf("watch without backend: %+v", event)
	}
}

type countingAdapter struct {
	applies atomic.Int32
	fail    atomic.Bool
}

func (*countingAdapter) Name() string                                     { return "counting" }
func (*countingAdapter) Render(Config, RenderRequest) ([]Artifact, error) { return nil, nil }
func (a *countingAdapter) Apply(context.Context, Config, ApplyRequest) (AdapterState, error) {
	a.applies.Add(1)
	if a.fail.Load() {
		return AdapterState{}, errors.New("reload failed")
	}
	return AdapterState{State: "applied"}, nil
}

func TestWatchRetriesFailuresAndThrottlesHealthyChecks(t *testing.T) {
	retry, health := RetryInterval, HealthInterval
	RetryInterval, HealthInterval = 20*time.Millisecond, time.Hour
	defer func() { RetryInterval, HealthInterval = retry, health }()
	e := newEngine(t)
	adapter := &countingAdapter{}
	adapter.fail.Store(true)
	e.Adapters = []Adapter{adapter}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { e.Watch(ctx); close(done) }()
	defer func() { cancel(); <-done }()
	deadline := time.Now().Add(2 * time.Second)
	for adapter.applies.Load() < 5 {
		if time.Now().After(deadline) {
			t.Fatal("failed reload was not retried")
		}
		time.Sleep(10 * time.Millisecond)
	}
	adapter.fail.Store(false)
	time.Sleep(100 * time.Millisecond)
	settled := adapter.applies.Load()
	time.Sleep(200 * time.Millisecond)
	if adapter.applies.Load() != settled {
		t.Fatal("healthy daemon kept reapplying on every tick")
	}
}
