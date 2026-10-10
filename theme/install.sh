#!/usr/bin/env bash
set -euo pipefail

source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dotfiles_directory="${DOTFILES:-$(dirname "$source_directory")}"
theme_directory="${XDG_CONFIG_HOME:-$HOME/.config}/theme"
label="com.vlvkuznetsov.theme-sync"
domain="gui/$(id -u)"
plist="$HOME/Library/LaunchAgents/$label.plist"

[[ "$(uname -s)" == Darwin ]] || { echo 'This service installer requires macOS.' >&2; exit 1; }

if [[ "${1:-}" == uninstall ]]; then
  [[ $# == 1 ]] || { echo 'Usage: theme uninstall' >&2; exit 1; }
  launchctl bootout "$domain/$label" 2>/dev/null || true
  rm -f "$plist" "$theme_directory/theme-sync" "$theme_directory/macos-backend" "$theme_directory/engine.json" "$theme_directory/.platform-source.sha256"
  if [[ -d "$theme_directory/Theme Automation.app" ]]; then
    rm -rf "$theme_directory/Theme Automation.app"
  fi
  echo 'Theme service removed; preference, last palettes and prompt retained.'
  exit
fi

[[ "${1:-}" == install ]] || { echo 'Usage: theme install [--backend PATH]' >&2; exit 1; }
shift
backend=""
if [[ $# != 0 ]]; then
  [[ $# == 2 && "$1" == --backend ]] || { echo 'Usage: theme install [--backend PATH]' >&2; exit 1; }
  [[ -x "$2" ]] || { echo "Backend is not executable: $2" >&2; exit 1; }
  backend="$(cd "$(dirname "$2")" && pwd)/$(basename "$2")"
fi
command -v go >/dev/null || { echo 'Go >= 1.27 is required to build the theme engine.' >&2; exit 1; }
command -v python3 >/dev/null || { echo 'Python 3 (included in Apple Command Line Tools) is required for installation.' >&2; exit 1; }
mkdir -p "$theme_directory" "$(dirname "$plist")"
build_directory="$(mktemp -d "$theme_directory/.install.XXXXXX")"
trap 'rm -rf "$build_directory"' EXIT

(cd "$source_directory" && GOTOOLCHAIN=local go build -trimpath -o "$build_directory/theme-sync" ./cmd/theme)
xcrun swiftc -module-cache-path "$build_directory/module-cache" -parse-as-library -O \
  "$source_directory/backends/macos/main.swift" -o "$build_directory/macos-backend"
app="$build_directory/Theme Automation.app"
# Reusing the signed helper preserves its macOS Automation identity. The
# fingerprint covers source, metadata and a versioned build recipe.
platform_fingerprint="$(python3 - "$source_directory" <<'PYHASH'
import hashlib, pathlib, sys
root = pathlib.Path(sys.argv[1])
digest = hashlib.sha256(b'macos-platform-helper-v1\0')
for name in ('main.swift', 'Info.plist'):
    digest.update(name.encode() + b'\0')
    digest.update((root / 'platform/macos' / name).read_bytes())
print(digest.hexdigest())
PYHASH
)"
installed_app="$theme_directory/Theme Automation.app"
if [[ -x "$installed_app/Contents/MacOS/theme-macos" && -f "$theme_directory/.platform-source.sha256" &&
  "$(cat "$theme_directory/.platform-source.sha256")" == "$platform_fingerprint" ]]; then
  cp -pR "$installed_app" "$app"
else
  mkdir -p "$app/Contents/MacOS"
  xcrun swiftc -module-cache-path "$build_directory/module-cache" -parse-as-library -O \
    "$source_directory/platform/macos/main.swift" -o "$app/Contents/MacOS/theme-macos"
  cp "$source_directory/platform/macos/Info.plist" "$app/Contents/Info.plist"
  codesign --force --sign - --identifier com.vlvkuznetsov.theme-automation "$app"
fi
codesign --verify --strict "$app"
printf '%s\n' "$platform_fingerprint" > "$build_directory/.platform-source.sha256"
chmod 700 "$build_directory/theme-sync" "$build_directory/macos-backend"

tmux_executable="$(command -v tmux || true)"
tmux_socket="${TMUX_TMPDIR:-/tmp}/tmux-$(id -u)/default"
python3 - "$dotfiles_directory" "$theme_directory" "$build_directory" "$backend" "$tmux_executable" "$tmux_socket" "$label" <<'PY'
import json, os, pathlib, plistlib, shutil, sys
root, directory, build, backend, tmux, socket, label = sys.argv[1:]
root, directory = os.path.abspath(root), os.path.abspath(directory)
old_path = pathlib.Path(directory, 'engine.json')
if not backend and old_path.exists():
    backend = json.loads(old_path.read_text()).get('backend', '')
backend = backend or os.path.join(directory, 'macos-backend')
configuration = dict(root=root, directory=directory, backend=backend,
    platform_helper=os.path.join(directory, 'Theme Automation.app'), tmux=tmux, socket=socket)
pathlib.Path(build, 'engine.json').write_text(json.dumps(configuration, indent=2) + '\n')
plist = dict(Label=label,
    ProgramArguments=[os.path.join(directory, 'theme-sync'), 'watch', '--theme-dir', directory],
    RunAtLoad=True, KeepAlive=True, ThrottleInterval=5, LimitLoadToSessionType='Aqua', ProcessType='Background',
    StandardErrorPath=os.path.join(directory, 'service.log'))
pathlib.Path(build, 'service.plist').write_bytes(plistlib.dumps(plist))
# Validate rendering/policy in an isolated directory before stopping the service.
preflight = pathlib.Path(build, 'preflight')
preflight.mkdir()
for name in ('preference', 'mode'):
    source = pathlib.Path(directory, name)
    if source.exists():
        shutil.copyfile(source, preflight / name)
candidate = dict(configuration, directory=str(preflight), tmux='', platform_helper='')
if backend == os.path.join(directory, 'macos-backend'):
    candidate['backend'] = os.path.join(build, 'macos-backend')
pathlib.Path(build, 'preflight.json').write_text(json.dumps(candidate))
PY
plutil -lint "$build_directory/service.plist" >/dev/null
plutil -lint "$app/Contents/Info.plist" >/dev/null
"$build_directory/theme-sync" sync --config "$build_directory/preflight.json"

# Keep a complete rollback until the new LaunchAgent has been bootstrapped.
mkdir "$build_directory/previous"
for name in theme-sync macos-backend engine.json .platform-source.sha256 wezterm-bridge.lua 'Theme Automation.app'; do
  if [[ -e "$theme_directory/$name" ]]; then cp -pR "$theme_directory/$name" "$build_directory/previous/$name"; fi
done
if [[ -f "$plist" ]]; then cp -p "$plist" "$build_directory/previous/service.plist"; fi
was_loaded=false
if launchctl print "$domain/$label" >/dev/null 2>&1; then was_loaded=true; fi
rollback() {
  launchctl bootout "$domain/$label" 2>/dev/null || true
  for name in theme-sync macos-backend engine.json .platform-source.sha256 wezterm-bridge.lua 'Theme Automation.app'; do
    rm -rf "$theme_directory/$name"
    if [[ -e "$build_directory/previous/$name" ]]; then cp -pR "$build_directory/previous/$name" "$theme_directory/$name"; fi
  done
  rm -f "$plist"
  if [[ -f "$build_directory/previous/service.plist" ]]; then
    cp -p "$build_directory/previous/service.plist" "$plist"
    if $was_loaded; then launchctl bootstrap "$domain" "$plist" || true; fi
  fi
  echo 'Installation failed; previous service restored.' >&2
}
launchctl bootout "$domain/$label" 2>/dev/null || true
trap 'rollback; rm -rf "$build_directory"' ERR
for name in theme-sync macos-backend engine.json .platform-source.sha256 'Theme Automation.app'; do
  rm -rf "$theme_directory/$name"
  mv "$build_directory/$name" "$theme_directory/$name"
done
cp "$source_directory/wezterm.lua" "$build_directory/wezterm-bridge.lua"
mv -f "$build_directory/wezterm-bridge.lua" "$theme_directory/wezterm-bridge.lua"
chmod 600 "$theme_directory/engine.json" "$build_directory/service.plist"
mv -f "$build_directory/service.plist" "$plist"
launchctl bootstrap "$domain" "$plist"
trap - ERR

if [[ -e "$HOME/.p10k.zsh" || -L "$HOME/.p10k.zsh" ]]; then
  if [[ ! -e "$theme_directory/p10k-before-install.zsh" && ! -L "$theme_directory/p10k-before-install.zsh" ]]; then
    cp -P "$HOME/.p10k.zsh" "$theme_directory/p10k-before-install.zsh"
  fi
fi
ln -sfn "$source_directory/resources/p10k/p10k.zsh" "$HOME/.p10k.zsh"
echo 'Theme engine installed. Reload existing terminal configs once to adopt the engine.'
echo 'For Ghostty live reload, run theme authorize with Ghostty open and allow Automation.'
