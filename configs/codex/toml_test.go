package main

import (
	"bytes"
	"strings"
	"testing"
)

func parsed(t *testing.T, text string) map[string]any {
	t.Helper()
	m, err := decode([]byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMerge(t *testing.T) {
	for _, tc := range []struct {
		name, local, source, old, want string
		conflict                       bool
	}{
		{"bootstrap", "# Keep\nmodel = 'local' # choice\n[tui]\ntheme = 'Tokyo Night Day'\n", "model = 'shared'\n[tui]\ntheme = 'Tokyo Night Moon'\nvim_mode_default = true", "", "model = 'local'\n[tui]\ntheme = 'Tokyo Night Day'\nvim_mode_default = true", false},
		{"independent edits", "model = 'old'\n[tui]\nstatus_line = ['model']\nvim_mode_default = true", "model = 'new'\n[tui]\nstatus_line = ['context-used']\nvim_mode_default = false", "model = 'old'\n[tui]\nstatus_line = ['model']\nvim_mode_default = false", "model = 'new'\n[tui]\nstatus_line = ['context-used']\nvim_mode_default = true", false},
		{"conflict", "model = 'local'", "model = 'new'", "model = 'old'", "model = 'local'", true},
		{"deletion", "[features]\nmemories = true # keep comment\nlocal_only = true", "", "[features]\nmemories = true", "[features]\nlocal_only = true", false},
		{"local deletion", "", "model = 'new'", "model = 'old'", "", true},
		{"inline table", "tui = { vim_mode_default = false, theme = 'Day' } # keep", "[tui]\nvim_mode_default = true\nnew_key = 'value'", "[tui]\nvim_mode_default = false", "tui = { vim_mode_default = true, theme = 'Day', new_key = 'value' }", false},
		{"quoted key", "[\"mcp_servers\".'with.dot']\nargs = [\n 'old', # item\n]\n", "[mcp_servers.'with.dot']\nargs = ['new']", "[mcp_servers.'with.dot']\nargs = ['old']", "[mcp_servers.'with.dot']\nargs = ['new']", false},
		{"multiline string", "model = '''old\ntext''' # keep\nlocal = 'untouched'\n", "model = 'new'", "model = '''old\ntext'''", "model = 'new'\nlocal = 'untouched'", false},
		{"array table retained", "[[local]]\nname = 'one'\n[[local]]\nname = 'two'\n", "model = 'new'", "", "model = 'new'\n[[local]]\nname = 'one'\n[[local]]\nname = 'two'", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var old map[string]any
			if tc.old != "" {
				old = parsed(t, tc.old)
			} else if tc.name == "deletion" {
				old = map[string]any{}
			}
			result, conflicts, err := merge([]byte(tc.local), parsed(t, tc.source), old, "config.toml")
			if err != nil {
				t.Fatal(err)
			}
			if (len(conflicts) > 0) != tc.conflict {
				t.Fatalf("conflicts: %v", conflicts)
			}
			if !equal(parsed(t, string(result)), parsed(t, tc.want)) {
				t.Fatalf("unexpected result:\n%s", result)
			}
			if strings.Contains(tc.local, "# keep") && !bytes.Contains(result, []byte("# keep")) {
				t.Fatal("comment lost")
			}
		})
	}
}

func TestEditPreservesUnchangedBytes(t *testing.T) {
	before := []byte("# intro\r\n[\"tui\"] # header\r\nstatus_line  = [ 'old' ] # trailing\r\nunknown = '''\n[theme]\nleave this alone\n'''\r\n")
	result, err := editTOML(before, []string{"tui", "status_line"}, []string{"new"}, true)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != strings.Replace(string(before), "[ 'old' ]", "['new']", 1) {
		t.Fatalf("formatting changed:\n%s", result)
	}
	added, err := editTOML(result, []string{"tui", "vim_mode_default"}, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(added, []byte("# header\r\n\"vim_mode_default\" = true\r\n")) {
		t.Fatalf("insertion: %s", added)
	}
}

func TestRemovalDropsEmptiedLine(t *testing.T) {
	before := []byte("model = 'old'\n\neffort = 'medium'\nkeep = 1 # note\nlocal = 2\n")
	result, err := editTOML(before, []string{"model"}, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "\neffort = 'medium'\nkeep = 1 # note\nlocal = 2\n" {
		t.Fatalf("unexpected removal:\n%q", result)
	}
	result, err = editTOML(result, []string{"keep"}, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "\neffort = 'medium'\n # note\nlocal = 2\n" {
		t.Fatalf("comment not retained:\n%q", result)
	}
}
