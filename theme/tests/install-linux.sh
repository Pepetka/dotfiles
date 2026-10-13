#!/usr/bin/env bash
# Linux installer acceptance with stubbed systemctl and Go; run by tests/linux.sh.
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[[ "$(uname -s)" == Linux ]] || exit 0
binaries="${THEME_TEST_BINARIES:?prebuilt Linux binaries directory}"
test_directory="$(mktemp -d "${TMPDIR:-/tmp}/theme-install-linux.XXXXXX")"
trap 'rm -rf "$test_directory"' EXIT
mkdir -p "$test_directory/bin" "$test_directory/home" "$test_directory/config with spaces/theme"
export THEME_TEST_ROOT="$test_directory" THEME_TEST_BINARIES="$binaries"
export HOME="$test_directory/home"
export XDG_CONFIG_HOME="$test_directory/config with spaces"
export DOTFILES="$(dirname "$source_directory")"
export PATH="$test_directory/bin:$PATH"
export TMUX_TMPDIR="$test_directory/tmux"
cat >"$test_directory/bin/go" <<'SH'
#!/bin/sh
if [ "$1" = run ]; then
  shift 2
  exec "$THEME_TEST_BINARIES/installer" "$@"
fi
target=""
while [ "$#" -gt 0 ]; do
  case "$1" in
  -o) target="$2"; shift ;;
  ./cmd/theme) cp "$THEME_TEST_BINARIES/engine" "$target"; exit ;;
  ./backends/linux) cp "$THEME_TEST_BINARIES/linux-backend" "$target"; exit ;;
  esac
  shift
done
exit 1
SH
cat >"$test_directory/bin/systemctl" <<'SH'
#!/bin/sh
shift
echo "$*" >> "$THEME_TEST_ROOT/systemctl.log"
case "$1" in
show-environment) test ! -f "$THEME_TEST_ROOT/no-manager" ;;
is-enabled) test -f "$THEME_TEST_ROOT/enabled" ;;
is-active) if test -f "$THEME_TEST_ROOT/active"; then echo active; else echo inactive; exit 3; fi ;;
enable) touch "$THEME_TEST_ROOT/enabled" ;;
disable) rm -f "$THEME_TEST_ROOT/enabled" "$THEME_TEST_ROOT/active" ;;
restart)
  if test -f "$THEME_TEST_ROOT/fail-once"; then rm "$THEME_TEST_ROOT/fail-once"; exit 1; fi
  touch "$THEME_TEST_ROOT/active"
  ;;
show) printf 'ActiveState=inactive\nMainPID=0\nCanReload=no\n' ;;
esac
SH
# The portal is absent in the container; the backend falls back to GNOME settings.
cat >"$test_directory/bin/gsettings" <<'SH'
#!/bin/sh
echo "'prefer-dark'"
SH
chmod 700 "$test_directory/bin/"*
directory="$XDG_CONFIG_HOME/theme"
unit="$HOME/.config/systemd/user/theme-sync.service"
theme="$source_directory/theme"
fail() { echo "$*" >&2; exit 1; }

"$theme" install
[[ -f "$unit" && -f "$test_directory/enabled" && -f "$test_directory/active" ]]
[[ -x "$directory/linux-backend" && ! -e "$directory/macos-backend" && ! -e "$directory/Theme Automation.app" ]]
grep -q '"service": "systemd"' "$directory/engine.json"
grep -qF "ExecStart=\"$directory/theme-sync\" watch --theme-dir \"$directory\"" "$unit"
[[ "$("$directory/linux-backend" read 2>/dev/null)" == dark ]]
"$theme" auto
[[ "$(cat "$directory/mode")" == dark ]]
"$theme" light
[[ "$(cat "$directory/mode")" == light && -f "$directory/alacritty-theme.toml" ]]
"$theme" status | grep -q '"service": "systemd: active"'

# A failed start restores the previous unit and service state.
echo previous-unit >>"$unit"
touch "$test_directory/fail-once"
if "$theme" install; then
  fail 'Failed service start incorrectly reported success'
fi
[[ "$(tail -n 1 "$unit")" == previous-unit && -f "$test_directory/active" ]]
"$theme" install
[[ "$(tail -n 1 "$unit")" != previous-unit ]]

# Without a user manager the service install fails early and suggests manual mode.
touch "$test_directory/no-manager"
if "$theme" install 2>"$test_directory/error.log"; then
  fail 'Missing systemd user manager incorrectly accepted'
fi
grep -q 'theme install --manual' "$test_directory/error.log"
[[ -f "$test_directory/active" ]]

# Manual mode: no unit or backend, auto becomes the effective mode and is then rejected.
"$theme" auto
"$theme" install --manual
[[ ! -e "$unit" && ! -f "$test_directory/enabled" && ! -e "$directory/linux-backend" ]]
[[ "$(cat "$directory/preference")" == dark ]]
if "$theme" auto; then
  fail 'Manual mode incorrectly accepted auto'
fi
"$theme" light
[[ "$(cat "$directory/mode")" == light ]]
"$theme" status | grep -q '"backend": "manual"'

rm "$test_directory/no-manager"
"$theme" install
[[ -f "$unit" && -f "$test_directory/active" ]]
"$theme" uninstall
[[ ! -e "$unit" && ! -f "$test_directory/enabled" && ! -e "$directory/theme-sync" ]]
[[ "$(cat "$directory/mode")" == light ]]
echo 'Linux installer passed: systemd unit, portal fallback, rollback, missing manager, manual mode and uninstall'
