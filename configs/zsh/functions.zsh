unalias gcom 2>/dev/null
function gcom() {
  git checkout master 2>/dev/null || git checkout main
}

function yy() {
  local cwd tmp
  tmp="$(mktemp "${TMPDIR:-/tmp}/yazi-cwd.XXXXXX")" || return
  yazi "$@" --cwd-file="$tmp"
  if cwd="$(cat -- "$tmp")" && [ -n "$cwd" ] && [ "$cwd" != "$PWD" ]; then
    cd -- "$cwd"
  fi
  rm -f -- "$tmp"
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
