if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then
  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"
fi

export ZSH="$HOME/.oh-my-zsh"
export DOTFILES="$HOME/workspace/dotfiles"

# Add custom completions
fpath+=("$DOTFILES/shell/completions")

export PATH="/usr/local/bin:$PATH"
export PATH="$PATH:/Applications/WezTerm.app/Contents/MacOS"
export PATH="/opt/homebrew/opt/libpq/bin:$PATH"
export PATH="/opt/homebrew/opt/curl/bin:$PATH"
export DYLD_FALLBACK_LIBRARY_PATH="$(brew --prefix)/lib:$DYLD_FALLBACK_LIBRARY_PATH"

if [[ -n $SSH_CONNECTION ]]; then
  export EDITOR='vim'
else
  export EDITOR='nvim'
fi

ZSH_THEME="powerlevel10k/powerlevel10k"

DISABLE_LS_COLORS="true"
DISABLE_AUTO_TITLE="true"

# if [[ -n "$GHOSTTY_RESOURCES_DIR" ]]; then
#   export SNACKS_GHOSTTY=true
# fi

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

source $ZSH/oh-my-zsh.sh
source $DOTFILES/configs/aliases.zsh

# Private API keys and secrets. Keep this file out of dotfiles / git.
[[ ! -f ~/.config/zsh/secrets.zsh ]] || source ~/.config/zsh/secrets.zsh

autoload -Uz compinit
compinit
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

[[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh

# Auto-reload p10k when the ~/.p10k.zsh symlink target changes.
autoload -Uz add-zsh-hook
_p10k_config_target() { readlink ~/.p10k.zsh 2>/dev/null || echo ""; }
_p10k_last_target="$(_p10k_config_target)"
_p10k_check_reload() {
  local current="$(_p10k_config_target)"
  if [[ "$current" != "$_p10k_last_target" ]]; then
    _p10k_last_target="$current"
    [[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh
    (($+functions[p10k])) && p10k reload 2>/dev/null
  fi
}
add-zsh-hook precmd _p10k_check_reload

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"

export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"

[ -s "/Users/vlvkuznetsov/.bun/_bun" ] && source "/Users/vlvkuznetsov/.bun/_bun"

. "$HOME/.local/bin/env"

export PATH="/Users/vlvkuznetsov/.kimi-code/bin:$PATH"

# SDKMAN must stay at the end of the file.
export SDKMAN_DIR="$HOME/.sdkman"
[[ -s "$HOME/.sdkman/bin/sdkman-init.sh" ]] && source "$HOME/.sdkman/bin/sdkman-init.sh"
