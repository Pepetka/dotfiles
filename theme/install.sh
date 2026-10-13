#!/usr/bin/env bash
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
command -v go >/dev/null || {
  echo 'Go is required to run the theme installer (1.27+, or 1.21+ to download it automatically).' >&2
  exit 1
}
working_directory="$PWD"
cd "$source_directory"
# Older Go (e.g. Ubuntu packages) downloads the toolchain required by go.mod.
exec go run ./cmd/theme-install --source "$source_directory" --working-dir "$working_directory" "$@"
