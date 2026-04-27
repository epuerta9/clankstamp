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

M.needs_review = safe_call(function()
  vim.notify("clankstamp: needs_review persistence not wired yet", vim.log.levels.INFO)
end)
M.accept = safe_call(function()
  vim.notify("clankstamp: accept persistence not wired yet", vim.log.levels.INFO)
end)

-- doctor prints binary version + a quick stamp count. Useful when wiring up
-- a new machine; the future :checkhealth provider can extend this.
M.doctor = safe_call(function()
  local client = require("clankstamp.client")
  local ok, version = client.check_binary()
  if not ok then
    vim.notify("clankstamp binary check failed: " .. tostring(version), vim.log.levels.ERROR)
    return
  end
  local entries = client.list()
  vim.notify(string.format("clankstamp ok — %s, %d stamp(s) in this repo", version, #entries), vim.log.levels.INFO)
end)

function M.setup(opts) M._opts = opts or {} end

return M
