# clankstamp examples

Self-contained example stamps you can walk through in Neovim without
generating anything yourself. Each example is a directory containing both
the post-edit source files and a checked-in `.clankstamp/runs/<id>/` so the
tour opens straight away.

| Example | What it shows |
|---------|---------------|
| [`tenant-api-keys/`](./tenant-api-keys/) | A 4-step comprehension-ordered tour of an auth refactor: new data model → behavior change → tests. Demonstrates a full agent-curated `tour.jsonl` (vs the synthesized fallback). |

## How to walk through any example

```bash
cd examples/<name>
nvim
```

Then `:Clankstamp` (or `<leader>rr`) opens the picker.

The CLI also works standalone — `clankstamp list`, `clankstamp show
<run_id>`, `clankstamp validate` all read from the example's
`.clankstamp/` directly. (You don't need a `.git` directory; clankstamp
finds the store by walking up from cwd.)

## Want to add an example?

Hand-curated stamps live under `examples/<name>/` and consist of:

- The post-edit source files at the paths the hunks reference
- `.clankstamp/runs/run_<id>/manifest.json` — title, status, schema version
- `.clankstamp/runs/run_<id>/tour.jsonl` — the comprehension-ordered tour (one JSON record per line)
- `.clankstamp/runs/run_<id>/hunks.jsonl` — machine-extracted hunk records
- `.clankstamp/runs/run_<id>/events.jsonl` — at minimum a `run.started` + `run.finished`
- `.clankstamp/runs/run_<id>/patches/full.patch` — the unified diff
- `.clankstamp/index.jsonl` — one-line summary per stamp

Validate with `clankstamp validate` — it should report 0 errors.
