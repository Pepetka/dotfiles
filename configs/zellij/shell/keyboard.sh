#!/bin/bash
current_layout=$(defaults read ~/Library/Preferences/com.apple.HIToolbox.plist AppleSelectedInputSources | egrep -w 'KeyboardLayout Name' |sed -E 's/^.+ = \"?([^\"]+)\"?;$/\1/')

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
