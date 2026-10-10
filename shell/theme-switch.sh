#!/usr/bin/env bash
set -euo pipefail
script_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "${DOTFILES:-$(dirname "$script_directory")}/theme/theme" "$@"
