package engine

import (
	"context"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Drain output continuously, retaining only a bounded diagnostic tail.
type tailBuffer struct {
	mu   sync.Mutex
	data []byte
}

func (b *tailBuffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	n := len(data)
	b.data = append(b.data, data...)
	if len(b.data) > 65536 {
		b.data = b.data[len(b.data)-65536:]
	}
	return n, nil
}

func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}

func process(ctx context.Context, executable string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 100 * time.Millisecond
	return cmd
}

func Run(ctx context.Context, executable string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, CommandTimeout)
	defer cancel()
	cmd := process(ctx, executable, args...)
	output := &tailBuffer{}
	cmd.Stdout, cmd.Stderr = output, output
	err := cmd.Run()
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return output.String(), err
}

func Output(ctx context.Context, executable string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(ctx, CommandTimeout)
	defer cancel()
	cmd := process(ctx, executable, args...)
	stdout, stderr := &tailBuffer{}, &tailBuffer{}
	cmd.Stdout, cmd.Stderr = stdout, stderr
	err := cmd.Run()
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return stdout.String(), stderr.String(), err
}
