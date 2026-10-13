package engine

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

type Backend interface {
	Read(context.Context) (Mode, error)
	Watch(context.Context, chan<- System)
}

var ErrNoBackend = errors.New("automatic mode requires a system backend; run: theme install")

// NoBackend serves manual installations.
type NoBackend struct{}

func (NoBackend) Read(context.Context) (Mode, error) { return "", ErrNoBackend }
func (NoBackend) Watch(ctx context.Context, events chan<- System) {
	sendSystem(ctx, events, System{Err: ErrNoBackend})
}

func NewBackend(c Config) Backend {
	if c.Manual() {
		return NoBackend{}
	}
	return ExecutableBackend{Path: c.Backend}
}

type ExecutableBackend struct{ Path string }

func (b ExecutableBackend) Read(ctx context.Context) (Mode, error) {
	output, diagnostics, err := Output(ctx, b.Path, "read")
	if err != nil {
		return "", fmt.Errorf("backend read: %w: %s", err, diagnostics)
	}
	return ParseMode(output, false)
}

func sendSystem(ctx context.Context, events chan<- System, event System) bool {
	select {
	case events <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// A backend must publish its initial state promptly. Once initialized, silence
// is valid: there may be no appearance changes for days.
func (b ExecutableBackend) watchOnce(ctx context.Context, events chan<- System) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := process(ctx, b.Path, "watch")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	diagnostics := &tailBuffer{}
	cmd.Stderr = diagnostics
	if err := cmd.Start(); err != nil {
		return err
	}
	lines := make(chan string)
	scanErr := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			select {
			case lines <- scanner.Text():
			case <-ctx.Done():
				scanErr <- ctx.Err()
				return
			}
		}
		scanErr <- scanner.Err()
		close(lines)
	}()
	timer := time.NewTimer(CommandTimeout)
	defer timer.Stop()
	var watchErr error
loop:
	for {
		select {
		case <-ctx.Done():
			watchErr = ctx.Err()
			break loop
		case <-timer.C:
			watchErr = fmt.Errorf("backend watch: initial state timed out")
			break loop
		case line, ok := <-lines:
			if !ok {
				watchErr = <-scanErr
				if watchErr == nil {
					watchErr = io.EOF
				}
				break loop
			}
			mode, err := ParseMode(line, false)
			if err != nil {
				watchErr = err
				break loop
			}
			timer.Stop()
			if !sendSystem(ctx, events, System{Mode: mode}) {
				watchErr = ctx.Err()
				break loop
			}
		}
	}
	cancel()
	_ = stdout.Close()
	_ = cmd.Wait()
	return fmt.Errorf("backend watch: %w: %s", watchErr, diagnostics.String())
}

func (b ExecutableBackend) Watch(ctx context.Context, events chan<- System) {
	for ctx.Err() == nil {
		err := b.watchOnce(ctx, events)
		if ctx.Err() != nil {
			return
		}
		if !sendSystem(ctx, events, System{Err: err}) {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}
