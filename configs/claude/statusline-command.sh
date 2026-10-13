#!/bin/bash
input=$(cat)

model=$(printf '%s' "$input" | jq -r '.model.display_name // empty' 2>/dev/null)
effort=$(printf '%s' "$input" | jq -r '.effort.level // empty' 2>/dev/null)
used=$(printf '%s' "$input" | jq -r '.context_window.used_percentage // empty' 2>/dev/null)
five=$(printf '%s' "$input" | jq -r '.rate_limits.five_hour.used_percentage // empty' 2>/dev/null)
week=$(printf '%s' "$input" | jq -r '.rate_limits.seven_day.used_percentage // empty' 2>/dev/null)

parts=()

first="$model"
[ -n "$first" ] && [ -n "$effort" ] && first="$first $effort"
[ -n "$first" ] && parts+=("$first")

[ -n "$used" ] && parts+=("$(printf 'Context %.0f%% used' "$used")")

if [ -n "$five" ]; then
  left=$(awk -v u="$five" 'BEGIN { l = 100 - u; if (l < 0) l = 0; printf "%.0f", l }')
  parts+=("5h ${left}% left")
fi

if [ -n "$week" ]; then
  left=$(awk -v u="$week" 'BEGIN { l = 100 - u; if (l < 0) l = 0; printf "%.0f", l }')
  parts+=("weekly ${left}% left")
fi

out=""
for p in "${parts[@]}"; do
  [ -z "$out" ] && out="$p" || out="$out · $p"
done
printf '%s' "$out"
