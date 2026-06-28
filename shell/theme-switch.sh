#!/usr/bin/env bash
set -euo pipefail

# Unified theme switcher for tmux, ghostty, alacritty, wezterm and nvim.
# Manual modes only: dark (default) / light / toggle / status.

# Resolve dotfiles root: prefer $DOTFILES, fall back to script location.
if [[ -n "${DOTFILES:-}" ]]; then
  DOTFILES_DIR="$DOTFILES"
else
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  DOTFILES_DIR="$(dirname "$SCRIPT_DIR")"
fi

THEME_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/theme"
MODE_FILE="$THEME_DIR/mode"
GHOSTTY_THEME_LINK="$THEME_DIR/ghostty-theme.conf"

GHOSTTY_DARK="$DOTFILES_DIR/configs/ghostty/theme-dark.conf"
GHOSTTY_LIGHT="$DOTFILES_DIR/configs/ghostty/theme-light.conf"

ALACRITTY_THEME_LINK="$THEME_DIR/alacritty-theme.toml"
ALACRITTY_DARK="$DOTFILES_DIR/configs/alacritty/themes/tokyonight-moon.toml"
ALACRITTY_LIGHT="$DOTFILES_DIR/configs/alacritty/themes/tokyonight-day.toml"

TMUX_DARK="$DOTFILES_DIR/configs/tmux/tokyonight-moon.conf"
TMUX_LIGHT="$DOTFILES_DIR/configs/tmux/tokyonight-day.conf"

P10K_DARK="$DOTFILES_DIR/configs/p10k/p10k-dark.zsh"
P10K_LIGHT="$DOTFILES_DIR/configs/p10k/p10k-light.zsh"
P10K_LINK="$HOME/.p10k.zsh"

# Tmux reads theme files through source-file; using a regular file instead of a
# symlink avoids stale symlink resolution/caching issues when switching themes.
TMUX_THEME_FILE="$THEME_DIR/tmux-theme.conf"

usage() {
  echo "Usage: $(basename "$0") {dark|light|toggle|status}"
  exit 1
}

ensure_theme_dir() {
  mkdir -p "$THEME_DIR"
}

read_mode() {
  if [[ -f "$MODE_FILE" ]]; then
    cat "$MODE_FILE"
  else
    echo "dark"
  fi
}

write_mode() {
  printf '%s' "$1" >"$MODE_FILE"
}

update_symlink() {
  local link="$1"
  local target="$2"
  ln -sfn "$target" "$link"
}

copy_theme_file() {
  local source="$1"
  local dest="$2"
  # Overwrite in place so that tmux sees the updated content on the next
  # source-file without stale inode/path caching issues.
  cat "$source" >"$dest"
}

reload_ghostty() {
  if command -v killall >/dev/null 2>&1; then
    killall -SIGUSR2 ghostty 2>/dev/null || true
  fi
}

reload_tmux() {
  if command -v tmux >/dev/null 2>&1 && tmux info >/dev/null 2>&1; then
    # Source only the theme file: it updates tmux user options (@bg, @fg, ...)
    # and tmux redraws all format strings that reference them automatically.
    tmux source-file "${HOME}/.config/theme/tmux-theme.conf" 2>/dev/null || true
    # Force status bar redraw so the new colors appear immediately.
    tmux refresh-client -S 2>/dev/null || true
  fi
}

reload_wezterm() {
  # WezTerm watches its config file for changes. Touching the symlink updates its
  # mtime and triggers a config reload, so the new theme is picked up.
  if [[ -L "$HOME/.wezterm.lua" || -f "$HOME/.wezterm.lua" ]]; then
    touch "$HOME/.wezterm.lua" 2>/dev/null || true
  fi
}

apply_theme() {
  local mode="$1"

  ensure_theme_dir
  write_mode "$mode"

  case "$mode" in
  dark)
    update_symlink "$GHOSTTY_THEME_LINK" "$GHOSTTY_DARK"
    update_symlink "$ALACRITTY_THEME_LINK" "$ALACRITTY_DARK"
    update_symlink "$P10K_LINK" "$P10K_DARK"
    copy_theme_file "$TMUX_DARK" "$TMUX_THEME_FILE"
    ;;
  light)
    update_symlink "$GHOSTTY_THEME_LINK" "$GHOSTTY_LIGHT"
    update_symlink "$ALACRITTY_THEME_LINK" "$ALACRITTY_LIGHT"
    update_symlink "$P10K_LINK" "$P10K_LIGHT"
    copy_theme_file "$TMUX_LIGHT" "$TMUX_THEME_FILE"
    ;;
  *)
    echo "Unknown mode: $mode" >&2
    exit 1
    ;;
  esac

  reload_ghostty
  reload_tmux
  reload_wezterm
}

main() {
  local cmd="${1:-}"
  [[ -z "$cmd" ]] && usage

  case "$cmd" in
  dark | light)
    apply_theme "$cmd"
    echo "Theme set to $cmd"
    ;;
  toggle)
    local current
    current="$(read_mode)"
    if [[ "$current" == "dark" ]]; then
      apply_theme "light"
      # echo "Theme set to light"
    else
      apply_theme "dark"
      # echo "Theme set to dark"
    fi
    ;;
  status)
    read_mode
    echo
    ;;
  -h | --help | help)
    usage
    ;;
  *)
    usage
    ;;
  esac
}

main "$@"
