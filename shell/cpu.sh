#!/bin/bash

if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    cpu_usage=$(top -l 1 | grep "CPU usage" | awk '{print $3}' | sed 's/%//')
    printf "%2.0f%%\n" "${cpu_usage%.*}"
else
    # Linux (Ubuntu)
    cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')
    printf "%2.0f%%\n" "${cpu_usage%.*}"
fi
