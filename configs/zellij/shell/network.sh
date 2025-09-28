#!/bin/sh

GREEN='\033[0;32m'
RED='\033[0;31m'

HOST="8.8.8.8"

format_network_output() {
    local icon=$1
    local color=$2

    printf " #[fg=\$%s]#[bg=\$%s,fg=\$crust]%s " "$color" "$color" "$icon"
}

if ping -c 1 -W 1000 "$HOST" &> /dev/null; then
    format_network_output "󰖩 " "green"
else
    format_network_output "󰖪 " "red"
fi
