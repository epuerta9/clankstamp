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

vim.api.nvim_create_user_command("Clankstamp", function() clankstamp.list() end, {})
vim.api.nvim_create_user_command("ClankstampList", function() clankstamp.list() end, {})
vim.api.nvim_create_user_command("ClankstampTour", function() clankstamp.tour() end, {})
vim.api.nvim_create_user_command("ClankstampNext", function() clankstamp.next() end, {})
vim.api.nvim_create_user_command("ClankstampPrev", function() clankstamp.prev() end, {})
vim.api.nvim_create_user_command("ClankstampDiff", function() clankstamp.diff() end, {})
vim.api.nvim_create_user_command("ClankstampOpenFile", function() clankstamp.open_file() end, {})
vim.api.nvim_create_user_command("ClankstampMarkUnderstood", function() clankstamp.mark_understood() end, {})
vim.api.nvim_create_user_command("ClankstampNeedsReview", function() clankstamp.needs_review() end, {})
vim.api.nvim_create_user_command("ClankstampAccept", function() clankstamp.accept() end, {})

-- Default keymaps; users can override by setting `vim.g.clankstamp_no_default_maps = 1`.
if vim.g.clankstamp_no_default_maps ~= 1 then
  local map = function(lhs, rhs, desc)
    vim.keymap.set("n", lhs, rhs, { desc = desc, silent = true })
  end
  map("<leader>rr", function() clankstamp.list() end, "clankstamp: list stamps")
  map("<leader>rn", function() clankstamp.next() end, "clankstamp: next tour step")
  map("<leader>rp", function() clankstamp.prev() end, "clankstamp: previous tour step")
  map("<leader>rd", function() clankstamp.diff() end, "clankstamp: show step diff")
  map("<leader>ro", function() clankstamp.open_file() end, "clankstamp: open step file")
  map("<leader>ru", function() clankstamp.mark_understood() end, "clankstamp: mark step understood")
end
