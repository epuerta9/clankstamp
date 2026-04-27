-- clankstamp.highlight — extmark-based hunk highlighting.
--
-- The current tour step's file ranges are painted in a side-color so the user
-- can see what changed without leaving the buffer. Old extmarks are cleared
-- on every step transition.

local M = {}

local ns = vim.api.nvim_create_namespace("clankstamp")

-- ClankstampHunk is the highlight group applied to each line of the active
-- step's file ranges. Linked to DiffAdd by default; users can override:
--   vim.api.nvim_set_hl(0, "ClankstampHunk", { bg = "#1e2f1e" })
vim.api.nvim_set_hl(0, "ClankstampHunk", { link = "DiffAdd", default = true })

-- highlighted_bufs tracks which buffers we've drawn into so clear_all can
-- target only those (rather than walking every loaded buffer).
local highlighted_bufs = {}

function M.highlight_range(buf, line_start, line_end)
  if not buf or not vim.api.nvim_buf_is_valid(buf) then return end
  -- JSON line numbers are 1-based; extmarks are 0-based.
  local s = math.max(0, (line_start or 1) - 1)
  local e = math.max(s, (line_end or line_start or 1) - 1)
  local last = vim.api.nvim_buf_line_count(buf) - 1
  if s > last then return end
  if e > last then e = last end
  for ln = s, e do
    pcall(vim.api.nvim_buf_set_extmark, buf, ns, ln, 0, {
      end_row = ln + 1,
      hl_group = "ClankstampHunk",
      hl_eol = true,
      priority = 200,
    })
  end
  highlighted_bufs[buf] = true
end

function M.clear_buf(buf)
  pcall(vim.api.nvim_buf_clear_namespace, buf, ns, 0, -1)
end

function M.clear_all()
  for buf in pairs(highlighted_bufs) do
    if vim.api.nvim_buf_is_valid(buf) then
      M.clear_buf(buf)
    end
  end
  highlighted_bufs = {}
end

return M
