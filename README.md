# Dotfiles

This repository contains configuration files for setting up a development environment across macOS, Linux, and WSL.

For the matching Neovim configuration, see [nvim-config](https://github.com/Pepetka/nvim-config).

## Configuration Files

### Terminal & Shell

- **`.zshrc`** — Main Zsh configuration with Oh My Zsh, Powerlevel10k, plugins, and environment settings
- **`aliases.zsh`** — Custom aliases, functions, and the `theme` command
- **`zsh/secrets.zsh.example`** — Example file for private API keys and secrets; copy to `~/.config/zsh/secrets.zsh`
- **`alacritty.toml`** — Alacritty terminal emulator configuration
- **`theme/resources/alacritty/`** — Alacritty dark/light palettes (old paths are compatibility links)
- **`wezterm.lua`** — WezTerm terminal emulator configuration
- **`ghostty/config`** — Ghostty terminal emulator configuration
- **`theme/resources/ghostty/`** — Ghostty theme selections

### Terminal Multiplexers

- **`tmux.conf`** — Tmux configuration with custom keybindings, powerline-style status bar, popups, scratchpad, and plugins
- **`theme/resources/tmux/`** — Tmux dark/light palettes using tmux user options (`@bg`, `@fg`, ...)

### Shell Prompt

- **`theme/resources/p10k/p10k.zsh`** — One Powerlevel10k configuration using the terminal ANSI palette

### Shell Scripts

- **`theme/theme`** — Theme engine command; `shell/theme-switch.sh` remains a compatibility wrapper
- **`theme/`** — Portable Go engine, macOS backend, application adapters and resources
- **`shell/keyboard.sh`** — macOS input source indicator used by the Tmux status bar (shows `RU`/`EN`/etc.)

### Editor

- **[nvim-config](https://github.com/Pepetka/nvim-config)** — Matching Neovim configuration with native LSP,
  fzf-lua, and TokyoNight theme switching

## Screenshots

### Terminal — TokyoNight Moon (dark)

![Terminal dark](images/terminal.png)

### Terminal — TokyoNight Day (light)

![Terminal light](images/terminal-light.png)

## Shared Theme Engine

One engine selects TokyoNight Moon/Day for Ghostty, WezTerm, Alacritty, tmux and Neovim.
It supports persistent manual overrides and automatic macOS appearance tracking.
All theme sources, application adapters, palettes, tests and installation logic live in [theme/](theme/README.md).

```bash
theme install    # requires Go >= 1.27, Apple Command Line Tools and Python 3
theme authorize  # with Ghostty open: allow Theme Automation to reload its config
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
contract, while application integration uses adapter interfaces. The first installer supports macOS.
See [architecture, migration, extending and verification](theme/README.md) for details.

## Color Schemes

- **TokyoNight Moon** — primary dark color scheme for tmux, Ghostty, Alacritty, WezTerm, Neovim, and p10k
- **TokyoNight Day** — primary light color scheme for tmux, Ghostty, Alacritty, WezTerm, Neovim, and p10k

## Installation Paths

Files should be placed/symlinked in the following locations:

```text
~/.zshrc
~/.p10k.zsh
~/.config/alacritty/alacritty.toml
~/.config/ghostty/config
~/.config/theme/
~/.config/zsh/secrets.zsh
~/.wezterm.lua
~/.tmux.conf
~/.tmux/keyboard.sh
```

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/Pepetka/dotfiles.git ~/workspace/dotfiles
   ```

2. Create necessary directories:

   ```bash
   mkdir -p ~/.config/alacritty
   mkdir -p ~/.config/ghostty
   mkdir -p ~/.config/theme
   mkdir -p ~/.config/zsh
   mkdir -p ~/.tmux
   ```

3. Create symbolic links:

   ```bash
   # Zsh configuration
   ln -sf ~/workspace/dotfiles/configs/.zshrc ~/.zshrc

   # Terminal emulators
   ln -sf ~/workspace/dotfiles/configs/alacritty.toml ~/.config/alacritty/alacritty.toml
   ln -sf ~/workspace/dotfiles/configs/ghostty/config ~/.config/ghostty/config
   ln -sf ~/workspace/dotfiles/configs/wezterm.lua ~/.wezterm.lua

   # Tmux
   ln -sf ~/workspace/dotfiles/configs/tmux.conf ~/.tmux.conf
   ln -sf ~/workspace/dotfiles/shell/keyboard.sh ~/.tmux/keyboard.sh
   ```

   Aliases are sourced directly from `$DOTFILES/configs/aliases.zsh` in `.zshrc`, so no separate symlink is needed.

4. Run `theme install` on macOS to create theme state, link the ANSI p10k config and start the engine; then `theme authorize` with Ghostty open.

5. Install Tmux plugins:
   - Install [Tmux Plugin Manager (TPM)](https://github.com/tmux-plugins/tpm)
   - Open Tmux and press `prefix + I` to install plugins

## Prerequisites

### Required

- **[Oh My Zsh](https://ohmyz.sh/)** — Framework for Zsh configuration
- **[Powerlevel10k](https://github.com/romkatv/powerlevel10k)** — Zsh theme
  - `~/.p10k.zsh` points to the single ANSI config installed by `theme install`; do not run `p10k configure`
    unless you want to overwrite it

### Optional Terminal Emulators

- **[Alacritty](https://alacritty.org/)** — Fast, cross-platform terminal emulator
- **[WezTerm](https://wezfurlong.org/wezterm/)** — GPU-accelerated terminal emulator
- **[Ghostty](https://ghostty.org/)** — Fast, native, feature-rich terminal emulator

### Optional Terminal Multiplexers

- **[Tmux](https://github.com/tmux/tmux)** — Terminal multiplexer
- **[Tmux Plugin Manager](https://github.com/tmux-plugins/tpm)** — Plugin manager for Tmux

### Zsh Plugins (installed via Oh My Zsh)

The `.zshrc` includes several plugins that enhance the shell experience:

- `zsh-autosuggestions` — Command suggestions
- `zsh-syntax-highlighting` — Syntax highlighting
- `fzf` — Fuzzy finder integration
- `git` — Git aliases and functions
- `nvm` — Node Version Manager integration
- `npm`, `bun`, `docker`, `direnv`, `extract`, `sudo`, and more (see `.zshrc` for full list)

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
- **TokyoNight Moon** color scheme

### Aliases & Functions

The configuration includes numerous aliases for:

- Git operations (`gs`, `ga`, `gc`, `gcom`, etc.)
- Navigation (`ls` with `eza`, directory shortcuts)
- Development tools (npm, docker, etc.)
- AI coding agent providers (`deepseek`, `zai`, `minimax`, `kimic`)

### Custom Functions

- `theme` — Select auto/dark/light, install, inspect and reconcile shared theming
- `yy()` — Yazi file manager integration
- `gcom()` — Smart git checkout `master`/`main`
- `hurl_pretty()` — Pretty-printed HTTP responses
- `ts3()` — Tmux three-pane development layout

## License

This repository is distributed under the MIT License. See the [LICENSE](LICENSE) file for details.
