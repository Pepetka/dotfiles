# Dotfiles

This repository contains configuration files (dotfiles) for setting up a development environment.

## Configuration Files

### Terminal & Shell
- **`.zshrc`** — Main Zsh configuration with Oh My Zsh, plugins, and custom settings
- **`aliases.zsh`** — Custom aliases and functions for shell productivity
- **`alacritty.toml`** — Alacritty terminal emulator configuration
- **`wezterm.lua`** — WezTerm terminal emulator configuration

### Terminal Multiplexers
- **`tmux.conf`** — Tmux configuration with custom keybindings and plugins
- **`zellij/`** — Zellij terminal multiplexer configuration and layouts
  - `config.kdl` — Main Zellij configuration
  - `layouts/` — Custom layouts
  - `plugins/` — Custom plugins
  - `shell/` — Shell scripts for status bar (CPU, memory, network, etc.)

## Installation Paths

Files should be placed in the following locations:

```
~/.zshrc
~/.config/alacritty/alacritty.toml
~/.wezterm.lua
~/.tmux.conf
~/.oh-my-zsh/aliases/aliases.zsh
~/.config/zellij/
```

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/your-username/dotfiles.git ~/dotfiles
   ```

2. Create necessary directories:
   ```bash
   mkdir -p ~/.config/alacritty
   mkdir -p ~/.config/zellij
   mkdir -p ~/.oh-my-zsh/aliases
   ```

3. Create symbolic links:
   ```bash
   # Zsh configuration
   ln -sf ~/dotfiles/configs/.zshrc ~/.zshrc
   ln -sf ~/dotfiles/configs/aliases.zsh ~/.oh-my-zsh/aliases/aliases.zsh

   # Terminal emulators
   ln -sf ~/dotfiles/configs/alacritty.toml ~/.config/alacritty/alacritty.toml
   ln -sf ~/dotfiles/configs/wezterm.lua ~/.wezterm.lua

   # Terminal multiplexers
   ln -sf ~/dotfiles/configs/tmux.conf ~/.tmux.conf
   ln -sf ~/dotfiles/configs/zellij ~/.config/zellij
   ```

## Prerequisites

### Required
- **[Oh My Zsh](https://ohmyz.sh/)** — Framework for Zsh configuration
- **[Powerlevel10k](https://github.com/romkatv/powerlevel10k)** — Zsh theme (referenced in .zshrc)

### Optional Terminal Emulators
- **[Alacritty](https://alacritty.org/)** — Fast, cross-platform terminal emulator
- **[WezTerm](https://wezfurlong.org/wezterm/)** — GPU-accelerated terminal emulator

### Optional Terminal Multiplexers
- **[Tmux](https://github.com/tmux/tmux)** — Terminal multiplexer
- **[Tmux Plugin Manager](https://github.com/tmux-plugins/tpm)** — Plugin manager for Tmux
- **[Zellij](https://zellij.dev/)** — Modern terminal multiplexer

### Zsh Plugins (installed via Oh My Zsh)
The .zshrc includes several plugins that enhance the shell experience:
- `zsh-autosuggestions` — Command suggestions
- `zsh-syntax-highlighting` — Syntax highlighting
- `fzf` — Fuzzy finder integration
- `git` — Git aliases and functions
- `nvm` — Node Version Manager integration
- And more (see .zshrc for full list)

## Features

### Aliases & Functions
The configuration includes numerous aliases for:
- Git operations (`gs`, `ga`, `gc`, etc.)
- Navigation (`ls` with eza, directory shortcuts)
- Development tools (npm, docker, etc.)
- Text editing shortcuts

### Custom Functions
- `yy()` — Yazi file manager integration
- `gcom()` — Smart git checkout main/master
- `curl()` — Proxy handling for specific domains
- `hurl_pretty()` — Pretty-printed HTTP responses

## License

This repository is distributed under the MIT License. See the [LICENSE](LICENSE) file for details.

