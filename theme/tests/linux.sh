#!/usr/bin/env bash
# Run Go tests and the Linux installer acceptance inside an Ubuntu container.
# Override the image base with THEME_LINUX_BASE; skipped when Docker is unavailable.
set -euo pipefail
source_directory="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
root_directory="$(dirname "$source_directory")"
if ! command -v docker >/dev/null || ! docker info >/dev/null 2>&1; then
  echo 'Docker is unavailable; skipping Linux checks'
  exit 0
fi
case "$(docker info --format '{{.Architecture}}')" in
aarch64 | arm64) architecture=arm64 ;;
x86_64 | amd64) architecture=amd64 ;;
*) echo 'Unsupported Docker architecture' >&2; exit 1 ;;
esac
binaries="$(mktemp -d "${TMPDIR:-/tmp}/theme-linux.XXXXXX")"
trap 'rm -rf "$binaries"' EXIT
packages=(./backends/linux ./cmd/theme ./internal/adapters ./internal/engine ./internal/install ./internal/service)
(
  cd "$source_directory"
  export GOOS=linux GOARCH="$architecture" CGO_ENABLED=0
  go build -o "$binaries/engine" ./cmd/theme
  go build -o "$binaries/installer" ./cmd/theme-install
  go build -o "$binaries/linux-backend" ./backends/linux
  for package in "${packages[@]}"; do
    go test -c -o "$binaries/$(basename "$package").test" "$package"
  done
)
image=dotfiles-theme-linux
docker build -q --build-arg "BASE=${THEME_LINUX_BASE:-ubuntu:24.04}" -t "$image" "$source_directory/tests/linux" >/dev/null
# Each test binary runs from its package directory, as go test does.
docker run --rm -v "$root_directory:/dotfiles:ro" -v "$binaries:/binaries:ro" \
  -e THEME_TEST_BINARIES=/binaries -e PACKAGES="${packages[*]}" "$image" bash -c '
set -euo pipefail
for package in $PACKAGES; do
  (cd "/dotfiles/theme/$package" && "/binaries/$(basename "$package").test" -test.count=1 >/tmp/test.log 2>&1) || { cat /tmp/test.log; exit 1; }
  echo "ok  $package"
done
bash /dotfiles/theme/tests/install-linux.sh
'
echo 'Linux container checks passed'
