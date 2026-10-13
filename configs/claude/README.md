# Claude Code configuration

- `settings.json` — reference copy of `~/.claude/settings.json`. Claude Code rewrites
  that file itself, so it is not linked: compare and copy changes by hand.
- `statusline-command.sh` — linked to `~/.claude/statusline-command.sh`; requires `jq`.
- `mcp.json` — portable user MCP servers. Context7 reads `CONTEXT7_API_KEY` from the
  environment; Playwright requires Node/npm. Pencil and other machine-only servers stay local.

User MCP servers live in `~/.claude.json` together with session state, so add them
through the CLI instead of copying the file:

```sh
for name in $(jq -r '.mcpServers | keys[]' configs/claude/mcp.json); do
  claude mcp add-json -s user "$name" "$(jq -c ".mcpServers[\"$name\"]" configs/claude/mcp.json)"
done
```

Global instructions and skills are shared with Codex; see [configs/agents](../agents/).
