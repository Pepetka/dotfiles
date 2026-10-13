# Dotfiles

This repository contains configuration files for setting up a development environment across macOS, Linux, and WSL.

For the matching Neovim configuration, see [nvim-config](https://github.com/Pepetka/nvim-config).

## Configuration Files

### Terminal & Shell

- **`configs/zsh/.zshrc`** — Main Zsh configuration with Oh My Zsh, Powerlevel10k, plugins, and environment settings
- **`configs/zsh/aliases.zsh`** — Custom aliases and the `theme` command
- **`configs/zsh/functions.zsh`** — Git, Yazi and Tmux helpers
- **`configs/zsh/secrets.zsh.example`** — Example file for private API keys and secrets
- **`configs/zsh/local.zsh.example`** — Optional device-specific tool paths and settings
- **`configs/zsh/local.after.zsh.example`** — Optional device-specific setup that registers completions
- **`configs/alacritty/alacritty.toml`** — Alacritty terminal emulator configuration
- **`theme/resources/alacritty/`** — Alacritty dark/light palettes
- **`configs/wezterm/wezterm.lua`** — WezTerm terminal emulator configuration
- **`configs/ghostty/config`** — Ghostty terminal emulator configuration
- **`theme/resources/ghostty/`** — Ghostty theme selections
- **`configs/hunk/config.toml`** — Hunk diff viewer settings, keybindings and Tokyo Night Moon/Day themes

### Terminal Multiplexers

- **`configs/tmux/tmux.conf`** — Tmux configuration with custom keybindings, powerline-style status bar, popups,
  scratchpad, and plugins
- **`theme/resources/tmux/`** — Tmux dark/light palettes using tmux user options (`@bg`, `@fg`, ...)

### Shell Prompt

- **`theme/resources/p10k/p10k.zsh`** — One Powerlevel10k configuration using the terminal ANSI palette

### Shell Scripts

- **`theme/theme`** — Theme engine command; `shell/theme-switch.sh` remains a compatibility wrapper
- **`theme/`** — Portable Go engine, macOS backend, application adapters and resources
- **`shell/keyboard.sh`** — Keyboard layout indicator for the Tmux status bar (`RU`/`EN`/etc.) on macOS and
  Linux (GNOME, KDE, X11)

### Editor

- **[nvim-config](https://github.com/Pepetka/nvim-config)** — Matching Neovim configuration with native LSP,
  fzf-lua, and TokyoNight theme switching

### AI CLI

- **[configs/agents/](configs/agents/)** — Shared global instructions (`AGENTS.md`, linked to
  `~/.codex/AGENTS.md` and `~/.claude/CLAUDE.md`) and skills, linked by `link-skills.sh`
- **[configs/codex/](configs/codex/README.md)** — Reference copy of portable Codex preferences, updated manually

## Screenshots

### Terminal — TokyoNight Moon (dark)

![Terminal dark](images/terminal.png)

### Terminal — TokyoNight Day (light)

![Terminal light](images/terminal-light.png)

## Shared Theme Engine

One engine selects TokyoNight Moon/Day for Ghostty, WezTerm, Alacritty, tmux and Neovim.
It supports persistent manual overrides and automatic appearance tracking on macOS
and Linux (XDG Desktop Portal, e.g. Ubuntu GNOME), or a manual-only mode without a service.
All theme sources, application adapters, palettes, tests and installation logic live in [theme/](theme/README.md).

```bash
theme install    # requires Go (and Apple Command Line Tools on macOS)
theme install --manual  # no service or system backend: dark/light only
theme authorize  # macOS, with Ghostty open: allow Theme Automation to reload its config
theme auto       # follow the system (default)
theme dark       # persistent manual dark mode
theme light      # persistent manual light mode
theme status     # service, preference, mode and per-application diagnostics
theme sync       # reapply without changing preference
theme uninstall  # retain preference, last palettes and ANSI prompt
```

Reload existing terminal configurations once after migration. Thereafter changes do not require input or focus changes.
Neovim retains its existing dark/light mode-file contract. Powerlevel10k uses the terminal ANSI palette.
The macOS backend is separate from the portable Go engine; additional backends use an executable `read/watch`
contract, while application integration uses adapter interfaces. Installation supports macOS (LaunchAgent)
and Linux (systemd user service).
See [architecture, migration, extending and verification](theme/README.md) for details.

## Color Schemes

- **TokyoNight Moon** — primary dark color scheme for tmux, Ghostty, Alacritty, WezTerm, Neovim, and p10k
- **TokyoNight Day** — primary light color scheme for tmux, Ghostty, Alacritty, WezTerm, Neovim, and p10k

## Installation Paths

Files should be placed/symlinked in the following locations:

```text
~/.zshrc
~/.config/alacritty/alacritty.toml
~/.config/ghostty/config
~/.config/hunk/config.toml
~/.config/theme/
~/.config/zsh/local.zsh        # optional, local file
~/.config/zsh/local.after.zsh  # optional, local file
~/.config/zsh/secrets.zsh      # optional, local file
~/.wezterm.lua
~/.tmux.conf
~/.tmux/keyboard.sh
```

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/Pepetka/dotfiles.git "$HOME/dotfiles"
   export DOTFILES="$HOME/dotfiles" # Use your actual checkout location.
   ```

2. Create necessary directories:

   ```bash
   config_root="${XDG_CONFIG_HOME:-$HOME/.config}"
   mkdir -p "$config_root/alacritty" "$config_root/ghostty" "$config_root/hunk"
   mkdir -p "$config_root/theme" "$config_root/zsh" "$HOME/.tmux"
   mkdir -p "${ZDOTDIR:-$HOME}"
   ```

3. Back up existing files, then create symbolic links for the applications you use:

   ```bash
   # Zsh configuration
   ln -s "$DOTFILES/configs/zsh/.zshrc" "${ZDOTDIR:-$HOME}/.zshrc"

   # Terminal emulators
   ln -s "$DOTFILES/configs/alacritty/alacritty.toml" "$config_root/alacritty/alacritty.toml"
   ln -s "$DOTFILES/configs/ghostty/config" "$config_root/ghostty/config"
   ln -s "$DOTFILES/configs/hunk/config.toml" "$config_root/hunk/config.toml"
   ln -s "$DOTFILES/configs/wezterm/wezterm.lua" "$HOME/.wezterm.lua"

   # Tmux
   ln -s "$DOTFILES/configs/tmux/tmux.conf" "$HOME/.tmux.conf"
   ln -s "$DOTFILES/shell/keyboard.sh" "$HOME/.tmux/keyboard.sh"

   # AI agents: shared instructions and skills
   ln -s "$DOTFILES/configs/agents/AGENTS.md" "$HOME/.codex/AGENTS.md"
   ln -s "$DOTFILES/configs/agents/AGENTS.md" "$HOME/.claude/CLAUDE.md"
   sh "$DOTFILES/configs/agents/link-skills.sh"
   ```

   Rerun `link-skills.sh` after adding a skill to `configs/agents/skills` or installing one into `~/.agents/skills`.

   Aliases are sourced directly from `$DOTFILES/configs/zsh/aliases.zsh` in `.zshrc`, so no separate symlink is needed.
   `.zshrc` derives `DOTFILES` from its symlink target; the checkout may live anywhere.
   An explicitly exported `DOTFILES` takes precedence. Keep `XDG_CONFIG_HOME` consistent
   between the shell, terminal applications and service installation when using a custom root.

4. Run `"$DOTFILES/theme/theme" install` to build the engine, publish the palettes and start the service
   (`--manual` without a service, e.g. in WSL or containers). On macOS, then run
   `"$DOTFILES/theme/theme" authorize` with Ghostty open.

5. Install Tmux plugins:
   - Install [Tmux Plugin Manager (TPM)](https://github.com/tmux-plugins/tpm)
   - Open Tmux and press `prefix + I` to install plugins

## Migrating an existing checkout

Configuration sources now live in `configs/<application>/`. Recreate existing
application symlinks using the paths above (back up real files before replacing them).
The old flat paths and palette compatibility links in `configs/` have been removed.
Copy and configure `configs/zsh/local.zsh.example` before reloading Zsh: tool initialization
previously embedded in `.zshrc` now belongs to the device-local file.

Palettes have one canonical source in `theme/resources/<application>/`.
The engine reads those resources directly and publishes the selected palette under
`${XDG_CONFIG_HOME:-$HOME/.config}/theme/`; terminal configs consume that generated
state to follow dark/light changes. These generated files are not duplicate source links.

On Ubuntu, `theme install` builds the engine and the portal backend and enables the
`theme-sync` systemd user service; `theme install --manual` installs the engine
without a service for manual `theme dark`/`theme light` switching (also on macOS).
Without the engine, terminals fall back to their built-in colors. WezTerm uses Moon
only while the installed bridge exists without a published theme (e.g. after uninstall).

## Prerequisites

### Shell setup

- **Zsh** — Required for `.zshrc`; aliases and completion work without optional frameworks
- **[Oh My Zsh](https://ohmyz.sh/)** — Optional framework for plugins
- **[Powerlevel10k](https://github.com/romkatv/powerlevel10k)** — Optional Zsh theme
  - Zsh loads `theme/resources/p10k/p10k.zsh` directly when Powerlevel10k is available

The shared Zsh config initializes completion, optional Oh My Zsh plugins, aliases
and the ANSI prompt. It does not discover Homebrew, alter library paths or start SDKs.
`ls` uses eza only when it is installed; otherwise the system command stays available.
`EDITOR` defaults to `nvim` when installed, otherwise `vim`, otherwise `vi`; the config-editing
aliases open `$EDITOR`.
The default plugins are `z`, `sudo`, `zsh-autosuggestions` and `zsh-syntax-highlighting`;
the Oh My Zsh `git` plugin is omitted because `aliases.zsh` defines its own Git aliases.
Path entries are deduplicated.

Copy `configs/zsh/local.zsh.example` to
`${XDG_CONFIG_HOME:-$HOME/.config}/zsh/local.zsh` and uncomment the settings this
machine needs. It loads after the Powerlevel10k instant prompt and before framework
setup, and contains examples for Homebrew, user executable paths, NVM, Bun,
editor and plugin preferences. Setup that registers completions, such as SDKMAN,
belongs in `local.after.zsh` (see `configs/zsh/local.after.zsh.example`), which loads
after the framework; loaded earlier, such scripts run a second `compinit`.
On Ubuntu, use native packages and user paths; Homebrew is optional. Ubuntu's
`/etc/zsh/zshrc` runs `compinit` before Oh My Zsh does it again; skip the first call:

```bash
echo 'skip_global_compinit=1' >> "${ZDOTDIR:-$HOME}/.zshenv"
```

Ubuntu 24.04 and older ship tmux below 3.5, which ignores `extended-keys-format`.
Alacritty's `SimpleFullscreen` startup mode is macOS-only; on Linux the window opens
windowed.
Keep API keys in `zsh/secrets.zsh` under the same config root, outside Git.
Codex authentication, trusted projects and device-specific MCP configuration
remain local; `configs/codex/config.toml` is only a reference for manual comparison.

The keyboard-layout indicator reads macOS input sources, GNOME `input-sources` (`gsettings`), KDE
(`org.kde.keyboard` over `gdbus`) or X11 (`xkb-switch`, else `setxkbmap` with `xset`). Tmux hides it when the
script prints nothing at startup, e.g. over SSH or without a graphical session.
On macOS the indicator maps input sources such as `ABC`, `U.S.`, `RussianWin` or `Ukrainian-PC`
to the same `EN`/`RU`/`UA` codes. Check shell portability with `bash shell/tests/portability.sh`.

### Optional Terminal Emulators

- **[Alacritty](https://alacritty.org/)** — Fast, cross-platform terminal emulator
- **[WezTerm](https://wezfurlong.org/wezterm/)** — GPU-accelerated terminal emulator
- **[Ghostty](https://ghostty.org/)** — Fast, native, feature-rich terminal emulator

### Optional Terminal Multiplexers

- **[Tmux](https://github.com/tmux/tmux)** — Terminal multiplexer
- **[Tmux Plugin Manager](https://github.com/tmux-plugins/tpm)** — Plugin manager for Tmux

### Zsh Plugins (installed via Oh My Zsh)

The default list enables only plugins that are installed:

- `z` — Directory jumping
- `sudo` — Prefix the previous command with `sudo`
- `zsh-autosuggestions` — Command suggestions (custom plugin, clone into `$ZSH_CUSTOM/plugins`)
- `zsh-syntax-highlighting` — Syntax highlighting (custom plugin, clone into `$ZSH_CUSTOM/plugins`)

Set `plugins=(...)` in `local.zsh` for more, e.g. `fzf`, `npm`, `bun`, `docker` or `direnv`.

## Features

### Tmux

- **Prefix** changed to `Ctrl + a`
- **Vi-style copy mode** with system clipboard integration
- **Mouse support** for selection, resizing, and scrolling
- **Popups**:
  - `prefix + f` — floating scratchpad session
  - `prefix + Ctrl + f` — fzf session switcher
  - `prefix + ?` — keybindings cheatsheet
- **Synchronize panes** toggle with `prefix + e` and visual `SYNC` indicator
- **Window management**:
  - `prefix + Shift + h/l` — swap window left/right
  - `prefix + x/X/Q` — kill pane/window/session
- **Powerline-style status bar** with session, layout, CPU/RAM, battery, and time widgets
- **TokyoNight Moon/Day** palette selected by the theme engine

### Aliases & Functions

The configuration includes numerous aliases for:

- Git operations (`gs`, `ga`, `gc`, `gst`, etc.)
- Listing (`ls`/`ls2` with `eza` when installed)
- Opening configurations in `$EDITOR` (`zsho`, `tmux.conf`, `ghostty.conf`, `zshlocal`, etc.)
- Tmux and NVM shortcuts (`renum`, `nvmc`, `nvml`)

### Custom Functions

- `theme` — Select auto/dark/light, install, inspect and reconcile shared theming
- `yy()` — Yazi file manager integration
- `gcom()` — Smart git checkout `master`/`main`
- `ts3()` — Tmux three-pane development layout

## License

This repository is distributed under the MIT License. See the [LICENSE](LICENSE) file for details.
