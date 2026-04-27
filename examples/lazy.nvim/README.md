# lazy.nvim / LazyVim integration

LazyVim (and any nvim config built on lazy.nvim with the default
`performance.rtp.reset = true`) silently drops pack/start plugins from
`runtimepath`. That means `clankstamp nvim install` writes the plugin to
disk but Neovim never sees it — `:Clankstamp` doesn't exist, the leader
keymaps don't fire.

The fix is to add clankstamp as a lazy.nvim plugin spec.

## Install

```bash
cp clankstamp.lua ~/.config/nvim/lua/plugins/
```

Then in Neovim:

```vim
:Lazy reload
:Clankstamp
```

That's it. The first `:Clankstamp` (or `<leader>rr`) will trigger
lazy.nvim to load the plugin, register the user commands and `<Plug>`
mappings, and open the picker.

## What's in [clankstamp.lua](./clankstamp.lua)

Three things matter:

1. **`dir = ...`** points lazy at the locally-installed plugin tree at
   `~/.local/share/nvim/site/pack/clankstamp/start/clankstamp.nvim/`,
   which is exactly where `clankstamp nvim install` wrote it. No
   network needed; the plugin you run is the one your binary embeds.

2. **`cmd = { ... }`** lazy-loads on first `:Clankstamp` (and the typo
   aliases `:ClankStamp` / `:CS`).

3. **`keys = { ... }`** lazy-loads on first `<leader>rr` etc., wiring
   each key to a `<Plug>` mapping. Using `<Plug>` instead of a Lua
   function reference is what lets lazy.nvim's keys-based loading
   replay the keystroke after the plugin loads.

The `init = function() vim.g.clankstamp_no_default_maps = 1 end` line
prevents our plugin's own VimEnter keymap setup from running, since
lazy.nvim is already binding the same keys via `keys = {}`.

## Verify

```vim
:ClankstampDoctor
```

Should print binary version, stamp count, leader value, and the state
of each `<leader>r…` keymap (whether it's owned by us, shadowed, or
unbound).
