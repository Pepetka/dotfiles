if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then
  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"
fi

export ZSH="$HOME/.oh-my-zsh"
export DOTFILES="$HOME/workspace/dotfiles"

export PATH="/usr/local/bin:$PATH"
export PATH="$PATH:/Applications/WezTerm.app/Contents/MacOS"
export PATH="/opt/homebrew/opt/libpq/bin:$PATH"
export PATH="/opt/homebrew/opt/curl/bin:$PATH"
export DYLD_FALLBACK_LIBRARY_PATH="$(brew --prefix)/lib:$DYLD_FALLBACK_LIBRARY_PATH"

export NVIM="$HOME/.config/nvim"
export NVIM_APPNAME="nvim-new"

if [[ -n $SSH_CONNECTION ]]; then
  export EDITOR='vim'
else
  export EDITOR='nvim'
fi

ZSH_THEME="powerlevel10k/powerlevel10k"

DISABLE_LS_COLORS="true"
DISABLE_AUTO_TITLE="true"

if [[ -n "$GHOSTTY_RESOURCES_DIR" ]]; then
  export SNACKS_GHOSTTY=true
fi

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
source $ZSH/aliases/aliases.zsh
source $ZSH/zshenv/zshenv.zsh

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
