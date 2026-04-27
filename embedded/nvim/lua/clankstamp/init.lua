-- clankstamp.nvim — public API.
--
-- The plugin entry (plugin/clankstamp.lua) wires user commands and default
-- keymaps to the functions exposed here. The actual logic lives in
-- ./client.lua (binary shellouts), ./picker.lua (vim.ui.select wrapper),
-- ./tour.lua (panel + navigation), and ./highlight.lua (extmark management).

local M = {}

-- safe_call wraps an entry point so any errors surface as :notify messages
-- instead of raw stack traces. Saves the user from having to read Lua
-- tracebacks for "binary not on PATH" or "stamp not found" type failures.
local function safe_call(fn)
  return function(...)
    local ok, err = pcall(fn, ...)
    if not ok then
      vim.notify("clankstamp: " .. tostring(err), vim.log.levels.ERROR)
    end
  end
end

-- list opens the picker; on selection, opens the tour for that stamp.
M.list = safe_call(function()
  local picker = require("clankstamp.picker")
  local tour = require("clankstamp.tour")
  picker.pick(function(entry)
    if entry then tour.open(entry.run_id) end
  end)
end)

M.tour = safe_call(function() require("clankstamp.tour").open_panel() end)
M.next = safe_call(function() require("clankstamp.tour").advance(1) end)
M.prev = safe_call(function() require("clankstamp.tour").advance(-1) end)
M.diff = safe_call(function() require("clankstamp.tour").show_diff() end)
M.open_file = safe_call(function() require("clankstamp.tour").open_file_at_cursor() end)
M.mark_understood = safe_call(function() require("clankstamp.tour").mark_understood() end)
M.close = safe_call(function() require("clankstamp.tour").close() end)
M.toggle_overlay = safe_call(function() require("clankstamp.tour").toggle_overlay() end)

M.needs_review = safe_call(function()
  vim.notify("clankstamp: needs_review persistence not wired yet", vim.log.levels.INFO)
end)
M.accept = safe_call(function()
  vim.notify("clankstamp: accept persistence not wired yet", vim.log.levels.INFO)
end)

-- doctor prints a diagnostic report: binary version, stamp count, and the
-- state of the default keymaps. Useful when wiring up a new machine — the
-- top failure mode is "leader keymap doesn't fire because another plugin
-- shadows <leader>r…".
M.doctor = safe_call(function()
  local client = require("clankstamp.client")
  local lines = { "clankstamp doctor:" }

  local ok, version = client.check_binary()
  if ok then
    table.insert(lines, "  binary:    " .. tostring(version))
  else
    table.insert(lines, "  binary:    NOT FOUND — " .. tostring(version))
  end

  local list_ok, entries = pcall(client.list)
  if list_ok then
    table.insert(lines, "  stamps:    " .. tostring(#entries))
  else
    table.insert(lines, "  stamps:    error — " .. tostring(entries))
  end

  -- Leader + keymap state. maparg returns the raw RHS or "" if unbound.
  -- maparg with a 4th arg returns a table with a `desc` field for richer
  -- output, which lets us tell "ours" from "shadowed by another plugin".
  local leader = vim.g.mapleader
  table.insert(lines, "  leader:    " .. (leader == nil and "<NIL — defaults to \\>" or vim.inspect(leader)))
  table.insert(lines, "  keymaps:")
  local checks = {
    "<leader>rr", "<leader>rn", "<leader>rp",
    "<leader>rd", "<leader>ro", "<leader>ru",
  }
  for _, lhs in ipairs(checks) do
    local info = vim.fn.maparg(lhs, "n", false, true)
    if type(info) == "table" and info.lhs then
      local desc = info.desc or "(no desc)"
      local owner = desc:match("^clankstamp:") and "ours" or "SHADOWED"
      table.insert(lines, string.format("    %-14s %s — %s", lhs, owner, desc))
    else
      table.insert(lines, string.format("    %-14s NOT BOUND", lhs))
    end
  end

  -- Commands sanity check.
  table.insert(lines, "  commands:")
  for _, c in ipairs({ "Clankstamp", "ClankstampNext", "ClankstampDoctor", "CS" }) do
    local exists = vim.fn.exists(":" .. c) == 2
    table.insert(lines, string.format("    :%-22s %s", c, exists and "ok" or "MISSING"))
  end

  vim.notify(table.concat(lines, "\n"), vim.log.levels.INFO)
end)

function M.setup(opts) M._opts = opts or {} end

return M
