#!/usr/bin/env bash
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
test_directory="$(mktemp -d "${TMPDIR:-/tmp}/theme-app-tests.XXXXXX")"
trap 'rm -rf "$test_directory"' EXIT
(cd "$source_directory" && go build -o "$test_directory/engine" ./cmd/theme)
python3 - "$source_directory" "$test_directory" <<'PY'
import json, os, pathlib, shutil, subprocess, sys
source, temporary = map(pathlib.Path, sys.argv[1:])
root = source.parent
config_root = temporary / 'config with spaces'
directory = config_root / 'theme'
directory.mkdir(parents=True)
for app, name in [('ghostty', 'config'), ('alacritty', 'alacritty.toml')]:
    target = config_root / app / name
    target.parent.mkdir()
    target.symlink_to(root / 'configs' / ('ghostty/config' if app == 'ghostty' else name))
shutil.copyfile(source / 'wezterm.lua', directory / 'wezterm-bridge.lua')
backend = temporary / 'backend'
backend.write_text('#!/bin/sh\necho light\n')
backend.chmod(0o700)
settings = directory / 'engine.json'
settings.write_text(json.dumps(dict(root=str(root), directory=str(directory), backend=str(backend))))
env = dict(os.environ, XDG_CONFIG_HOME=str(config_root))
def run(command):
    result = subprocess.run(command, env=env, text=True, capture_output=True, timeout=10)
    if result.returncode:
        raise RuntimeError(f'{command}: {result.stdout}\n{result.stderr}')
    return result.stdout
for mode, style, background in [('dark', 'moon', '#222436'), ('light', 'day', '#e1e2e7')]:
    run([str(temporary / 'engine'), mode, '--config', str(settings)])
    ghostty = shutil.which('ghostty')
    if ghostty:
        content = run([ghostty, '+show-config'])
        assert f'background = {background}\n' in content
        assert f'theme = TokyoNight {style.title()}\n' in content
        run([ghostty, '+validate-config', '--config-file=' + str(config_root / 'ghostty/config')])
    wezterm = shutil.which('wezterm')
    if wezterm:
        check = temporary / 'wezterm-check.lua'
        quoted = json.dumps(str(root / 'configs/wezterm.lua'), ensure_ascii=False)
        check.write_text(f'local config = dofile({quoted})\nassert(config.color_scheme == "tokyonight_{style}")\nreturn config\n')
        run([wezterm, '--config-file', str(check), 'show-keys'])
    alacritty = shutil.which('alacritty')
    native_alacritty = pathlib.Path('/Applications/Alacritty.app/Contents/MacOS/alacritty')
    if not alacritty and native_alacritty.exists():
        alacritty = str(native_alacritty)
    if alacritty:
        content = run([alacritty, 'migrate', '--dry-run', '--config-file', str(config_root / 'alacritty/alacritty.toml')])
        assert background in content
# Status must report persisted policy corruption even if the cached adapter
# snapshot still describes a healthy earlier reconciliation.
(directory / 'preference').write_text('broken\n')
result = subprocess.run([str(temporary / 'engine'), 'status', '--config', str(settings)], env=env, text=True, capture_output=True, timeout=10)
assert result.returncode != 0
reported = json.loads(result.stdout)
assert reported['preference'] == 'broken' and 'preference' in reported['state_error']
assert (directory / 'preference').read_text() == 'broken\n'
print('Native app configs passed: generated dark/light, symlink imports, XDG/spaces and honest CLI status')
PY
