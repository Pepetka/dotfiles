# Theme engine

One policy (`auto`, `dark`, `light`) drives Neovim, Ghostty, WezTerm, Alacritty,
tmux and the ANSI Powerlevel10k prompt. Sources and resources live here; generated
state lives under `${XDG_CONFIG_HOME:-~/.config}/theme`.

## Use

```sh
theme install                 # engine, system backend and service (macOS or Linux)
theme install --manual        # engine only: manual dark/light, no service or backend
theme authorize               # macOS, with Ghostty open: allow helper Automation once
theme auto                    # follow the system appearance (default; needs the service)
theme dark                    # persistent manual override
theme light                   # persistent manual override
theme sync                    # force reapplication without changing preference
theme status                  # backend, service, policy, mode and per-adapter diagnostics
theme uninstall               # stop service; keep preference and last palettes
```

Installation needs Go 1.27+, or Go 1.21+ which downloads the 1.27 toolchain
declared in `go.mod` (e.g. Ubuntu's `golang-go`); on macOS also Apple Command Line Tools. The only Go dependency, `godbus/dbus` for the
Linux backend, is pinned in `go.mod`/`go.sum`. Neither installation nor tests require Python.

### Manual mode

`theme install --manual` builds the engine once without a system backend or
service, on macOS and Linux. `theme dark` and `theme light` apply every adapter,
including the Ghostty reload; `theme auto` is rejected without changing the
preference. A persisted `auto` becomes the current effective mode during
installation, and the palettes are published immediately. Running `theme install`
later restores the service and backend and, if manual installation replaced `auto`,
restores `auto` as well; `--manual` again removes the service and backend.
Automatic installation first checks that the system backend can read the appearance
and otherwise fails without changing the current installation.
Reload existing terminal configurations once when
migrating from independent appearance selection. Existing Neovim instances
already observing `mode` need no migration. Reinstall after moving dotfiles or
changing the config root. `shell/theme-switch.sh` remains a compatibility wrapper
for the original shell command; former resource symlinks have been removed.

Manual overrides survive system appearance changes, service restarts and login.
CLI switches also work with the service stopped. Continuous automatic updates
and recovery require the service. Uninstall retains the chosen policy and generated
files. The prompt is sourced by `.zshrc` directly from `resources/p10k/p10k.zsh`;
installation no longer manages `~/.p10k.zsh`.

## Architecture

`cmd/theme` is the CLI; `internal/engine` owns policy resolution, serialized
publication, backend supervision and retry. `internal/adapters` owns rendering
and application reloads. `backends/macos` and `backends/linux` only observe the
system. The separate `platform/macos` helper handles Ghostty Automation; it is not
an appearance backend. `resources` owns the existing Moon/Day palettes and prompt
configurations.

`cmd/theme-install` and `internal/install` build the binaries, generate
`engine.json` and the service definition, validate the candidate in isolation and
restore the previous installation if anything fails. `internal/service` renders and
controls the per-user service: a LaunchAgent on macOS, a systemd user unit on Linux.
`install.sh` only invokes the Go installer. Native macOS tools compile Swift and
preserve signed bundle metadata. The helper fingerprint is compatible with earlier installs.

Runtime files:

| File | Meaning |
| --- | --- |
| `engine.json` | Installed source root, backend (empty in manual mode), service manager, platform helper, tmux socket |
| `preference` | Persistent `auto`, `dark` or `light` selection |
| `mode` | Effective `dark` or `light`; stable contract with Neovim |
| `*-theme.*` | Generated application-specific configuration |
| `wezterm-bridge.lua` | Installed bridge to the generated WezTerm theme |
| `status.json` | Last reconciliation, backend error and adapter results; status validates policy/mode files |
| `service.log` | Transitions into/out of synchronization failure |

The service definition lives outside the theme directory:
`~/Library/LaunchAgents/com.vlvkuznetsov.theme-sync.plist` on macOS and
`~/.config/systemd/user/theme-sync.service` on Linux (the user manager's default
search path, independent of a custom `XDG_CONFIG_HOME`).

CLI and daemon use the same file lock and reread preference under that lock.
Outputs are replaced atomically; unchanged content is not rewritten. Legacy
symlinks are replaced without modifying their targets. A failing adapter does
not block publication or other adapters. Failed reloads retry every second even if
the file already has the desired bytes; while healthy, the daemon reacts to backend
events and otherwise rechecks reload targets (new tmux servers or Ghostty processes)
every 30 seconds, since applications read the published files at startup.
`sync` explicitly requests reapplication.

If an automatic backend fails, the engine keeps the last valid effective mode,
using dark only on a cold start without valid state. Manual modes do not require
a working backend. Invalid persisted preference is reported and retained; use
`theme auto`, `theme dark` or `theme light` to replace it explicitly.

## Add a system backend

A backend is an executable with this protocol:

```text
backend read  -> one dark/light line, exit 0
backend watch -> initial dark/light line within 3 seconds, then changed modes
```

Write diagnostics to stderr; stdout contains only modes. The engine restarts a
failed watcher after one second. Once initialized, a watcher may remain silent
until the next change. Install with `theme install --backend /absolute/path`;
subsequent installs preserve that choice. No Go-core modification is required.

The macOS implementation reads refreshed global preferences, observes appearance,
wake and session notifications, and polls once per second to recover missed events.

The Linux implementation reads `org.freedesktop.appearance color-scheme` from the
XDG Desktop Portal (GNOME, including Ubuntu, KDE and others): `1` is dark, `0` and `2`
are light. It subscribes to `SettingChanged` before its initial read and rereads every
two seconds to recover from a restarted portal. Without a portal it falls back to
`gsettings get org.gnome.desktop.interface color-scheme`. The systemd unit belongs to
`graphical-session.target` and logs to `service.log`; `journalctl --user -u theme-sync`
shows service starts. Without a systemd user manager (WSL, containers) use `--manual`.

## Add an application adapter

Implement `engine.Adapter` and register it in `internal/adapters/registry.go`:

- `Name` supplies the diagnostic key.
- `Render(Config, Mode)` returns generated files and has no side effects.
- `Apply(context.Context, Config, ApplyRequest)` applies the desired mode and
  returns an independent state/error. Requests include previous state, content
  changes and explicit force so reload failures and application restarts recover.

Application-specific behavior belongs in the adapter, not the core's event loop.
Platform-specific actions stay in platform helpers. Source palette assets are
never rewritten when changing the mode.

## Live updates

- Ghostty on Linux reloads through its own systemd unit
  (`systemctl --user reload app-com.mitchellh.ghostty.service`, i.e. SIGUSR2) when
  that unit is active and reloadable. Ghostty started outside the unit keeps the
  published file until a manual `reload_config` (`ctrl+shift+,`); the engine never
  signals it directly because versions without a handler would terminate.
- Ghostty on macOS includes `ghostty-theme.conf`; a hidden helper requests native
  `reload_config` without focus changes. It detects running Ghostty before acting
  and never launches it. `theme authorize` uses the same LaunchServices helper
  identity as the daemon. Ordinary retries never prompt for Automation.
- WezTerm watches its generated Lua file and reads it through `dofile`, avoiding
  module caches. Alacritty uses its live imported palette. Their `published`
  status does not claim a GUI acknowledgement.
- Neovim watches `mode` with a periodic recovery check; terminal/fzf ANSI colors
  follow the host terminal palette.
- tmux updates the default local socket (including `TMUX_TMPDIR` captured during
  installation), verifies the server mode marker and refreshes attached clients.
  Separate custom sockets and remote servers are outside this installation.
- The prompt uses one ANSI configuration. Previously emitted truecolor output
  from arbitrary programs cannot be recolored retroactively.

Ghostty requires 1.3+ and Automation permission for **Theme Automation**. Keep
`macos-applescript` enabled. Open Ghostty before `theme authorize`; if permission
is denied, status explains the error and other adapters continue. Reinstallation reuses the signed helper
when its source and metadata are unchanged, preserving
its Automation identity. A source change requiring a helper rebuild may require renewed
permission according to macOS code-signing/TCC policy.

## CLI utilities

Hunk and Codex provide manual **Tokyo Night Moon** (dark) and **Tokyo Night Day**
(light) variants using the installed Neovim TokyoNight palettes. Neither the
engine nor a background process changes their theme selection.

Hunk loads both themes from `configs/hunk/config.toml`. Link this self-contained
configuration to `${XDG_CONFIG_HOME:-~/.config}/hunk/config.toml` to share themes
and keybindings. Transparency remains enabled. Press `t` to choose a theme;
an existing session may need one reopen to discover changes to the theme catalog.
Subsequent selections apply immediately. Hunk uses a bundled foreground for
unhighlighted files, independently of custom syntax scope overrides.

For Codex, link `resources/codex/Tokyo Night Moon.tmTheme` and
`resources/codex/Tokyo Night Day.tmTheme` into `$CODEX_HOME/themes`
(default `~/.codex/themes`), or use the [Codex installer](../configs/codex/README.md)
to install themes together with shared settings and profiles. Choose a theme with
`/theme`. Neutral text and ordinary code backgrounds inherit the terminal colors.
Added and removed diff lines use subtle green and red backgrounds matched to
each variant. Codex persists your selection to
`[tui].theme` in `$CODEX_HOME/config.toml`; preserve other existing settings.
Codex 0.162 still owns and caches some UI colors independently of syntax themes.
No Codex adapter or client rebuild is needed.

Choose Day on a light terminal background and Moon on a dark one.

## Verification

```sh
theme/tests/run.sh
# Linux only: cross-compiled tests and installer acceptance in an Ubuntu container
theme/tests/linux.sh          # THEME_LINUX_BASE=<image> overrides ubuntu:24.04
# The previous test entrypoint delegates to the same suite:
shell/tests/theme-sync.sh
```

`run.sh` needs Go, zsh and Lua or Neovim; Docker is optional. The race detector is
used when cgo and a C compiler are available (on Ubuntu: `build-essential`).

The suite covers policy persistence, CLI/daemon concurrency, atomic publication,
symlink safety, honest status for corrupt state, canceled requests, late helper replies,
independent failures/retries, backend crashes/malformed output/
timeouts, bounded child-process output, a real isolated tmux server, WezTerm's
bridge, Linux cross-build, Swift compilation, installer rollback with mocked
launchd/systemd commands, manual mode, signed-helper reuse/invalidation, and the
Linux backend against a fake portal on a private D-Bus session bus. It never changes desktop
appearance or the user's tmux server.

Run the matching Neovim theme, UI and startup suites as well. For manual acceptance,
test `auto/dark/light` with running terminals, Ghostty focused/unfocused, multiple
windows and quick terminal, an idle prompt, tmux and an open Neovim/fzf picker;
then repeat after sleep and service restart. GUI repaint acknowledgement and
macOS permission prompts require a real desktop check.
