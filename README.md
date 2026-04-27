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

Early but working: `init`, `create --from-worktree`, `list`, `show`, `validate`,
`skill install`, `nvim install` all work end-to-end. The Neovim plugin renders
the picker, opens files at the step's line range, highlights hunks, and walks
through the tour with `n`/`p`. See [PRD.md](./PRD.md) for the full product spec.

## Try it without writing any code

```bash
git clone https://github.com/epuerta9/clankstamp
cd clankstamp/examples/tenant-api-keys
nvim
```

Then in Neovim run `:Clankstamp` (or press `<leader>rr`) to walk through a
curated 4-step tour of an auth refactor. See
[examples/tenant-api-keys/README.md](./examples/tenant-api-keys/README.md)
for the walkthrough.

### Using lazy.nvim?

`clankstamp nvim install` drops the plugin under
`~/.local/share/nvim/site/pack/clankstamp/start/clankstamp.nvim/`, which Neovim
auto-loads. **But lazy.nvim defaults to `performance.rtp.reset = true`, which
strips that path off `runtimepath`.** If your `:Clankstamp` command doesn't
exist, that's why. Add this to your lazy plugin list:

```lua
{
  "epuerta9/clankstamp",
  -- Use the local install written by `clankstamp nvim install`. Drop this
  -- `dir` line once we cut a tagged release.
  dir = vim.fn.expand("~/.local/share/nvim/site/pack/clankstamp/start/clankstamp.nvim"),
  cmd = { "Clankstamp", "ClankstampList", "ClankstampNext", "ClankstampDoctor", "CS" },
  keys = {
    { "<leader>rr", "<Plug>(ClankstampList)", desc = "clankstamp: list stamps" },
    { "<leader>rn", "<Plug>(ClankstampNext)", desc = "clankstamp: next step" },
    { "<leader>rp", "<Plug>(ClankstampPrev)", desc = "clankstamp: prev step" },
    { "<leader>rd", "<Plug>(ClankstampDiff)", desc = "clankstamp: diff" },
    { "<leader>ro", "<Plug>(ClankstampOpen)", desc = "clankstamp: open file" },
  },
}
```

If your `:Clankstamp` command is missing and you want to know which side of
the loader it's on, run:

```vim
:set rtp?     " is clankstamp.nvim in there? if not, lazy.nvim reset rtp
:source ~/.local/share/nvim/site/pack/clankstamp/start/clankstamp.nvim/plugin/clankstamp.lua
:Clankstamp   " if this works after sourcing, it's a load-path issue, not a code issue
:ClankstampDoctor   " prints binary version, stamp count, leader, and keymap state
```

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
