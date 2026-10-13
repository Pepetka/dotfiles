package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

const Unit = "theme-sync.service"

// The user manager rarely inherits a custom XDG_CONFIG_HOME from the shell, so
// the unit lives in the default search path rather than the theme config root.
type Systemd struct{ Home string }

func (Systemd) Name() string   { return "systemd" }
func (s Systemd) Path() string { return filepath.Join(s.Home, ".config/systemd/user", Unit) }

// Escape systemd specifiers; ExecStart arguments are additionally quoted and
// protected from environment-variable expansion.
func unitValue(s string) string { return strings.ReplaceAll(s, "%", "%%") }
func unitArgument(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "$", "$$").Replace(s)
	return `"` + unitValue(s) + `"`
}

func (Systemd) Definition(directory string) []byte {
	return []byte(fmt.Sprintf(`[Unit]
Description=Theme synchronization
PartOf=graphical-session.target
After=graphical-session.target

[Service]
ExecStart=%s watch --theme-dir %s
Restart=always
RestartSec=5
StandardError=append:%s

[Install]
WantedBy=graphical-session.target
`, unitArgument(filepath.Join(directory, "theme-sync")), unitArgument(directory), unitValue(filepath.Join(directory, "service.log"))))
}

func systemctl(args ...string) (string, error) {
	return run("systemctl", append([]string{"--user"}, args...)...)
}

func (Systemd) Validate(string) error {
	if _, err := systemctl("show-environment"); err != nil {
		return errors.Join(errors.New("the systemd user manager is unavailable; use: theme install --manual"), err)
	}
	return nil
}

func (Systemd) Start() error {
	for _, args := range [][]string{{"daemon-reload"}, {"enable", Unit}, {"restart", Unit}} {
		if _, err := systemctl(args...); err != nil {
			return err
		}
	}
	return nil
}

// Stop also forgets a definition that the installer is about to remove.
func (Systemd) Stop() {
	_, _ = systemctl("disable", "--now", Unit)
	_, _ = systemctl("daemon-reload")
}

func (Systemd) Loaded() bool { _, err := systemctl("is-enabled", "--quiet", Unit); return err == nil }

func (Systemd) State() string {
	output, _ := systemctl("is-active", Unit)
	if output == "" {
		return "not loaded"
	}
	return strings.SplitN(output, "\n", 2)[0]
}
