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
cmd("ClankstampDiff", clankstamp.diff)
cmd("ClankstampOpenFile", clankstamp.open_file)
cmd("ClankstampMarkUnderstood", clankstamp.mark_understood)
cmd("ClankstampNeedsReview", clankstamp.needs_review)
cmd("ClankstampAccept", clankstamp.accept)
cmd("ClankstampDoctor", clankstamp.doctor)

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
plug("ClankstampDiff", clankstamp.diff)
plug("ClankstampOpen", clankstamp.open_file)
plug("ClankstampUnderstood", clankstamp.mark_understood)

-- Default keymaps. Deferred to VimEnter so that `vim.g.mapleader` reflects
-- the user's final value: lazy.nvim and similar managers can run our plugin
-- before the user's leader assignment lands. We also skip any binding that's
-- already taken so we don't shadow the user's own `<leader>r…` prefix.
if vim.g.clankstamp_no_default_maps ~= 1 then
  vim.api.nvim_create_autocmd("VimEnter", {
    once = true,
    callback = function()
      local maps = {
        { "<leader>rr", clankstamp.list,            "clankstamp: list stamps" },
        { "<leader>rn", clankstamp.next,            "clankstamp: next tour step" },
        { "<leader>rp", clankstamp.prev,            "clankstamp: previous tour step" },
        { "<leader>rd", clankstamp.diff,            "clankstamp: show step diff" },
        { "<leader>ro", clankstamp.open_file,       "clankstamp: open step file" },
        { "<leader>ru", clankstamp.mark_understood, "clankstamp: mark step understood" },
      }
      for _, m in ipairs(maps) do
        local lhs, rhs, desc = m[1], m[2], m[3]
        if vim.fn.maparg(lhs, "n") == "" then
          vim.keymap.set("n", lhs, rhs, { silent = true, desc = desc })
        end
      end
    end,
  })
end
