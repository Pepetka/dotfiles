if [[ -r "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh" ]]; then
  source "${XDG_CACHE_HOME:-$HOME/.cache}/p10k-instant-prompt-${(%):-%n}.zsh"
fi

export ZSH="$HOME/.oh-my-zsh"
export PATH="/usr/local/bin:$PATH"
export PATH="$PATH:/Applications/WezTerm.app/Contents/MacOS"
export DYLD_FALLBACK_LIBRARY_PATH="$(brew --prefix)/lib:$DYLD_FALLBACK_LIBRARY_PATH"

if [[ -n $SSH_CONNECTION ]]; then
  export EDITOR='vim'
else
  export EDITOR='nvim'
fi

ZSH_THEME="powerlevel10k/powerlevel10k"

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

# if [[ $TERM != "screen" ]] && [[ $TERM != "screen-256color" ]]; then
#   if command -v tmux &> /dev/null; then
#     [ -z "$TMUX" ] && exec tmux
#   fi
# fi

autoload -Uz compinit
compinit
WORDCHARS=${WORDCHARS//:/}

function _scripts_completion() {
  local prev_word
  prev_word=${words[2]}

  # Срабатываем только если это 'run' или 'run-script'
  if [[ "$prev_word" != "run" && "$prev_word" != "run-script" ]]; then
    return 1
  fi

  # Собираем список скриптов
  if [[ -f package.json ]]; then
    local scripts
    scripts=($(jq -r '.scripts | keys[]' package.json 2>/dev/null))
    compadd -a scripts
  fi
}

compdef _scripts_completion pnpm

# To customize prompt, run `p10k configure` or edit ~/.p10k.zsh.
[[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion

# bun completions
[ -s "/Users/vlvkuznetsov/.bun/_bun" ] && source "/Users/vlvkuznetsov/.bun/_bun"

# bun
export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"
