-- clankstamp Lua module — public API surface.
--
-- The Lua plugin is intentionally thin: UI, keymaps, buffers, highlights, and
-- the picker live here. Stamp parsing, git integration, and tour generation
-- happen in the Go binary, called over stdio via `jobstart()`.
--
-- Each public function below is a stub for v0; see PRD §14 for the full UX.

local M = {}

local function not_implemented(name)
  vim.notify("clankstamp." .. name .. ": not implemented yet — see PRD.md", vim.log.levels.INFO)
end

function M.list() not_implemented("list") end
function M.tour() not_implemented("tour") end
function M.next() not_implemented("next") end
function M.prev() not_implemented("prev") end
function M.diff() not_implemented("diff") end
function M.open_file() not_implemented("open_file") end
function M.mark_understood() not_implemented("mark_understood") end
function M.needs_review() not_implemented("needs_review") end
function M.accept() not_implemented("accept") end

-- setup() is optional in v0; included so users can pre-wire config without
-- breaking when the real implementation lands.
function M.setup(opts)
  M._opts = opts or {}
end

return M
