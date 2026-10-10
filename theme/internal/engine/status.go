package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Status reports current persisted policy/mode even when reconciliation failed
// before it could update the cached adapter results. It never repairs state.
func CurrentStatus(ctx context.Context, directory string) (Status, error) {
	unlock, err := lock(ctx, directory)
	if err != nil {
		return Status{}, err
	}
	defer unlock()
	status := ReadStatus(directory)
	var failures []error
	data, err := os.ReadFile(filepath.Join(directory, "status.json"))
	if err == nil {
		err = json.Unmarshal(data, &status)
	}
	if err != nil {
		failures = append(failures, fmt.Errorf("status.json: %w", err))
	}
	data, err = os.ReadFile(filepath.Join(directory, "preference"))
	if errors.Is(err, os.ErrNotExist) {
		status.Preference = Auto
	} else if err == nil {
		status.Preference = Mode(strings.TrimSpace(string(data)))
		_, err = ParseMode(string(data), true)
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		failures = append(failures, fmt.Errorf("preference: %w", err))
	}
	data, err = os.ReadFile(filepath.Join(directory, "mode"))
	if err == nil {
		var mode Mode
		mode, err = ParseMode(string(data), false)
		if err == nil {
			status.Mode = mode
		}
	}
	if err != nil {
		failures = append(failures, fmt.Errorf("mode: %w", err))
	}
	status.StateError = ""
	err = errors.Join(failures...)
	if err != nil {
		status.StateError = err.Error()
	}
	return status, err
}
