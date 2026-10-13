package main

import (
	"fmt"
	"strings"
)

// The portal's org.freedesktop.appearance color-scheme: 1 prefers dark,
// 2 prefers light and 0 means no preference, which desktops render light.
func portalMode(value uint32) string {
	if value == 1 {
		return "dark"
	}
	return "light"
}

// GNOME's org.gnome.desktop.interface color-scheme, as printed by gsettings.
func gsettingsMode(output string) (string, error) {
	switch strings.Trim(strings.TrimSpace(output), "'") {
	case "prefer-dark":
		return "dark", nil
	case "default", "prefer-light":
		return "light", nil
	}
	return "", fmt.Errorf("unexpected GNOME color-scheme %q", strings.TrimSpace(output))
}
