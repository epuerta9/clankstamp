# clankstamp

> Diffs show what changed. Chat says why. **clankstamp** lets you replay both.

`clankstamp` is a Go CLI, a JSONL artifact format, an embedded agent skill, and a Neovim plugin
for creating human-readable **replay tours** of AI coding work.

A **stamp** is the durable artifact an agent leaves behind after a meaningful change. Each stamp
captures *what* changed, *why* it changed, *where* it changed (file + hunk), the evidence
supporting the change (tests, migrations, git refs), and the review questions a human should
answer. The Neovim plugin walks you through the stamp step-by-step: it opens the right buffer,
highlights the hunk, and shows a side panel with the rationale and review checklist.

## Status

Early scaffolding. See [PRD.md](./PRD.md) for the full product spec.

## Components (planned)

- **`clankstamp` Go binary** — CLI for creating, validating, indexing, and viewing stamps.
- **Embedded Neovim plugin** — installed by `clankstamp nvim install`; replay UI inside nvim.
- **Embedded agent skill** — `SKILL.md` for Claude Code and the open `SKILL.md` standard,
  installed by `clankstamp skill install`.
- **`.clankstamp/` directory** — project-local JSONL artifact store.

## Quick goal

```bash
# After an agent finishes work:
clankstamp create --from-worktree --title "Add tenant API keys" --auto-tour

# In nvim:
<leader>rr   " open the stamp picker
<leader>rn   " step forward through the tour
```

## Install (planned)

```bash
brew install epuerta9/tap/clankstamp
clankstamp nvim install
clankstamp skill install --target claude-code --scope global
clankstamp doctor
```

## Repo layout

```
clankstamp/
  cmd/clankstamp/        # CLI entry point
  internal/              # CLI, replay engine, git, schema, tour, nvim, skill, doctor
  embedded/
    nvim/                # embedded Lua plugin (installed by `clankstamp nvim install`)
    skills/              # embedded SKILL.md (standard + claude-code variants)
    schemas/             # JSON schemas for manifest, events, tour steps
  PRD.md                 # full product spec
```

## License

MIT. See [LICENSE](./LICENSE).
