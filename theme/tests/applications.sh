#!/usr/bin/env bash
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$source_directory"
THEME_APPLICATION_TESTS=1 go test -count=1 -v ./internal/integration
