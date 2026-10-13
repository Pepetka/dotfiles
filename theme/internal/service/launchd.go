package service

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"
)

type Launchd struct{ Home, Domain string }

func (Launchd) Name() string     { return "launchd" }
func (l Launchd) Path() string   { return filepath.Join(l.Home, "Library/LaunchAgents", Label+".plist") }
func (l Launchd) target() string { return l.Domain + "/" + Label }

func (Launchd) Definition(directory string) []byte {
	escape := func(s string) string { var b bytes.Buffer; _ = xml.EscapeText(&b, []byte(s)); return b.String() }
	var args strings.Builder
	for _, value := range []string{filepath.Join(directory, "theme-sync"), "watch", "--theme-dir", directory} {
		fmt.Fprintf(&args, "<string>%s</string>", escape(value))
	}
	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>%s</string>
<key>ProgramArguments</key><array>%s</array>
<key>RunAtLoad</key><true/><key>KeepAlive</key><true/>
<key>ThrottleInterval</key><integer>5</integer>
<key>LimitLoadToSessionType</key><string>Aqua</string>
<key>ProcessType</key><string>Background</string>
<key>StandardErrorPath</key><string>%s</string>
</dict></plist>
`, Label, args.String(), escape(filepath.Join(directory, "service.log"))))
}

func (Launchd) Validate(path string) error {
	_, err := run("plutil", "-lint", path)
	return err
}

func (l Launchd) Start() error {
	_, err := run("launchctl", "bootstrap", l.Domain, l.Path())
	return err
}

func (l Launchd) Stop()        { _, _ = run("launchctl", "bootout", l.target()) }
func (l Launchd) Loaded() bool { _, err := run("launchctl", "print", l.target()); return err == nil }

func (l Launchd) State() string {
	output, err := run("launchctl", "print", l.target())
	if err != nil {
		return "not loaded"
	}
	for _, line := range strings.Split(output, "\n") {
		if state, ok := strings.CutPrefix(strings.TrimSpace(line), "state = "); ok {
			return state
		}
	}
	return "loaded"
}
