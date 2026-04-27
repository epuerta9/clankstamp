-- clankstamp.highlight — extmark-based hunk highlighting + intent overlay.
--
-- Two visual layers, each in its own namespace so we can clear them
-- independently:
--
--   ns_lines    — line-background highlights on the hunk range. The
--                 "what changed" cue. Always on.
--   ns_overlay  — virt_lines above the hunk showing title / why / risk.
--                 The "why this changed" cue. Toggleable so power-users
--                 can flip back to clean code.

local M = {}

local ns_lines = vim.api.nvim_create_namespace("clankstamp.lines")
local ns_overlay = vim.api.nvim_create_namespace("clankstamp.overlay")

-- Public so other modules can compose in this namespace if they want
-- (e.g. tests asserting on extmark count).
M.ns_lines = ns_lines
M.ns_overlay = ns_overlay

-- ClankstampHunk is applied to each line of the active step's file ranges.
-- Linked to DiffAdd by default; users can override:
--   vim.api.nvim_set_hl(0, "ClankstampHunk", { bg = "#1e2f1e" })
vim.api.nvim_set_hl(0, "ClankstampHunk", { link = "DiffAdd", default = true })

-- Overlay highlight groups — linked to standard diagnostic colors so they
-- inherit any colorscheme automatically.
vim.api.nvim_set_hl(0, "ClankstampOverlayTitle", { link = "Title", default = true })
vim.api.nvim_set_hl(0, "ClankstampOverlayWhy",   { link = "Comment", default = true })
vim.api.nvim_set_hl(0, "ClankstampOverlayRiskHigh",   { link = "DiagnosticError", default = true })
vim.api.nvim_set_hl(0, "ClankstampOverlayRiskMedium", { link = "DiagnosticWarn",  default = true })
vim.api.nvim_set_hl(0, "ClankstampOverlayRiskLow",    { link = "DiagnosticHint",  default = true })
vim.api.nvim_set_hl(0, "ClankstampOverlayConn",  { link = "Special", default = true })

local highlighted_bufs = {}
local overlay_bufs = {}

-- wrap is a tiny word-wrapper: splits `s` at whitespace into segments of at
-- most `width` chars each. Used to keep the inline `why` line readable.
local function wrap(s, width)
  if #s <= width then return { s } end
  local out, cur = {}, ""
  for word in s:gmatch("%S+") do
    if #cur == 0 then
      cur = word
    elseif #cur + 1 + #word <= width then
      cur = cur .. " " .. word
    else
      table.insert(out, cur)
      cur = word
    end
  end
  if cur ~= "" then table.insert(out, cur) end
  return out
end

function M.highlight_range(buf, line_start, line_end)
  if not buf or not vim.api.nvim_buf_is_valid(buf) then return end
  -- JSON line numbers are 1-based; extmarks are 0-based.
  local s = math.max(0, (line_start or 1) - 1)
  local e = math.max(s, (line_end or line_start or 1) - 1)
  local last = vim.api.nvim_buf_line_count(buf) - 1
  if s > last then return end
  if e > last then e = last end
  for ln = s, e do
    pcall(vim.api.nvim_buf_set_extmark, buf, ns_lines, ln, 0, {
      end_row = ln + 1,
      hl_group = "ClankstampHunk",
      hl_eol = true,
      priority = 200,
    })
  end
  highlighted_bufs[buf] = true
end

-- overlay_step places virt_lines above `line_start` showing the step's
-- title, why, risk, and any cross-step connections — so the intent reads
-- as inline commentary above the hunk, not in a separate panel.
--
-- Each entry in `connection_summaries` is a plain string like
-- "tests step 3 (Auth middleware now checks API keys)".
function M.overlay_step(buf, line_start, step, connection_summaries)
  if not buf or not vim.api.nvim_buf_is_valid(buf) or not line_start then return end
  if line_start < 1 then line_start = 1 end
  local last = vim.api.nvim_buf_line_count(buf) - 1
  local row = math.min(line_start - 1, last)

  local risk_hl = "ClankstampOverlayWhy"
  if step.risk == "high" then
    risk_hl = "ClankstampOverlayRiskHigh"
  elseif step.risk == "medium" then
    risk_hl = "ClankstampOverlayRiskMedium"
  elseif step.risk == "low" then
    risk_hl = "ClankstampOverlayRiskLow"
  end

  local virt = {}
  table.insert(virt, { { string.format("▸ step %s · %s", step.order or "?", step.title or ""), "ClankstampOverlayTitle" } })
  if step.why and step.why ~= "" then
    -- Wrap the why at ~80 chars so a long sentence doesn't run off-screen.
    for _, line in ipairs(wrap(step.why, 80)) do
      table.insert(virt, { { "  why: " .. line, "ClankstampOverlayWhy" } })
    end
  end
  if step.risk and step.risk ~= "" then
    table.insert(virt, { { "  risk: " .. step.risk, risk_hl } })
  end
  if connection_summaries and #connection_summaries > 0 then
    for _, c in ipairs(connection_summaries) do
      table.insert(virt, { { "  → " .. c, "ClankstampOverlayConn" } })
    end
  end
  table.insert(virt, { { "  ─── code ───", "ClankstampOverlayWhy" } })

  pcall(vim.api.nvim_buf_set_extmark, buf, ns_overlay, row, 0, {
    virt_lines = virt,
    virt_lines_above = true,
    priority = 100,
  })
  overlay_bufs[buf] = true
end

function M.clear_buf(buf)
  pcall(vim.api.nvim_buf_clear_namespace, buf, ns_lines, 0, -1)
  pcall(vim.api.nvim_buf_clear_namespace, buf, ns_overlay, 0, -1)
end

function M.clear_overlays()
  for buf in pairs(overlay_bufs) do
    if vim.api.nvim_buf_is_valid(buf) then
      pcall(vim.api.nvim_buf_clear_namespace, buf, ns_overlay, 0, -1)
    end
  end
  overlay_bufs = {}
end

function M.clear_all()
  for buf in pairs(highlighted_bufs) do
    if vim.api.nvim_buf_is_valid(buf) then
      M.clear_buf(buf)
    end
  end
  for buf in pairs(overlay_bufs) do
    if vim.api.nvim_buf_is_valid(buf) then
      pcall(vim.api.nvim_buf_clear_namespace, buf, ns_overlay, 0, -1)
    end
  end
  highlighted_bufs = {}
  overlay_bufs = {}
end

return M
