package engine

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Replace a symlink itself, even if its target already contains the desired
// bytes. Generated files must never write through links into palette sources.
func AtomicWrite(path string, data []byte) (bool, error) {
	info, err := os.Lstat(path)
	if err == nil && info.Mode().IsRegular() {
		old, err := os.ReadFile(path)
		if err == nil && bytes.Equal(old, data) {
			return false, nil
		}
	}
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return false, err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return false, err
	}
	if closeErr != nil {
		return false, closeErr
	}
	if err = os.Rename(file.Name(), path); err != nil {
		return false, err
	}
	return true, nil
}

func lock(ctx context.Context, directory string) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(directory, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			unlock := func() { _ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN); _ = file.Close() }
			if err := ctx.Err(); err != nil {
				unlock()
				return nil, err
			}
			return unlock, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			_ = file.Close()
			return nil, err
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}
