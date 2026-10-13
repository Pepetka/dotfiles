if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then
  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"
fi

# Resolve the checkout through the installed .zshrc symlink.
export DOTFILES="${DOTFILES:-${${(%):-%x}:A:h:h:h}}"
typeset -U path fpath

# Paths, SDK initialization and machine preferences belong here.
[[ ! -r "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/local.zsh" ]] || source "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/local.zsh"
[[ ! -r "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/secrets.zsh" ]] || source "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/secrets.zsh"

if [[ -z "$EDITOR" ]]; then
  if (( $+commands[nvim] )); then
    export EDITOR=nvim
  elif (( $+commands[vim] )); then
    export EDITOR=vim
  else
    export EDITOR=vi
  fi
fi
export ZSH="${ZSH:-$HOME/.oh-my-zsh}"
export ZSH_CUSTOM="${ZSH_CUSTOM:-$ZSH/custom}"
ZSH_THEME="${ZSH_THEME-powerlevel10k/powerlevel10k}"
# eza replaces ls; without it keep Oh My Zsh's colored ls.
(( ! $+commands[eza] )) || DISABLE_LS_COLORS=true
DISABLE_AUTO_TITLE=true
# local.zsh may supply its own plugin list, including an empty one.
(( ${+plugins} )) || plugins=(z sudo zsh-autosuggestions zsh-syntax-highlighting)
fpath+=("$DOTFILES/theme/completions")

if [[ -r "$ZSH/oh-my-zsh.sh" ]]; then
  () {
    local plugin
    local -a available=()
    for plugin in $plugins; do
      if [[ -f "$ZSH_CUSTOM/plugins/$plugin/$plugin.plugin.zsh" || -f "$ZSH_CUSTOM/plugins/$plugin/_$plugin" ||
        -f "$ZSH/plugins/$plugin/$plugin.plugin.zsh" || -f "$ZSH/plugins/$plugin/_$plugin" ]]; then
        available+=("$plugin")
      fi
    done
    plugins=("${available[@]}")
  }
  if [[ -n "$ZSH_THEME" && ! -r "$ZSH_CUSTOM/$ZSH_THEME.zsh-theme" &&
    ! -r "$ZSH_CUSTOM/themes/$ZSH_THEME.zsh-theme" && ! -r "$ZSH/themes/$ZSH_THEME.zsh-theme" ]]; then
    ZSH_THEME=""
  fi
  source "$ZSH/oh-my-zsh.sh"
else
  autoload -Uz compinit
  compinit
fi

source "$DOTFILES/configs/zsh/aliases.zsh"
source "$DOTFILES/configs/zsh/functions.zsh"
WORDCHARS=${WORDCHARS//:/}

# Initialization that needs completion (compdef) runs after the framework.
[[ ! -r "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/local.after.zsh" ]] || source "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/local.after.zsh"

# Load the shared ANSI prompt directly from its canonical source.
if (( $+functions[p10k] )); then
  source "$DOTFILES/theme/resources/p10k/p10k.zsh"
fi
