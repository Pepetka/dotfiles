if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then
  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"
fi

# Resolve the real config file, including ~/.zshrc symlinks. Allow an explicit override.
export DOTFILES="${DOTFILES:-${${(%):-%x}:A:h:h}}"

# Device-specific paths and settings stay outside the repository.
[[ ! -r "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/local.zsh" ]] || source "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/local.zsh"
export ZSH="${ZSH:-$HOME/.oh-my-zsh}"
export ZSH_CUSTOM="${ZSH_CUSTOM:-$ZSH/custom}"

typeset -U path fpath

# Add custom completions
fpath+=("$DOTFILES/shell/completions")

() {
  local brew_executable="${commands[brew]:-}" brew_prefix candidate
  if [[ -n "$HOMEBREW_PREFIX" && -x "$HOMEBREW_PREFIX/bin/brew" ]]; then
    brew_executable="$HOMEBREW_PREFIX/bin/brew"
  fi
  # Discover Homebrew before adding its prefix; these are standard installation locations.
  if [[ -z "$brew_executable" ]]; then
    for candidate in /opt/homebrew/bin/brew /usr/local/bin/brew "$HOME/.linuxbrew/bin/brew" /home/linuxbrew/.linuxbrew/bin/brew; do
      if [[ -x "$candidate" ]]; then
        brew_executable="$candidate"
        break
      fi
    done
  fi
  if [[ -n "$brew_executable" ]] && brew_prefix="$("$brew_executable" --prefix 2>/dev/null)" && [[ -d "$brew_prefix" ]]; then
    for candidate in "$brew_prefix/sbin" "$brew_prefix/bin" "$brew_prefix/opt/libpq/bin" "$brew_prefix/opt/curl/bin"; do
      [[ ! -d "$candidate" ]] || path=("$candidate" $path)
    done
    if [[ "$OSTYPE" == darwin* && -d "$brew_prefix/lib" ]]; then
      local -Ua library_paths
      library_paths=("$brew_prefix/lib" ${(s.:.)DYLD_FALLBACK_LIBRARY_PATH})
      export DYLD_FALLBACK_LIBRARY_PATH="${(j.:.)library_paths}"
    fi
  fi
  local go_path
  if (($+commands[go])) && go_path="$(go env GOPATH 2>/dev/null)"; then
    for candidate in ${(s.:.)go_path}; do
      [[ ! -d "$candidate/bin" ]] || path+=("$candidate/bin")
    done
  fi
  for candidate in "$HOME/.local/bin" "$HOME/bin"; do
    [[ ! -d "$candidate" ]] || path=("$candidate" $path)
  done
  if [[ "$OSTYPE" == darwin* ]]; then
    for candidate in "$HOME/Applications/WezTerm.app/Contents/MacOS" /Applications/WezTerm.app/Contents/MacOS; do
      [[ ! -d "$candidate" ]] || path+=("$candidate")
    done
  fi
}

if [[ -n $SSH_CONNECTION ]]; then
  export EDITOR='vim'
else
  export EDITOR='nvim'
fi

ZSH_THEME="${ZSH_THEME-powerlevel10k/powerlevel10k}"

DISABLE_LS_COLORS="true"
DISABLE_AUTO_TITLE="true"

plugins=(
  z
  git
  fzf
  nvm
  npm
  bun
  sudo
  docker
  direnv
  extract
  alias-tips
  autoupdate
  web-search
  docker-compose
  git-auto-fetch
  zsh-autosuggestions
  zsh-syntax-highlighting
)

if [[ -r "$ZSH/oh-my-zsh.sh" ]]; then
  () {
    local plugin
    local -a available_plugins=()
    for plugin in $plugins; do
      if [[ -f "$ZSH_CUSTOM/plugins/$plugin/$plugin.plugin.zsh" || -f "$ZSH_CUSTOM/plugins/$plugin/_$plugin" ||
        -f "$ZSH/plugins/$plugin/$plugin.plugin.zsh" || -f "$ZSH/plugins/$plugin/_$plugin" ]]; then
        available_plugins+=("$plugin")
      fi
    done
    plugins=("${available_plugins[@]}")
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
source "$DOTFILES/configs/aliases.zsh"

# Private API keys and secrets. Keep this file out of dotfiles / git.
[[ ! -r "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/secrets.zsh" ]] || source "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/secrets.zsh"
WORDCHARS=${WORDCHARS//:/}

function _scripts_completion() {
  local prev_word=${words[2]}

  if [[ "$prev_word" != "run" && "$prev_word" != "run-script" ]]; then
    return 1
  fi

  if [[ -f package.json ]]; then
    local scripts
    scripts=($(jq -r '.scripts | keys[]' package.json 2>/dev/null))
    compadd -a scripts
  fi
}

compdef _scripts_completion pnpm

if (($+functions[p10k])) && [[ -r "$HOME/.p10k.zsh" ]]; then
  source "$HOME/.p10k.zsh"
fi

export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
if ((! $+functions[nvm])) && [[ -s "$NVM_DIR/nvm.sh" ]]; then
  source "$NVM_DIR/nvm.sh"
fi
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"

export BUN_INSTALL="${BUN_INSTALL:-$HOME/.bun}"
[[ ! -d "$BUN_INSTALL/bin" ]] || path=("$BUN_INSTALL/bin" $path)

[[ ! -s "$BUN_INSTALL/_bun" ]] || source "$BUN_INSTALL/_bun"

[[ ! -r "$HOME/.local/bin/env" ]] || source "$HOME/.local/bin/env"

export KIMI_CODE_HOME="${KIMI_CODE_HOME:-$HOME/.kimi-code}"
[[ ! -d "$KIMI_CODE_HOME/bin" ]] || path=("$KIMI_CODE_HOME/bin" $path)

# SDKMAN must stay at the end of the file.
export SDKMAN_DIR="${SDKMAN_DIR:-$HOME/.sdkman}"
[[ ! -s "$SDKMAN_DIR/bin/sdkman-init.sh" ]] || source "$SDKMAN_DIR/bin/sdkman-init.sh"
