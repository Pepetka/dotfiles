#!/bin/bash

get_battery_icon_and_color() {
    local battery_level=$1

    if [[ $battery_level -ge 80 ]]; then
        echo "󱊣" "green"
    elif [[ $battery_level -ge 50 ]]; then
        echo "󱊢" "yellow"
    elif [[ $battery_level -ge 20 ]]; then
        echo "󱊡" "peach"
    else
        echo "󰂎" "red"
    fi
}

format_battery_output() {
    local icon=$1
    local color=$2
    local status_text=$3

    printf " #[fg=\$%s]#[bg=\$%s,fg=\$crust]%s #[bg=\$surface1,fg=\$%s,bold] %s#[fg=\$surface1]" "$color" "$color" "$icon" "$color" "$status_text"
}

if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    battery_info=$(pmset -g batt)
    battery_percent=$(echo "$battery_info" | grep -o '[0-9]*%' | head -1 | sed 's/%//')
    charging_status=$(echo "$battery_info" | grep -o 'charging\|discharging\|charged' | head -1)

    read -r icon color <<< "$(get_battery_icon_and_color "$battery_percent")"

    if [[ "$charging_status" == "charging" ]]; then
        status_text="$battery_percent%+"
    elif [[ "$charging_status" == "charged" ]]; then
        status_text="$battery_percent%"
        color="green"
    else
        status_text="$battery_percent%"
    fi

    format_battery_output "$icon" "$color" "$status_text"
else
    # Linux (Ubuntu)
    battery_path=""
    for bat in /sys/class/power_supply/BAT*; do
        if [[ -d "$bat" ]]; then
            battery_path="$bat"
            break
        fi
    done

    if [[ -z "$battery_path" ]]; then
        printf ""
        exit 0
    fi

    capacity=$(cat "${battery_path}/capacity")
    status=$(cat "${battery_path}/status")

    read -r icon color <<< "$(get_battery_icon_and_color "$capacity")"

    case "$status" in
        "Charging")
            status_text="$capacity%+"
            ;;
        "Full")
            status_text="$capacity%"
            color="green"
            ;;
        *)
            status_text="$capacity%"
            ;;
    esac

    format_battery_output "$icon" "$color" "$status_text"
fi
