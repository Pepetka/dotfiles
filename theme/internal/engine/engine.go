package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func ReadStatus(directory string) Status {
	var status Status
	data, _ := os.ReadFile(filepath.Join(directory, "status.json"))
	_ = json.Unmarshal(data, &status)
	if status.Adapters == nil {
		status.Adapters = make(map[string]AdapterState)
	}
	return status
}

// Reconcile reads preference under the publication lock. Daemon events never
// carry a preference snapshot that could overwrite a newer CLI selection.
func (e *Engine) Reconcile(ctx context.Context, preference Mode, system *System, force bool) (Status, error) {
	unlock, err := lock(ctx, e.Config.Directory)
	if err != nil {
		return Status{}, err
	}
	defer unlock()
	previous := ReadStatus(e.Config.Directory)
	var failures []error
	preferencePath := filepath.Join(e.Config.Directory, "preference")
	if preference != "" {
		if _, err := ParseMode(string(preference), true); err != nil {
			return Status{}, err
		}
		if _, err := AtomicWrite(preferencePath, []byte(string(preference)+"\n")); err != nil {
			return Status{}, err
		}
	} else {
		data, readErr := os.ReadFile(preferencePath)
		if errors.Is(readErr, os.ErrNotExist) {
			preference = Auto
			if _, err := AtomicWrite(preferencePath, []byte("auto\n")); err != nil {
				return Status{}, err
			}
		} else if readErr != nil {
			return Status{}, readErr
		} else {
			preference, err = ParseMode(string(data), true)
			if err != nil {
				return Status{}, err
			}
		}
	}
	status := Status{Preference: preference, System: previous.System, Adapters: make(map[string]AdapterState)}
	if preference == Auto {
		if system == nil {
			mode, readErr := e.Backend.Read(ctx)
			system = &System{Mode: mode, Err: readErr}
		}
		if system.Err == nil {
			if _, parseErr := ParseMode(string(system.Mode), false); parseErr != nil {
				system = &System{Err: parseErr}
			}
		}
		if system.Err != nil {
			status.BackendError = system.Err.Error()
			failures = append(failures, system.Err)
			data, _ := os.ReadFile(filepath.Join(e.Config.Directory, "mode"))
			status.Mode, _ = ParseMode(string(data), false)
			if status.Mode == "" {
				status.Mode = Dark
			}
		} else {
			status.Mode, status.System = system.Mode, system.Mode
		}
	} else {
		status.Mode = preference
		if system != nil {
			if system.Mode == Dark || system.Mode == Light {
				status.System = system.Mode
			}
			if system.Err != nil {
				status.BackendError = system.Err.Error()
			}
		}
	}
	if _, err := AtomicWrite(filepath.Join(e.Config.Directory, "mode"), []byte(string(status.Mode)+"\n")); err != nil {
		return status, err
	}

	// Publish every healthy adapter before performing external reloads.
	type pending struct {
		adapter Adapter
		changed bool
	}
	var ready []pending
	for _, adapter := range e.Adapters {
		artifacts, renderErr := adapter.Render(e.Config, RenderRequest{Mode: status.Mode, Preference: preference})
		changed := false
		if renderErr == nil {
			for _, artifact := range artifacts {
				if artifact.Name == "." || filepath.Base(artifact.Name) != artifact.Name {
					renderErr = fmt.Errorf("invalid artifact name %q", artifact.Name)
					break
				}
				updated, writeErr := AtomicWrite(filepath.Join(e.Config.Directory, artifact.Name), artifact.Data)
				changed = changed || updated
				if writeErr != nil {
					renderErr = writeErr
					break
				}
			}
		}
		if renderErr != nil {
			status.Adapters[adapter.Name()] = AdapterState{State: "error", Error: renderErr.Error()}
			failures = append(failures, fmt.Errorf("%s: %w", adapter.Name(), renderErr))
		} else {
			ready = append(ready, pending{adapter, changed})
		}
	}
	for _, item := range ready {
		state, applyErr := item.adapter.Apply(ctx, e.Config, ApplyRequest{Mode: status.Mode, Preference: preference, Previous: previous.Adapters[item.adapter.Name()], Changed: item.changed, Force: force})
		if applyErr != nil {
			state.State, state.Error = "error", applyErr.Error()
			failures = append(failures, fmt.Errorf("%s: %w", item.adapter.Name(), applyErr))
		}
		status.Adapters[item.adapter.Name()] = state
	}
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return status, err
	}
	if _, err := AtomicWrite(filepath.Join(e.Config.Directory, "status.json"), append(data, '\n')); err != nil {
		return status, err
	}
	return status, errors.Join(failures...)
}

func (e *Engine) Watch(ctx context.Context) {
	events := make(chan System, 16)
	go e.Backend.Watch(ctx, events)
	mode, err := e.Backend.Read(ctx)
	system := System{Mode: mode, Err: err}
	ticker := time.NewTicker(RetryInterval)
	defer ticker.Stop()
	lastError := ""
	var lastUpdate time.Time
	update := func() {
		lastUpdate = time.Now()
		_, err := e.Reconcile(ctx, "", &system, false)
		message := ""
		if err != nil {
			message = err.Error()
		}
		if message != lastError {
			if message == "" {
				log.Print("Synchronization recovered")
			} else {
				log.Print(message)
			}
			lastError = message
		}
	}
	update()
	for {
		select {
		case <-ctx.Done():
			return
		case system = <-events:
			// Coalesce a burst, keeping the latest observation.
		drain:
			for {
				select {
				case system = <-events:
				default:
					break drain
				}
			}
			update()
		case <-ticker.C:
			// Retry failures promptly; otherwise only a periodic health check.
			if lastError != "" || time.Since(lastUpdate) >= HealthInterval {
				update()
			}
		}
	}
}
