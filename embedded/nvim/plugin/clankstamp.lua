-- clankstamp.nvim — entry point
--
-- Loaded automatically when Neovim scans the `pack/clankstamp/start/` directory.
-- This file only registers user commands and default keymaps. The real logic
-- lives under `lua/clankstamp/`.

if vim.g.loaded_clankstamp == 1 then
  return
end
vim.g.loaded_clankstamp = 1

local ok, clankstamp = pcall(require, "clankstamp")
if not ok then
  vim.notify("clankstamp: failed to load lua module: " .. tostring(clankstamp), vim.log.levels.ERROR)
  return
end

local function cmd(name, fn)
  vim.api.nvim_create_user_command(name, function() fn() end, {})
end

-- Canonical commands.
cmd("Clankstamp", clankstamp.list)
cmd("ClankstampList", clankstamp.list)
cmd("ClankstampTour", clankstamp.tour)
cmd("ClankstampNext", clankstamp.next)
cmd("ClankstampPrev", clankstamp.prev)
cmd("ClankstampNextStamp", clankstamp.next_stamp)
cmd("ClankstampPrevStamp", clankstamp.prev_stamp)
cmd("ClankstampDiff", clankstamp.diff)
cmd("ClankstampOpenFile", clankstamp.open_file)
cmd("ClankstampMarkUnderstood", clankstamp.mark_understood)
cmd("ClankstampNeedsReview", clankstamp.needs_review)
cmd("ClankstampAccept", clankstamp.accept)
cmd("ClankstampDoctor", clankstamp.doctor)
cmd("ClankstampClose", clankstamp.close)
cmd("ClankstampQuit", clankstamp.close) -- ergonomic alias
cmd("ClankstampToggleOverlay", clankstamp.toggle_overlay)

-- Aliases — typo tolerance and a short form. "Clank stamp" reads as two words
-- so :ClankStamp (camelCase) is a near-universal first guess; cover it.
cmd("ClankStamp", clankstamp.list)
cmd("CS", clankstamp.list)

-- <Plug> mappings — the proper way for users to wire their own keys without
-- relying on default leader bindings (which can collide with other plugins'
-- `<leader>r…` prefixes). Example user config:
--   vim.keymap.set("n", "<leader>rr", "<Plug>(ClankstampList)")
local function plug(name, fn)
  vim.keymap.set("n", "<Plug>(" .. name .. ")", fn, { silent = true, desc = "clankstamp: " .. name })
end
plug("ClankstampList", clankstamp.list)
plug("ClankstampNext", clankstamp.next)
plug("ClankstampPrev", clankstamp.prev)
plug("ClankstampNextStamp", clankstamp.next_stamp)
plug("ClankstampPrevStamp", clankstamp.prev_stamp)
plug("ClankstampDiff", clankstamp.diff)
plug("ClankstampOpen", clankstamp.open_file)
plug("ClankstampUnderstood", clankstamp.mark_understood)
plug("ClankstampClose", clankstamp.close)
plug("ClankstampToggleOverlay", clankstamp.toggle_overlay)

-- Default keymaps. Deferred to VimEnter so that `vim.g.mapleader` reflects
-- the user's final value: lazy.nvim and similar managers can run our plugin
-- before the user's leader assignment lands. We also skip any binding that's
-- already taken so we don't shadow the user's own `<leader>r…` prefix.
--
-- Hot-reload note: a `VimEnter` autocmd registered AFTER startup never fires —
-- the event is one-shot at boot. So when this file is re-sourced via
-- `:Lazy reload` or a hot-reload keymap, we must bind the maps directly
-- instead of waiting on an autocmd that won't run again. `vim.v.vim_did_enter`
-- is set to 1 once VimEnter has fired, which lets us pick the right path.
if vim.g.clankstamp_no_default_maps ~= 1 then
  local function bind_default_maps()
    local maps = {
      { "<leader>rr", clankstamp.list,            "clankstamp: list stamps" },
      { "<leader>rn", clankstamp.next,            "clankstamp: next tour step (auto-advances across stamps)" },
      { "<leader>rp", clankstamp.prev,            "clankstamp: previous tour step (auto-advances across stamps)" },
      { "<leader>rN", clankstamp.next_stamp,      "clankstamp: next stamp (created_at order)" },
      { "<leader>rP", clankstamp.prev_stamp,      "clankstamp: previous stamp (created_at order)" },
      { "<leader>rd", clankstamp.diff,            "clankstamp: show step diff" },
      { "<leader>ro", clankstamp.open_file,       "clankstamp: open step file" },
      { "<leader>ru", clankstamp.mark_understood, "clankstamp: mark step understood" },
      { "<leader>rq", clankstamp.close,           "clankstamp: close tour" },
      { "<leader>rt", clankstamp.toggle_overlay,  "clankstamp: toggle in-buffer overlay" },
    }
    for _, m in ipairs(maps) do
      local lhs, rhs, desc = m[1], m[2], m[3]
      if vim.fn.maparg(lhs, "n") == "" then
        vim.keymap.set("n", lhs, rhs, { silent = true, desc = desc })
      end
    end
  end

  if vim.v.vim_did_enter == 1 then
    -- Hot-reload path: VimEnter already fired this session, bind now.
    bind_default_maps()
  else
    -- Normal startup path: wait for VimEnter so the user's mapleader is set.
    vim.api.nvim_create_autocmd("VimEnter", {
      once = true,
      callback = bind_default_maps,
    })
  end
end
