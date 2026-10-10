#!/bin/bash
command -v defaults >/dev/null 2>&1 || exit 0
current_layout=$(defaults read "$HOME/Library/Preferences/com.apple.HIToolbox.plist" AppleSelectedInputSources 2>/dev/null | awk '/KeyboardLayout Name/' | sed -E 's/^.+ = \"?([^\"]+)\"?;$/\1/')

case "$current_layout" in
"Russian")
  echo "RU"
  ;;
"ABC")
  echo "EN"
  ;;
*)
  echo "$current_layout"
  ;;
esac
