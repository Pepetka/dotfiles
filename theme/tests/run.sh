#!/usr/bin/env bash
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
root_directory="$(dirname "$source_directory")"
test_directory="$(mktemp -d "${TMPDIR:-/tmp}/theme-tests.XXXXXX")"
trap 'rm -rf "$test_directory"' EXIT
export GOCACHE="${GOCACHE:-$test_directory/go-cache}"
(cd "$source_directory" && go test -race ./... && go vet ./... && GOOS=linux GOARCH=amd64 go build -o "$test_directory/theme-linux" ./cmd/theme && GOOS=linux GOARCH=amd64 go build -o "$test_directory/install-linux" ./cmd/theme-install)
for script in "$source_directory/theme" "$source_directory/install.sh" "$source_directory/tests/"*.sh "$root_directory/shell/theme-switch.sh" "$root_directory/shell/tests/theme-sync.sh"; do
  bash -n "$script"
done
for config in "$root_directory/configs/aliases.zsh" "$root_directory/configs/.zshrc" "$source_directory/resources/p10k/"*.zsh "$source_directory/completions/_theme"; do
  zsh -n "$config"
done
if DOTFILES="$root_directory" zsh -c 'source "$DOTFILES/configs/aliases.zsh"; theme invalid' >/dev/null 2>&1; then
  echo 'Invalid theme command incorrectly succeeded' >&2
  exit 1
fi
if rg -n '#[[:xdigit:]]{6}' "$source_directory/resources/p10k/p10k.zsh"; then
  echo 'ANSI prompt contains fixed RGB colors' >&2
  exit 1
fi
mkdir -p "$test_directory/config with spaces/theme"
if command -v lua >/dev/null; then
  lua "$source_directory/tests/wezterm.lua" "$source_directory" "$test_directory/config with spaces"
elif command -v nvim >/dev/null; then
  NVIM_LOG_FILE=/dev/null nvim --headless -u NONE -i NONE -n -l "$source_directory/tests/wezterm.lua" "$source_directory" "$test_directory/config with spaces"
fi
if [[ "$(uname -s)" == Darwin ]]; then
  xcrun swiftc -module-cache-path "$test_directory/swift-cache" -parse-as-library -O \
    "$source_directory/backends/macos/main.swift" -o "$test_directory/macos-backend"
  xcrun swiftc -module-cache-path "$test_directory/swift-cache" -parse-as-library -O \
    "$source_directory/platform/macos/main.swift" -o "$test_directory/theme-macos"
  mode="$("$test_directory/macos-backend" read)"
  [[ "$mode" == dark || "$mode" == light ]]
  plutil -lint "$source_directory/platform/macos/Info.plist" >/dev/null
  bash "$source_directory/tests/install.sh"
  bash "$source_directory/tests/applications.sh"
fi
echo 'Theme engine, adapters, shell and platform checks passed'
