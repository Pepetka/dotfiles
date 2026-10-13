#!/usr/bin/env bash
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
test_directory="$(mktemp -d "${TMPDIR:-/tmp}/dotfiles-portability.XXXXXX")"
test_directory="$(cd "$test_directory" && pwd -P)"
cleanup() {
  local test_status=$?
  if [[ "$test_status" != 0 ]]; then
    for log in "$test_directory/"{startup,override,editor,custom}.log; do
      [[ ! -f "$log" ]] || cat "$log" >&2
    done
  fi
  rm -rf "$test_directory"
  exit "$test_status"
}
trap cleanup EXIT
repository="$test_directory/relocated repo"
test_home="$test_directory/test home"
config_root="$test_directory/config with spaces"
mkdir -p "$repository/theme" "$test_home" "$config_root/zsh" "$test_directory/bin"
cp -R "$source_directory/configs" "$repository/configs"
cp -R "$source_directory/shell" "$repository/shell"
cp -R "$source_directory/theme/completions" "$repository/theme/completions"
cp -R "$source_directory/theme/resources" "$repository/theme/resources"
ln -s "$repository/configs/zsh/.zshrc" "$test_home/.zshrc"
# The shared shell must not probe package managers or language runtimes.
for tool in brew go; do
  cat >"$test_directory/bin/$tool" <<'SH'
#!/bin/sh
: > "$PORTABILITY_TEST_PROBE"
exit 99
SH
done
cat >"$test_directory/bin/nvim" <<'SH'
#!/bin/sh
printf '%s\n' "$@" > "$PORTABILITY_TEST_ARGS"
SH
chmod 700 "$test_directory/bin/"*
zsh_executable="$(command -v zsh)"
run_zsh() {
  env -i HOME="$test_home" XDG_CONFIG_HOME="$config_root" \
    PATH="$test_directory/bin:/usr/bin:/bin" \
    PORTABILITY_TEST_PROBE="$test_directory/probed" \
    PORTABILITY_TEST_REPO="$repository" PORTABILITY_TEST_ARGS="$test_directory/args" \
    PORTABILITY_TEST_BREW_PREFIX="${test_brew_prefix:-}" PORTABILITY_TEST_GOPATH="${test_go_path:-}" \
    "$zsh_executable" -f "$@"
}
# Resolve a relocated repository through ~/.zshrc, tolerate absent tools and optional files.
run_zsh -c '
  source "$HOME/.zshrc"
  [[ "$DOTFILES" == "$PORTABILITY_TEST_REPO" ]] || exit 1
  (( $+functions[theme] && $+functions[compdef] )) || exit 1
  [[ -z "$BUN_INSTALL" && -z "$KIMI_CODE_HOME" && -z "$HOMEBREW_PREFIX" ]] || exit 1
  (( ${fpath[(Ie)$DOTFILES/theme/completions]} )) || exit 1
  (( ! $+aliases[ls] )) || exit 1
  [[ "$EDITOR" == nvim && -z "$DISABLE_LS_COLORS" ]] || exit 1
  (( ! ${plugins[(Ie)git]} )) || exit 1
  source "$HOME/.zshrc"
  typeset -Ua unique_path
  unique_path=("${path[@]}")
  (( $#path == $#unique_path )) || exit 1
  eval alacritty.conf
' >"$test_directory/startup.log" 2>&1
[[ ! -s "$test_directory/startup.log" ]]
[[ "$(cat "$test_directory/args")" == "$config_root/alacritty/alacritty.toml" ]]
mkdir -p "$test_home/alternate dotfiles"
ln -s "$repository/configs" "$test_home/alternate dotfiles/configs"
run_zsh -c '
  export DOTFILES="$HOME/alternate dotfiles"
  source "$HOME/.zshrc"
  [[ "$DOTFILES" == "$HOME/alternate dotfiles" ]] || exit 1
' >"$test_directory/override.log" 2>&1
[[ ! -s "$test_directory/override.log" ]]
# Hosts without Neovim, such as minimal Ubuntu servers, fall back to vim.
# compinit only needs mv to write its dump; /bin may contain a packaged nvim.
mkdir -p "$test_directory/minimal bin"
ln -s "$(command -v mv)" "$test_directory/minimal bin/mv"
run_zsh -c '
  path=("$HOME/../minimal bin")
  source "$HOME/.zshrc"
  [[ "$EDITOR" == vim ]] || exit 1
' >"$test_directory/editor.log" 2>&1
[[ ! -s "$test_directory/editor.log" ]]
test_brew_prefix="$test_directory/custom brew"
test_go_path="$test_directory/go one:$test_directory/go two"
mkdir -p "$test_brew_prefix/"{bin,sbin,lib,opt/curl/bin,opt/libpq/bin} \
  "$test_directory/go one/bin" "$test_directory/go two/bin" \
  "$test_home/custom bun/bin" "$test_home/custom kimi/bin" \
  "$test_home/custom omz/plugins/git" "$test_home/custom plugins/plugins/zsh-autosuggestions" \
  "$test_home/custom plugins/themes/powerlevel10k" "$test_home/custom sdk/bin" "$test_home/.local/bin"
touch "$test_home/custom omz/plugins/git/git.plugin.zsh" \
  "$test_home/custom plugins/plugins/zsh-autosuggestions/zsh-autosuggestions.plugin.zsh" \
  "$test_home/custom plugins/themes/powerlevel10k/powerlevel10k.zsh-theme"
cat >"$config_root/zsh/local.zsh" <<'ZSH'
export ZSH="$HOME/custom omz"
export ZSH_CUSTOM="$HOME/custom plugins"
export BUN_INSTALL="$HOME/custom bun"
export KIMI_CODE_HOME="$HOME/custom kimi"
export SDKMAN_DIR="$HOME/custom sdk"
export HOMEBREW_PREFIX="$PORTABILITY_TEST_BREW_PREFIX"
plugins=(git missing-plugin zsh-autosuggestions)
path=("$HOMEBREW_PREFIX/opt/curl/bin" "$HOMEBREW_PREFIX/opt/libpq/bin" "$HOMEBREW_PREFIX/bin" "$BUN_INSTALL/bin" "$KIMI_CODE_HOME/bin" $path)
for go_root in ${(s.:.)PORTABILITY_TEST_GOPATH}; do
  path+=("$go_root/bin")
done
unset go_root
source "$BUN_INSTALL/_bun"
source "$SDKMAN_DIR/bin/sdkman-init.sh"
source "$HOME/.local/bin/env"
ZSH
cat >"$test_home/custom omz/oh-my-zsh.sh" <<'ZSH'
compdef() { :; }
p10k() { :; }
export PORTABILITY_OMZ=loaded
ZSH
echo 'export PORTABILITY_BUN=loaded' >"$test_home/custom bun/_bun"
echo 'export PORTABILITY_SDK=loaded' >"$test_home/custom sdk/bin/sdkman-init.sh"
echo 'export PORTABILITY_LOCAL_ENV=loaded' >"$test_home/.local/bin/env"
echo 'export PORTABILITY_PROMPT=loaded' >"$repository/theme/resources/p10k/p10k.zsh"
echo 'export PORTABILITY_SECRETS=loaded' >"$config_root/zsh/secrets.zsh"
run_zsh -c '
  source "$HOME/.zshrc"
  [[ "$PORTABILITY_OMZ/$PORTABILITY_BUN/$PORTABILITY_SDK/$PORTABILITY_LOCAL_ENV/$PORTABILITY_PROMPT/$PORTABILITY_SECRETS" == loaded/loaded/loaded/loaded/loaded/loaded ]] || exit 1
  [[ "${(j:,:)plugins}" == git,zsh-autosuggestions && "$ZSH_THEME" == powerlevel10k/powerlevel10k ]] || exit 1
  for expected in "$PORTABILITY_TEST_BREW_PREFIX/bin" "$PORTABILITY_TEST_BREW_PREFIX/opt/libpq/bin" "$PORTABILITY_TEST_BREW_PREFIX/opt/curl/bin" "$HOME/custom bun/bin" "$HOME/custom kimi/bin"; do
    (( ${path[(Ie)$expected]} )) || exit 1
  done
  for expected in ${(s.:.)PORTABILITY_TEST_GOPATH}; do
    expected+="/bin"
    (( ${path[(Ie)$expected]} )) || exit 1
  done
  (( ${path[(Ie)$PORTABILITY_TEST_BREW_PREFIX/opt/curl/bin]} < ${path[(Ie)$PORTABILITY_TEST_BREW_PREFIX/bin]} )) || exit 1
  source "$HOME/.zshrc"
  typeset -Ua unique_path
  unique_path=("${path[@]}")
  (( $#path == $#unique_path )) || exit 1
' >"$test_directory/custom.log" 2>&1
[[ ! -s "$test_directory/custom.log" ]]
[[ ! -e "$test_directory/probed" ]]
echo 'Portable Zsh config passed: relocated repo, symlinks/spaces, XDG, missing tools, explicit local setup, optional plugins and repeated sourcing'

bash "$source_directory/shell/tests/keyboard.sh"
if command -v lua >/dev/null; then
  lua "$source_directory/shell/tests/terminals.lua" "$source_directory" "$config_root"
elif command -v nvim >/dev/null; then
  NVIM_LOG_FILE=/dev/null nvim --headless -u NONE -i NONE -n -l "$source_directory/shell/tests/terminals.lua" "$source_directory" "$config_root"
fi
