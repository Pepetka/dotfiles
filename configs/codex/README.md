# Codex configuration

Reference copy of the portable part of `${CODEX_HOME:-~/.codex}/config.toml`.
Nothing installs or synchronizes it: compare it with the local file and copy
changes by hand when needed.

Trust entries, onboarding state, Pencil and other machine-only settings stay
local. Context7 reads `CONTEXT7_API_KEY` from the environment. Playwright
requires Node/npm.

Themes live in `theme/resources/codex`; see [CLI utilities](../../theme/README.md#cli-utilities).
