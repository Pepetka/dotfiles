#!/bin/bash
# Print the active keyboard layout (EN, RU, ...) for the tmux status bar.
# Prints nothing when the system exposes no layout, e.g. over SSH or headless.

# Short codes from xkb layouts such as "us", "ru" or "de+nodeadkeys".
xkb_label() {
  case "$1" in
  "") ;;
  us* | gb*) echo EN ;;
  *)
    local code="${1%%[+(:]*}"
    printf '%s\n' "$code" | tr '[:lower:]' '[:upper:]'
    ;;
  esac
}

macos_layout() {
  command -v defaults >/dev/null 2>&1 || return 0
  local layout
  layout=$(defaults read "$HOME/Library/Preferences/com.apple.HIToolbox.plist" AppleSelectedInputSources 2>/dev/null | awk '/KeyboardLayout Name/' | sed -E 's/^.+ = \"?([^\"]+)\"?;$/\1/')
  # Names such as "ABC", "U.S.", "RussianWin" or "Ukrainian-PC" map to the codes used on Linux.
  case "$layout" in
  "") ;;
  ABC* | U.S.* | US* | British* | Australian | Canadian) echo EN ;;
  Russian*) echo RU ;;
  Ukrainian*) echo UA ;;
  *) echo "$layout" ;;
  esac
}

# GNOME keeps the active input source first in mru-sources: [('xkb', 'us'), ...].
gnome_layout() {
  command -v gsettings >/dev/null 2>&1 || return 1
  local key sources
  for key in mru-sources sources; do
    sources=$(gsettings get org.gnome.desktop.input-sources "$key" 2>/dev/null) || return 1
    if [[ "$sources" =~ \(\'[^\']*\',\ \'([^\']+)\'\) ]]; then
      xkb_label "${BASH_REMATCH[1]}"
      return 0
    fi
  done
  return 1
}

# KDE: getLayout returns the active index into getLayoutsList's (layout, variant, name) tuples.
kde_layout() {
  [[ "${XDG_CURRENT_DESKTOP:-}" == *KDE* ]] && command -v gdbus >/dev/null 2>&1 || return 1
  local call=(gdbus call --session --dest org.kde.keyboard --object-path /Layouts --method)
  local index layouts
  index=$("${call[@]}" org.kde.KeyboardLayouts.getLayout 2>/dev/null) || return 1
  layouts=$("${call[@]}" org.kde.KeyboardLayouts.getLayoutsList 2>/dev/null) || return 1
  [[ "$index" =~ uint32\ ([0-9]+) ]] || return 1
  index="${BASH_REMATCH[1]}"
  xkb_label "$(grep -o "('[^']*', '[^']*', '[^']*')" <<<"$layouts" | sed -n "$((index + 1))s/^('\([^']*\)'.*/\1/p")"
}

# X11: the second xkb group sets bit 0x1000 of the LED mask.
x11_layout() {
  [[ -n "${DISPLAY:-}" ]] || return 1
  if command -v xkb-switch >/dev/null 2>&1; then
    xkb_label "$(xkb-switch -p 2>/dev/null)"
    return 0
  fi
  command -v setxkbmap >/dev/null 2>&1 && command -v xset >/dev/null 2>&1 || return 1
  local layouts mask group
  layouts=$(setxkbmap -query 2>/dev/null | awk '/^layout:/ {print $2}')
  mask=$(xset -q 2>/dev/null | awk '/LED mask/ {print $NF}')
  [[ -n "$layouts" && "$mask" =~ ^[0-9a-fA-F]+$ ]] || return 1
  group=$(((16#$mask >> 12) & 1))
  IFS=, read -ra layouts <<<"$layouts"
  xkb_label "${layouts[group]:-${layouts[0]}}"
}

case "$(uname -s)" in
Darwin) macos_layout ;;
Linux)
  [[ -n "${WAYLAND_DISPLAY:-}" || -n "${DISPLAY:-}" ]] || exit 0
  kde_layout || gnome_layout || x11_layout || true
  ;;
esac
