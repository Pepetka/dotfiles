# Basics
alias e="exit"
alias cls="clear"
alias v="nvim"

# Configs
alias zsho='$EDITOR "${ZDOTDIR:-$HOME}/.zshrc"'
alias zshr='source "${ZDOTDIR:-$HOME}/.zshrc"'
alias ohmyzsh='$EDITOR "${ZSH:-$HOME/.oh-my-zsh}"'
alias alacritty.conf='$EDITOR "${XDG_CONFIG_HOME:-$HOME/.config}/alacritty/alacritty.toml"'
alias wezterm.conf='$EDITOR "${WEZTERM_CONFIG_FILE:-$HOME/.wezterm.lua}"'
alias ghostty.conf='$EDITOR "${XDG_CONFIG_HOME:-$HOME/.config}/ghostty/config"'
alias tmux.conf='$EDITOR ~/.tmux.conf'
alias aliases='$EDITOR "$DOTFILES/configs/zsh/aliases.zsh"'
alias zshlocal='$EDITOR "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/local.zsh"'
alias zshsecrets='$EDITOR "${XDG_CONFIG_HOME:-$HOME/.config}/zsh/secrets.zsh"'

unalias theme 2>/dev/null
theme() {
  "$DOTFILES/theme/theme" "$@"
}

# Listing: keep the system ls when eza is not installed.
if (($+commands[eza])); then
  alias ls="eza --tree --level=1 --icons=always --no-time --no-user --no-permissions -a"
  alias ls2="eza --tree --level=2 --icons=always --no-time --no-user --no-permissions -a"
fi

# Git
alias gs="git status"
alias ga="git add"
alias gaa="git add ."
alias gpl="git pull"
alias gpld="git pull origin develop"
alias gplm="git pull origin master"
alias gps="git push"
alias gpf="git push --force-with-lease --force-if-includes"
alias 'gpf!'="git push --force"
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

# Tmux
alias renum="tmux move-window -r"

# Node
alias nvmc="nvm current"
alias nvml="nvm list"
