#!/bin/sh
set -eu
source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$source_dir"
exec go run . --source "$source_dir" "$@"
