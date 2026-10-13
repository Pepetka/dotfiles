#!/bin/sh
# Links dotfiles skills into ~/.agents/skills (Codex), then every skill there into ~/.claude/skills.
set -eu

source_dir=$(CDPATH= cd -- "$(dirname -- "$0")/skills" && pwd)
agents_dir="$HOME/.agents/skills"
claude_dir="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/skills"
mkdir -p "$agents_dir" "$claude_dir"

link() {
  target=$1 path=$2
  if [ -L "$path" ]; then
    [ "$path" -ef "$target" ] && return
    rm "$path"
  elif [ -e "$path" ]; then
    echo "Skipped $path: not a symlink" >&2
    return
  fi
  ln -s "$target" "$path"
  echo "Linked $path -> $target"
}

# Removes dangling links that point into a managed location; anything else is left alone.
prune() {
  dir=$1 managed=$2
  for entry in "$dir"/*; do
    [ -L "$entry" ] && [ ! -e "$entry" ] || continue
    case $(readlink "$entry") in
      "$managed"/* | ../../.agents/skills/*)
        rm "$entry"
        echo "Removed dangling $entry"
        ;;
    esac
  done
}

for skill in "$source_dir"/*; do
  [ -f "$skill/SKILL.md" ] && link "$skill" "$agents_dir/${skill##*/}"
done
prune "$agents_dir" "$source_dir"

prune "$claude_dir" "$agents_dir"
for skill in "$agents_dir"/*; do
  [ -f "$skill/SKILL.md" ] && link "$agents_dir/${skill##*/}" "$claude_dir/${skill##*/}"
done
exit 0
