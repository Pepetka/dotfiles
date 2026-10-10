#!/usr/bin/env bash
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
[[ "$(uname -s)" == Darwin ]] || exit 0
test_directory="$(mktemp -d "${TMPDIR:-/tmp}/theme-install-tests.XXXXXX")"
test_directory="$(cd "$test_directory" && pwd -P)"
trap 'rm -rf "$test_directory"' EXIT
(cd "$source_directory" && go build -o "$test_directory/engine" ./cmd/theme)
(cd "$source_directory" && go build -o "$test_directory/installer" ./cmd/theme-install)
mkdir -p "$test_directory/bin" "$test_directory/home/Library/LaunchAgents" "$test_directory/config with spaces/theme"
export THEME_TEST_ROOT="$test_directory"
export THEME_TEST_SOURCE="$source_directory"
export HOME="$test_directory/home"
export XDG_CONFIG_HOME="$test_directory/config with spaces"
export DOTFILES="$(dirname "$source_directory")"
export PATH="$test_directory/bin:$PATH"
cat >"$test_directory/bin/go" <<'SH'
#!/bin/sh
if [ "$1" = run ]; then
  shift 2
  exec "$THEME_TEST_ROOT/installer" "$@"
fi
while [ "$#" -gt 0 ]; do
  if [ "$1" = -o ]; then cp "$THEME_TEST_ROOT/engine" "$2"; exit; fi
  shift
done
exit 1
SH
cat >"$test_directory/bin/xcrun" <<'SH'
#!/bin/sh
echo "$*" >> "$THEME_TEST_ROOT/swift.log"
while [ "$#" -gt 0 ]; do
  if [ "$1" = -o ]; then
    cat > "$2" <<'BACKEND'
#!/bin/sh
case "$1" in
read) echo light ;;
watch) echo light; sleep 30 ;;
esac
BACKEND
    chmod 700 "$2"
    exit
  fi
  shift
done
exit 1
SH
cat >"$test_directory/bin/codesign" <<'SH'
#!/bin/sh
exit 0
SH
cat >"$test_directory/bin/launchctl" <<'SH'
#!/bin/sh
echo "$*" >> "$THEME_TEST_ROOT/launchctl.log"
case "$1" in
print) test -f "$THEME_TEST_ROOT/loaded" ;;
bootout) rm -f "$THEME_TEST_ROOT/loaded" ;;
bootstrap)
  if test -f "$THEME_TEST_ROOT/fail-once"; then rm "$THEME_TEST_ROOT/fail-once"; exit 1; fi
  touch "$THEME_TEST_ROOT/loaded"
  ;;
esac
SH
chmod 700 "$test_directory/bin/"*
directory="$XDG_CONFIG_HOME/theme"
echo light >"$directory/preference"
echo light >"$directory/mode"
echo original >"$directory/alacritty-theme.toml"
echo original-prompt >"$HOME/.p10k.zsh"
echo legacy >"$directory/theme-sync"
chmod 700 "$directory/theme-sync"
echo legacy-plist >"$HOME/Library/LaunchAgents/com.vlvkuznetsov.theme-sync.plist"
touch "$test_directory/loaded" "$test_directory/fail-once"
if "$source_directory/theme" install; then
  echo 'Failed bootstrap incorrectly reported success' >&2
  exit 1
fi
[[ "$(cat "$directory/theme-sync")" == legacy ]]
[[ "$(cat "$directory/alacritty-theme.toml")" == original ]]
[[ "$(cat "$HOME/.p10k.zsh")" == original-prompt ]]
[[ -f "$test_directory/loaded" ]]
"$source_directory/theme" install
[[ "$(cat "$directory/preference")" == light ]]
[[ "$(cat "$directory/p10k-before-install.zsh")" == original-prompt ]]
[[ -f "$directory/wezterm-bridge.lua" ]]
helper_before="$(shasum -a 256 "$directory/Theme Automation.app/Contents/MacOS/theme-macos")"
compilations_before="$(rg -c 'platform/macos/main.swift' "$test_directory/swift.log")"
"$source_directory/theme" install
[[ "$(shasum -a 256 "$directory/Theme Automation.app/Contents/MacOS/theme-macos")" == "$helper_before" ]]
[[ "$(rg -c 'platform/macos/main.swift' "$test_directory/swift.log")" == "$compilations_before" ]]
[[ "$(cat "$directory/p10k-before-install.zsh")" == original-prompt ]]
# Invalid persisted policy must fail preflight without touching the running service.
launch_calls_before="$(wc -l <"$test_directory/launchctl.log")"
echo broken >"$directory/preference"
if "$source_directory/theme" install; then
  echo 'Invalid preference incorrectly passed preflight' >&2
  exit 1
fi
[[ "$(wc -l <"$test_directory/launchctl.log")" == "$launch_calls_before" ]]
[[ -f "$test_directory/loaded" ]]
echo light >"$directory/preference"
# Changing helper sources must invalidate the fingerprint, without editing
# any repository file or invoking a real service.
mkdir "$test_directory/alternate"
cp -R "$source_directory" "$test_directory/alternate/theme"
printf '\n// Test-only helper source revision\n' >>"$test_directory/alternate/theme/platform/macos/main.swift"
"$test_directory/alternate/theme/theme" install
[[ "$(rg -c 'platform/macos/main.swift' "$test_directory/swift.log")" == "$((compilations_before + 1))" ]]
[[ "$(plutil -extract directory raw "$directory/engine.json")" == "$directory" ]]
# Explicit backends must survive subsequent installations.
printf '#!/bin/sh\necho light\n' >"$test_directory/custom backend"
chmod 700 "$test_directory/custom backend"
(cd "$test_directory" && "$source_directory/theme" install --backend "custom backend")
"$source_directory/theme" install
[[ "$(plutil -extract backend raw "$directory/engine.json")" == "$test_directory/custom backend" ]]
# Never contact the user's real tmux socket from a test.
plutil -replace tmux -string '' "$directory/engine.json"
plutil -convert json "$directory/engine.json"
"$source_directory/theme" dark
[[ "$(cat "$directory/mode")" == dark ]]
"$source_directory/theme" auto
[[ "$(cat "$directory/mode")" == light ]]
"$source_directory/theme" uninstall
[[ "$(cat "$directory/preference")" == auto ]]
[[ "$(cat "$directory/mode")" == light ]]
[[ ! -f "$directory/theme-sync" ]]
[[ ! -f "$directory/.platform-source.sha256" ]]
echo 'Installer passed: preflight isolation, rollback, helper identity/rebuild, XDG/spaces, prompt backup, CLI and uninstall'
