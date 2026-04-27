-- clankstamp.picker — pick a stamp from .clankstamp/runs/.
--
-- v0 uses vim.ui.select (built-in, works with the user's preferred picker
-- via vim.ui.select-overriding plugins like dressing.nvim or telescope-ui-select).
-- Telescope-native integration is a v2 item per PRD §14.4.

local M = {}

function M.pick(callback)
  local client = require("clankstamp.client")
  local ok, entries = pcall(client.list)
  if not ok then
    vim.notify("clankstamp: " .. tostring(entries), vim.log.levels.ERROR)
    return
  end
  if not entries or #entries == 0 then
    vim.notify(
      "clankstamp: no stamps yet — run `clankstamp create --from-worktree --title \"...\"`",
      vim.log.levels.INFO
    )
    return
  end
  vim.ui.select(entries, {
    prompt = "clankstamp: pick a stamp",
    format_item = function(e)
      return string.format(
        "%s  %-8s  %dh/%ds  %s",
        e.run_id, e.status or "?", e.hunks or 0, e.tour_steps or 0, e.title or ""
      )
    end,
  }, function(choice)
    if choice and callback then callback(choice) end
  end)
end

return M
