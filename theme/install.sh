#!/usr/bin/env bash
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
command -v go >/dev/null || {
  echo 'Go >= 1.27 is required to run the theme installer.' >&2
  exit 1
}
working_directory="$PWD"
cd "$source_directory"
exec env GOTOOLCHAIN=local go run ./cmd/theme-install --source "$source_directory" --working-dir "$working_directory" "$@"
