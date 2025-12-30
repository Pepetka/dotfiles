#!/bin/bash

format_result() {
  local branch=$1

  printf " #[fg=\$yellow]#[bg=\$yellow,fg=\$crust] #[bg=\$yellow,fg=\$crust]$branch#[fg=\$yellow]"
}

if git rev-parse --git-dir > /dev/null 2>&1; then
  branch=$(git branch --show-current)
  format_result "$branch"
fi
