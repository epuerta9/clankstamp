# Example: Add tenant-scoped API keys

A self-contained, walkable clankstamp tour. Cloning this repo and opening
Neovim in this directory will give you the full UX without writing a single
line of code yourself.

## What's the change?

A small Go service originally only authenticated via user JWTs. The "agent"
extended it to support tenant-scoped API keys (`X-API-Key`), with a new
`APIKey` model, a `TenantID` field on `User`, the middleware updated to check
keys before falling back to JWT, and tests covering both paths.

## Walk through it

```bash
cd examples/tenant-api-keys
nvim
```

Then in Neovim:

```vim
:Clankstamp        " or <leader>rr — opens the picker
```

Pick the one stamp (`Add tenant-scoped API keys`) and Neovim opens the tour
panel on the right and the relevant file in the main window with the active
hunk highlighted.

| Key | Action |
|-----|--------|
| `n` / `<leader>rn` | Next step |
| `p` / `<leader>rp` | Previous step |
| `o` / `<CR>` | Re-focus the step's file |
| `d` / `<leader>rd` | Open the full unified patch in a new tab |
| `q` | Close the tour |

The tour has 4 comprehension-ordered steps:

1. **Tenant scoping on User** — `models/user.go:3-7` — the foundation
2. **New tenant-scoped APIKey model** — `models/user.go:9-17` — credentials
3. **Auth middleware now checks API keys** — `auth/middleware.go:10-23` — the behavior change (high risk)
4. **Tests cover both auth paths** — `tests/auth_test.go:8-23` — evidence

Notice the order: data model first, then the behavior that uses it, then the
tests that prove it. That's what an agent skill should produce — *not* the
file-alphabetical order a raw `git diff` gives you.

## What's actually in this directory

```
examples/tenant-api-keys/
├── auth/middleware.go             ← post-edit source (matches the hunks)
├── models/user.go                 ← post-edit source
├── tests/auth_test.go             ← post-edit source
└── .clankstamp/
    ├── index.jsonl                ← one-line summary per stamp
    └── runs/
        └── run_20260427_184208_DEMO-42/
            ├── manifest.json      ← stamp metadata (title, status, repo info)
            ├── events.jsonl       ← run.started / run.finished bookends
            ├── tour.jsonl         ← the curated 4-step tour
            ├── hunks.jsonl        ← machine-extracted hunk records
            └── patches/full.patch ← the unified diff against the baseline
```

You can also poke at the stamp from the CLI without opening Neovim:

```bash
cd examples/tenant-api-keys

clankstamp list                                      # one-line table
clankstamp show run_20260427_184208_DEMO-42          # human-readable
clankstamp show run_20260427_184208_DEMO-42 --json | jq .tour
clankstamp validate                                  # should report 0 errors
```

## Compared to a raw `git diff`

Open `.clankstamp/runs/run_20260427_184208_DEMO-42/patches/full.patch` in any
viewer. That's what your reviewer normally sees: three file diffs in
alphabetical order, no narrative, no risk callouts, no review questions. The
tour in `tour.jsonl` is the same change with comprehension order, rationale,
and review prompts attached.
