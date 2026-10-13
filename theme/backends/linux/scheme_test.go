package main

import "testing"

func TestColorSchemeMapping(t *testing.T) {
	for value, want := range map[uint32]string{0: "light", 1: "dark", 2: "light", 7: "light"} {
		if got := portalMode(value); got != want {
			t.Fatalf("portal %d: %s", value, got)
		}
	}
	for output, want := range map[string]string{"'prefer-dark'\n": "dark", "'default'\n": "light", "'prefer-light'": "light"} {
		if got, err := gsettingsMode(output); err != nil || got != want {
			t.Fatalf("gsettings %q: %s, %v", output, got, err)
		}
	}
	if _, err := gsettingsMode("'unknown'"); err == nil {
		t.Fatal("unknown GNOME value accepted")
	}
}
