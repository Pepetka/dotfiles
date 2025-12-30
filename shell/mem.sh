#!/bin/bash

if [[ "$OSTYPE" == "darwin"* ]]; then
    vm_stat_output=$(vm_stat)
    
    pages_wired=$(echo "$vm_stat_output" | awk '/Pages wired down/ {print $4}' | sed 's/\.//')
    pages_active=$(echo "$vm_stat_output" | awk '/Pages active/ {print $3}' | sed 's/\.//')
    pages_inactive=$(echo "$vm_stat_output" | awk '/Pages inactive/ {print $3}' | sed 's/\.//')
    pages_free=$(echo "$vm_stat_output" | awk '/Pages free/ {print $3}' | sed 's/\.//')
    pages_speculative=$(echo "$vm_stat_output" | awk '/Pages speculative/ {print $3}' | sed 's/\.//')
    
    total_pages=$((pages_wired + pages_active + pages_inactive + pages_free + pages_speculative))
    used_pages=$((pages_wired + pages_active))
    
    ram_usage=$((used_pages * 100 / total_pages))
    printf "%2d%%\n" "$ram_usage"
else
    while IFS=: read -r key value; do
        case $key in
            MemTotal) mem_total=${value//[^0-9]/} ;;
            MemAvailable) mem_available=${value//[^0-9]/} ;;
        esac
    done < /proc/meminfo
    
    mem_used=$((mem_total - mem_available))
    ram_usage=$((mem_used * 100 / mem_total))
    printf "%2d%%\n" "$ram_usage"
fi
