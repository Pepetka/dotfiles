#!/bin/bash
current_layout=$(setxkbmap -query | awk '/layout:/ {print $2}')

case "$current_layout" in
    "us")
        echo "EN"
        ;;
    "ru")
        echo "RU"
        ;;
    *)
        echo "$current_layout"
        ;;
esac
