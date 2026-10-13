# Codex configuration

Portable preferences and three profiles for Codex 0.134+:

- `codex --profile work`: medium reasoning, xhigh planning.
- `codex --profile deep`: xhigh reasoning and planning.
- `codex --profile review`: high reasoning, read-only command sandbox, no escalation approvals.

Profiles inherit the base model. Project layers and CLI flags may override them.
The review profile does not revoke permissions separately granted to MCP tools.

## Install and update

```sh
sh configs/codex/install.sh --dry-run
sh configs/codex/install.sh
```

Requires Go 1.27+, or Go 1.21+ which downloads the declared toolchain, on macOS or Linux. The installer runs the standalone Go module
in this directory; it does not start or change the theme engine. The TOML parser
is pinned in `go.mod`/`go.sum`. The first run downloads the Go dependency; cached
dependencies allow subsequent runs offline. Python, pip and virtual environments
are not required.

For installation without the Go toolchain on the target machine, build the
binary once and pass the source directory explicitly:

```sh
(cd configs/codex && go build -o /tmp/codex-sync .)
/tmp/codex-sync --source "$PWD/configs/codex" --dry-run
```

Targets `${CODEX_HOME:-~/.codex}`. Configuration files remain ordinary local files,
not symlinks. On first installation, existing values win and missing shared keys
are added. Subsequent installs compare the previous shared source, new source
and local values. Shared-only changes are applied, local-only changes preserved,
and conflicting edits abort the entire update. Resolve a conflict by setting the
local key to either the previous or new shared value, then rerun. Arrays replace
as a unit. Removed shared keys are removed only if locally unchanged.

The selected `tui.theme` is always preserved; Moon is only a default for a missing
selection. Both custom themes are installed from `theme/resources/codex` and remain
manual choices through `/theme`; the theme engine does not update the selection.
Existing theme files are updated only when identical to the last installed source;
unrecognized local theme files are never overwritten.

Trust entries, onboarding state, Pencil and other machine-only settings are not
shared. No authentication or session state is copied into the repository. Context7
reads `CONTEXT7_API_KEY` from the environment. Playwright requires Node/npm.
Search and notification settings are not overridden.

The synchronizer edits source ranges using the pinned TOML parser, preserving
unrelated bytes and comments instead of reserializing the whole document.
An edited inline table is reserialized as a unit; its trailing comment is retained.
It validates results before writing,
backs up modified files under `.dotfiles-sync/backups`, and atomically replaces
individual files. A failed write rolls back completed writes. Snapshot state lives
in `.dotfiles-sync/state.json`; keep it to retain three-way update history.
Concurrent synchronizers are locked; detected external edits abort the update.
Run installation between Codex setting changes: the client does not share this
lock, so a simultaneous write after the final check can still race.
Dry-run does not write target files or sync state; Go may populate its build and
module caches. Existing version-1 snapshots remain compatible after migration.

## Verification

```sh
sh configs/codex/install.sh --dry-run
(cd configs/codex && go test -race ./... && go vet ./...)
```
