package service

import (
	"encoding/xml"
	"path/filepath"
	"strings"
	"testing"
)

func TestServicePlistEscapesPaths(t *testing.T) {
	directory := "/config with spaces/кириллица & <tag> \"quotes\"/theme"
	data := Launchd{}.Definition(directory)
	var parsed struct {
		Dict struct {
			Keys    []string `xml:"key"`
			Strings []string `xml:"string"`
			Array   struct {
				Strings []string `xml:"string"`
			} `xml:"array"`
			Integer int        `xml:"integer"`
			True    []struct{} `xml:"true"`
		} `xml:"dict"`
	}
	if err := xml.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	d := parsed.Dict
	if len(d.Keys) != 8 || len(d.True) != 2 || d.Integer != 5 {
		t.Fatalf("incorrect launch policy: %+v", d)
	}
	want := []string{filepath.Join(directory, "theme-sync"), "watch", "--theme-dir", directory}
	if strings.Join(d.Array.Strings, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("incorrect arguments: %v", d.Array.Strings)
	}
	if d.Strings[0] != Label || d.Strings[1] != "Aqua" || d.Strings[2] != "Background" || d.Strings[3] != filepath.Join(directory, "service.log") {
		t.Fatalf("incorrect service metadata: %v", d.Strings)
	}
}

func TestSystemdUnitEscapesPaths(t *testing.T) {
	directory := `/config with spaces/100% $HOME "q" \b/theme`
	unit := string(Systemd{Home: "/home/user"}.Definition(directory))
	want := `ExecStart="/config with spaces/100%% $$HOME \"q\" \\b/theme/theme-sync" watch --theme-dir "/config with spaces/100%% $$HOME \"q\" \\b/theme"`
	if !strings.Contains(unit, want+"\n") {
		t.Fatalf("incorrect ExecStart:\n%s", unit)
	}
	if !strings.Contains(unit, "StandardError=append:/config with spaces/100%% $HOME \"q\" \\b/theme/service.log\n") {
		t.Fatalf("incorrect log path:\n%s", unit)
	}
	for _, line := range []string{"PartOf=graphical-session.target", "WantedBy=graphical-session.target", "Restart=always"} {
		if !strings.Contains(unit, line+"\n") {
			t.Fatalf("missing %s:\n%s", line, unit)
		}
	}
	if path := (Systemd{Home: "/home/user"}).Path(); path != filepath.Join("/home/user", ".config/systemd/user", Unit) {
		t.Fatalf("unit outside the default user search path: %s", path)
	}
}
