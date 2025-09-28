alias e="exit"
alias cls="clear"

alias v="nvim"

alias zsho="nvim ~/.zshrc"
alias zshr="source ~/.zshrc"

alias ohmyzsh="nvim ~/.oh-my-zsh"
alias alacritty.conf="nvim ~/.config/alacritty/alacritty.toml"
alias wezterm.conf="nvim ~/.wezterm.lua"
alias tmux.conf="nvim ~/.tmux.conf"
alias aliases="nvim $ZSH/aliases/aliases.zsh"

alias ls="eza --tree --level=1 --icons=always --no-time --no-user --no-permissions -a"
alias ls2="eza --tree --level=2 --icons=always --no-time --no-user --no-permissions -a"

alias gs="git status"
alias ga="git add"
alias gaa="git add ."
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
alias gh="git log --pretty=format:\"%h %ad | %s%d [%an]\" --graph --date=short"
alias gha="git log --pretty=format:\"%h %ad | %s%d [%an]\" --graph --date=short --all"
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

function curl() {
  domain=$(echo "$*" | awk -F'[/:]' '{print $4}')
  domain=$(echo "$domain" | tr -d '\r\n ')

  if [[ "$domain" == "api.openai.com" ]]; then
    command curl --proxy socks5h://127.0.0.1:1086 "$@"
  else
    command curl "$@"
  fi
}
