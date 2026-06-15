alias e="exit"
alias cls="clear"

alias v="$NVIM/11.7/bin/nvim"
alias vv="NVIM_APPNAME=nvim-new nvim"

alias zsho="nvim ~/.zshrc"
alias zshr="source ~/.zshrc"

alias ohmyzsh="nvim ~/.oh-my-zsh"
alias alacritty.conf="nvim ~/.config/alacritty/alacritty.toml"
alias wezterm.conf="nvim ~/.wezterm.lua"
alias tmux.conf="nvim ~/.tmux.conf"
alias aliases="nvim $ZSH/aliases/aliases.zsh"
alias zshenv="nvim $ZSH/zshenv/zshenv.zsh"

alias ls="eza --tree --level=1 --icons=always --no-time --no-user --no-permissions -a"
alias ls2="eza --tree --level=2 --icons=always --no-time --no-user --no-permissions -a"

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
alias gcom="git checkout master"
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
alias gcf="git clean -fd"
alias gf="git fetch"
alias got="git"
alias get="git"

alias renum="tmux move-window -r"

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

unalias gcom 2>/dev/null
function gcom() {
  git checkout master 2>/dev/null || git checkout main
}

alias nvmc="nvm current"
alias nvml="nvm list"

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
