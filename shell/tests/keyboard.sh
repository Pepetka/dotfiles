#!/usr/bin/env bash
# Keyboard layout indicator on each platform, with stubbed system tools only.
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
test_directory="$(mktemp -d "${TMPDIR:-/tmp}/keyboard-tests.XXXXXX")"
trap 'rm -rf "$test_directory"' EXIT
# Real desktop tools must never leak into a scenario: expose only text utilities.
tools="$test_directory/tools"
mkdir -p "$tools"
for tool in awk grep sed tr; do
  ln -s "$(command -v "$tool")" "$tools/$tool"
done

# layout OS [VAR=value...] -- runs keyboard.sh with the current stubs.
layout() {
  local system="$1"
  shift
  printf '#!/bin/sh\necho %s\n' "$system" >"$test_directory/bin/uname"
  chmod 700 "$test_directory/bin/"*
  env -i HOME="$test_directory" PATH="$test_directory/bin:$tools" "$@" "$(command -v bash)" "$source_directory/shell/keyboard.sh"
}
stub() {
  printf '#!/bin/sh\n%s\n' "$2" >"$test_directory/bin/$1"
}
expect() {
  local want="$1"
  shift
  local got
  got="$(layout "$@")"
  [[ "$got" == "$want" ]] || {
    echo "keyboard layout: expected '$want', got '$got' ($*)" >&2
    exit 1
  }
}
reset() {
  rm -rf "$test_directory/bin"
  mkdir -p "$test_directory/bin"
}

reset
expect '' Darwin
stub defaults "printf '%s\n' '\"KeyboardLayout Name\" = \"ABC\";'"
expect EN Darwin
stub defaults "printf '%s\n' '\"KeyboardLayout Name\" = \"Russian\";'"
expect RU Darwin

# Linux stays silent without a graphical session or supported desktop.
reset
stub gsettings "echo \"[('xkb', 'ru'), ('xkb', 'us')]\""
expect '' Linux
reset
expect '' Linux WAYLAND_DISPLAY=wayland-0

# GNOME: the active source is first in mru-sources; sources is the fallback.
stub gsettings "echo \"[('xkb', 'ru'), ('xkb', 'us')]\""
expect RU Linux WAYLAND_DISPLAY=wayland-0
stub gsettings "echo \"[('xkb', 'us'), ('xkb', 'ru')]\""
expect EN Linux WAYLAND_DISPLAY=wayland-0
stub gsettings "if [ \"\$3\" = mru-sources ]; then echo '@a(ss) []'; else echo \"[('xkb', 'de+nodeadkeys')]\"; fi"
expect DE Linux WAYLAND_DISPLAY=wayland-0

# KDE: active index into the layout list.
reset
stub gdbus "case \"\$*\" in *getLayoutsList*) echo \"([('us', '', 'English (US)'), ('ru', '', 'Russian')],)\" ;; *getLayout*) echo '(uint32 1,)' ;; esac"
expect RU Linux WAYLAND_DISPLAY=wayland-0 XDG_CURRENT_DESKTOP=KDE
expect '' Linux WAYLAND_DISPLAY=wayland-0

# X11: xkb-switch when installed, otherwise the xkb group from the LED mask.
reset
stub setxkbmap "printf 'rules:      evdev\nlayout:     us,ru\n'"
stub xset "echo '  auto repeat:  on    key click percent:  0    LED mask:  00001000'"
expect RU Linux DISPLAY=:0
stub xset "echo '  auto repeat:  on    key click percent:  0    LED mask:  00000000'"
expect EN Linux DISPLAY=:0
stub xkb-switch "echo ua"
expect UA Linux DISPLAY=:0

echo 'Keyboard layout passed: macOS, GNOME, KDE, X11 and silent sessions'
