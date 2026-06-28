# Basics
alias e="exit"
alias cls="clear"
alias v="nvim"

# Configs
alias zsho="nvim ~/.zshrc"
alias zshr="source ~/.zshrc"
alias ohmyzsh="nvim ~/.oh-my-zsh"
alias alacritty.conf="nvim ~/.config/alacritty/alacritty.toml"
alias wezterm.conf="nvim ~/.wezterm.lua"
alias ghostty.conf="nvim ~/.config/ghostty/config"
alias tmux.conf="nvim ~/.tmux.conf"
alias aliases="nvim $DOTFILES/configs/aliases.zsh"
alias zshenv="nvim $ZSH/zshenv/zshenv.zsh"

unalias theme 2>/dev/null
theme() {
  "$DOTFILES/shell/theme-switch.sh" "$@"
  [[ ! -f ~/.p10k.zsh ]] || source ~/.p10k.zsh
  (($+functions[p10k])) && p10k reload 2>/dev/null
  (($+functions[_p9k_precmd])) && _p9k_precmd 2>/dev/null
  [[ -o zle ]] && zle .reset-prompt 2>/dev/null
}

# Listing
alias ls="eza --tree --level=1 --icons=always --no-time --no-user --no-permissions -a"
alias ls2="eza --tree --level=2 --icons=always --no-time --no-user --no-permissions -a"

# Git
alias gs="git status"
alias ga="git add"
alias gaa="git add ."
alias gas="git add ./src"
alias gpl="git pull"
alias gpld="git pull develop"
alias gplm="git pull master"
alias gps="git push"
alias gc="git commit -m"
alias gcf="git commit --amend --no-edit"
alias gco="git checkout"
alias gcob="git checkout -b"
alias gcod="git checkout develop"
alias gb="git branch"
alias gm="git merge"
alias gha="git log --pretty=format:\"%h %ad | %s%d [%an]\" --graph --date=short"
alias gst="git stash"
alias gstl="git stash list"
alias gsta="git stash apply"
alias gstp="git stash pop"
alias gstc="git stash clear"
alias gt="git tag"
alias grest="git restore"
alias grests="git restore --staged"
alias greset="git reset --hard"
alias greseth="git reset HEAD^"
alias gclean="git clean -fd"
alias gf="git fetch"
alias got="git"
alias get="git"

unalias gcom 2>/dev/null
function gcom() {
  git checkout master 2>/dev/null || git checkout main
}

# Tmux
alias renum="tmux move-window -r"

# AI providers
alias deepseek='export ANTHROPIC_BASE_URL=https://api.deepseek.com/anthropic \
  export ANTHROPIC_AUTH_TOKEN=$DEEPSEEK_API_KEY \
  export ANTHROPIC_DEFAULT_HAIKU_MODEL=deepseek-chat \
  export ANTHROPIC_DEFAULT_SONNET_MODEL=deepseek-chat \
  export ANTHROPIC_DEFAULT_OPUS_MODEL=deepseek-reasoner \
  export API_TIMEOUT_MS=300000 \
  export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1 && claude'

alias zai='export ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic \
  export ANTHROPIC_AUTH_TOKEN=$Z_AI_API_KEY \
  export ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-4.7 \
  export ANTHROPIC_DEFAULT_SONNET_MODEL=glm-5 \
  export ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.1 \
  export API_TIMEOUT_MS=300000 \
  export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1 && claude'

alias minimax='export ANTHROPIC_BASE_URL=https://api.minimax.io/anthropic \
  export ANTHROPIC_AUTH_TOKEN=$MINIMAX_API_KEY \
  export ANTHROPIC_DEFAULT_HAIKU_MODEL=MiniMax-M2.7 \
  export ANTHROPIC_DEFAULT_SONNET_MODEL=MiniMax-M2.7 \
  export ANTHROPIC_DEFAULT_OPUS_MODEL=MiniMax-M3 \
  export API_TIMEOUT_MS=3000000 \
  export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1 && claude'

alias kimic='export ANTHROPIC_BASE_URL=https://api.kimi.com/coding \
  export ANTHROPIC_AUTH_TOKEN=$KIMI_API_KEY \
  export ANTHROPIC_DEFAULT_HAIKU_MODEL=kimi-k2.5 \
  export ANTHROPIC_DEFAULT_SONNET_MODEL=kimi-k2.5 \
  export ANTHROPIC_DEFAULT_OPUS_MODEL=K2.6 \
  export ENABLE_TOOL_SEARCH=false \
  export API_TIMEOUT_MS=300000 \
  export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1 && claude'

# Node
alias nvmc="nvm current"
alias nvml="nvm list"

# Functions
function yy() {
  local tmp="$(mktemp -t "yazi-cwd.XXXXXX")"
  yazi "$@" --cwd-file="$tmp"
  if cwd="$(cat -- "$tmp")" && [ -n "$cwd" ] && [ "$cwd" != "$PWD" ]; then
    cd -- "$cwd"
  fi
  rm -f -- "$tmp"
}

function hurl_pretty() {
  if [[ -n "$1" ]]; then
    hurl api.hurl | jq -r "$1" | jq | bat --language=json
  else
    hurl api.hurl | jq | bat --language=json
  fi
}

function ts3() {
  if [ -z "$TMUX" ]; then
    echo "Error: not inside tmux" >&2
    return 1
  fi

  local panes
  panes=$(tmux display-message -p '#{window_panes}')

  if [ "$panes" -ne 1 ]; then
    echo "Current window already has $panes panes, skipping split" >&2
    return 0
  fi

  local cmd1="${1:-nvim}"
  local cmd2="${2:-}"
  local cmd3="${3:-}"

  local left
  left=$(tmux display-message -p '#{pane_id}')
  local top_right
  top_right=$(tmux split-window -h -p 34 -P -F '#{pane_id}')
  local bottom_right
  bottom_right=$(tmux split-window -v -p 16 -P -F '#{pane_id}')

  [ -n "$cmd1" ] && tmux send-keys -t "$left" "$cmd1" C-m
  [ -n "$cmd2" ] && tmux send-keys -t "$top_right" "$cmd2" C-m
  [ -n "$cmd3" ] && tmux send-keys -t "$bottom_right" "$cmd3" C-m

  tmux select-pane -t "$bottom_right"
}
