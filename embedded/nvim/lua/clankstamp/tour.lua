-- clankstamp.tour — drives the side-panel + step-by-step navigation.
--
-- The Go binary returns `clankstamp show --json`:
--   { manifest, tour, hunks, events }
-- We open the side panel, render the active step, jump the main window to the
-- step's file, and highlight the relevant lines via the highlight namespace.
--
-- If the stamp's `tour` is empty, we synthesize one step per unique file from
-- the hunks. This means every stamp has a navigable tour even before the
-- agent skill has filled in tour.jsonl — the value of stepping through hunks
-- is real on its own, and richer agent-written tours layer on top later.

local M = {}

local hl = require("clankstamp.highlight")

-- Module-level state. The plugin is single-tour at a time in v0; opening a
-- new stamp replaces the previous one.
local state = {
  run_id = nil,
  payload = nil, -- result of `clankstamp show --json`
  steps = nil,   -- normalized step list (real tour or synthesized from hunks)
  step_idx = 1,
  panel_buf = nil,
  panel_win = nil,
}

-- synthesize_steps_from_hunks builds a fallback step list when tour.jsonl is
-- empty: one step per unique file, with line ranges spanning all hunks in
-- that file.
local function synthesize_steps_from_hunks(payload)
  local by_file = {}
  local order = {}
  local hunks = payload.hunks
  if type(hunks) ~= "table" then hunks = {} end
  for _, h in ipairs(hunks) do
    if not by_file[h.file] then
      by_file[h.file] = {
        min = h.new_start,
        max = h.new_start + math.max(h.new_lines or 1, 1) - 1,
      }
      table.insert(order, h.file)
    end
    local b = by_file[h.file]
    b.min = math.min(b.min, h.new_start)
    b.max = math.max(b.max, h.new_start + math.max(h.new_lines or 1, 1) - 1)
  end
  local steps = {}
  for i, file in ipairs(order) do
    table.insert(steps, {
      step_id = "auto_" .. i,
      order = i,
      title = "Changes in " .. file,
      summary = "(synthesized step — agent skill has not generated a tour yet)",
      kind = "auto",
      files = { {
        path = file,
        line_start = by_file[file].min,
        line_end = by_file[file].max,
      } },
    })
  end
  return steps
end

local function normalize_path(path)
  if path:match("^/") then return path end
  if not state.payload then return path end
  -- Prefer project_root (where .clankstamp/ currently lives) over manifest's
  -- recorded repo.root, which is just metadata from create time and may not
  -- match the current machine (e.g. a stamp checked into an examples/ dir).
  local root = state.payload.project_root
  if not root or root == "" then
    root = state.payload.manifest and state.payload.manifest.repo and state.payload.manifest.repo.root
  end
  if not root or root == "" then return path end
  return root .. "/" .. path
end

local function find_main_window()
  for _, w in ipairs(vim.api.nvim_list_wins()) do
    if w ~= state.panel_win and vim.api.nvim_win_is_valid(w) then
      return w
    end
  end
  return nil
end

-- open_panel creates the right-side scratch window that hosts the tour text.
-- It's idempotent: calling it when the panel is already up is a no-op.
function M.open_panel()
  if state.panel_win and vim.api.nvim_win_is_valid(state.panel_win) then
    return
  end
  local buf = vim.api.nvim_create_buf(false, true)
  vim.bo[buf].buftype = "nofile"
  vim.bo[buf].bufhidden = "wipe"
  vim.bo[buf].swapfile = false
  vim.bo[buf].filetype = "clankstamp_tour"
  vim.api.nvim_buf_set_name(buf, "clankstamp://tour")

  vim.cmd("botright vsplit")
  local win = vim.api.nvim_get_current_win()
  vim.api.nvim_win_set_buf(win, buf)
  vim.api.nvim_win_set_width(win, 52)
  vim.wo[win].number = false
  vim.wo[win].relativenumber = false
  vim.wo[win].signcolumn = "no"
  vim.wo[win].wrap = true
  vim.wo[win].linebreak = true
  vim.wo[win].cursorline = false
  vim.wo[win].winfixwidth = true

  state.panel_buf = buf
  state.panel_win = win

  -- Buffer-local maps for quick navigation without leaving the panel.
  local map = function(lhs, fn, desc)
    vim.keymap.set("n", lhs, fn, { buffer = buf, nowait = true, silent = true, desc = desc })
  end
  map("n", function() M.advance(1) end, "clankstamp: next step")
  map("p", function() M.advance(-1) end, "clankstamp: prev step")
  map("o", function() M.open_file_at_cursor() end, "clankstamp: open step file")
  map("d", function() M.show_diff() end, "clankstamp: show full diff")
  map("u", function() M.mark_understood() end, "clankstamp: mark step understood")
  map("q", function() M.close() end, "clankstamp: close tour")
  map("<CR>", function() M.open_file_at_cursor() end, "clankstamp: open step file")

  -- When the panel is closed by the user, drop our state so a fresh :Clankstamp
  -- works cleanly afterwards.
  vim.api.nvim_create_autocmd({ "BufWipeout", "WinClosed" }, {
    buffer = buf,
    once = true,
    callback = function()
      state.panel_buf = nil
      state.panel_win = nil
      hl.clear_all()
    end,
  })
end

local function set_panel_lines(lines)
  if not state.panel_buf or not vim.api.nvim_buf_is_valid(state.panel_buf) then return end
  vim.bo[state.panel_buf].modifiable = true
  vim.api.nvim_buf_set_lines(state.panel_buf, 0, -1, false, lines)
  vim.bo[state.panel_buf].modifiable = false
end

local function render_step(step)
  local m = state.payload and state.payload.manifest or {}
  local lines = {}
  table.insert(lines, string.format("clankstamp tour | step %d / %d", state.step_idx, #state.steps))
  table.insert(lines, string.rep("─", 50))
  table.insert(lines, "stamp: " .. (m.title or "?"))
  table.insert(lines, "id:    " .. (state.run_id or "?"))
  if m.task and m.task.id and m.task.id ~= "" then
    table.insert(lines, "task:  " .. m.task.id)
  end
  table.insert(lines, "")
  table.insert(lines, "▸ " .. (step.title or "(untitled step)"))
  table.insert(lines, string.rep("─", 50))
  table.insert(lines, "")
  if step.summary and step.summary ~= "" then
    for _, l in ipairs(vim.split(step.summary, "\n", { plain = true })) do
      table.insert(lines, l)
    end
    table.insert(lines, "")
  end
  if step.why and step.why ~= "" then
    table.insert(lines, "Why")
    for _, l in ipairs(vim.split(step.why, "\n", { plain = true })) do
      table.insert(lines, "  " .. l)
    end
    table.insert(lines, "")
  end
  if step.risk and step.risk ~= "" then
    table.insert(lines, "Risk: " .. step.risk)
    table.insert(lines, "")
  end
  if step.files and #step.files > 0 then
    table.insert(lines, "Files")
    for _, f in ipairs(step.files) do
      local s = "  " .. f.path
      if f.line_start and f.line_end then
        s = s .. ":" .. f.line_start .. "-" .. f.line_end
      end
      if f.symbol and f.symbol ~= "" then
        s = s .. "  (" .. f.symbol .. ")"
      end
      table.insert(lines, s)
    end
    table.insert(lines, "")
  end
  if step.review_questions and #step.review_questions > 0 then
    table.insert(lines, "Review")
    for _, q in ipairs(step.review_questions) do
      table.insert(lines, "  • " .. q)
    end
    table.insert(lines, "")
  end
  table.insert(lines, string.rep("─", 50))
  table.insert(lines, "[n] next  [p] prev  [o] open  [d] diff")
  table.insert(lines, "[u] understood  [q] quit")
  set_panel_lines(lines)
end

local function open_file_for_step(step)
  if not step.files or #step.files == 0 then return end
  local f = step.files[1]
  local path = normalize_path(f.path)
  local win = find_main_window()
  if not win then return end
  vim.api.nvim_set_current_win(win)
  if vim.fn.filereadable(path) == 1 then
    vim.cmd("edit " .. vim.fn.fnameescape(path))
    if f.line_start and f.line_start > 0 then
      local line_count = vim.api.nvim_buf_line_count(0)
      local target = math.min(f.line_start, line_count)
      vim.api.nvim_win_set_cursor(win, { target, 0 })
      vim.cmd("normal! zz")
    end
  else
    vim.notify("clankstamp: file not found: " .. path, vim.log.levels.WARN)
  end
end

local function highlight_step(step)
  hl.clear_all()
  if not step.files then return end
  for _, f in ipairs(step.files) do
    if f.line_start and f.line_end then
      local path = normalize_path(f.path)
      local buf = vim.fn.bufnr(path)
      if buf ~= -1 then
        hl.highlight_range(buf, f.line_start, f.line_end)
      end
    end
  end
end

function M.show_step()
  if not state.steps or #state.steps == 0 then return end
  local step = state.steps[state.step_idx]
  if not step then return end
  -- Open the file FIRST so the buffer exists, then highlight it.
  open_file_for_step(step)
  highlight_step(step)
  render_step(step)
end

-- open loads a stamp by run_id and starts the tour at step 1.
function M.open(run_id)
  local client = require("clankstamp.client")
  local payload = client.show(run_id)
  if not payload then
    vim.notify("clankstamp: empty payload for " .. run_id, vim.log.levels.ERROR)
    return
  end
  state.run_id = run_id
  state.payload = payload
  -- Defensive: vim.json.decode renders JSON null as vim.NIL (userdata), not
  -- a Lua nil. Treat anything that isn't a table-with-elements as "no tour".
  local tour = payload.tour
  if type(tour) == "table" and #tour > 0 then
    state.steps = tour
  else
    state.steps = synthesize_steps_from_hunks(payload)
  end
  state.step_idx = 1
  if #state.steps == 0 then
    vim.notify(
      "clankstamp: stamp " .. run_id .. " has no tour and no hunks (empty diff?)",
      vim.log.levels.WARN
    )
    return
  end
  M.open_panel()
  M.show_step()
end

function M.advance(delta)
  if not state.steps or #state.steps == 0 then return end
  local idx = state.step_idx + delta
  if idx < 1 then idx = 1 end
  if idx > #state.steps then idx = #state.steps end
  state.step_idx = idx
  M.show_step()
end

function M.show_diff()
  if not state.run_id or not state.payload then return end
  local root = state.payload.project_root
  if not root or root == "" then
    root = state.payload.manifest and state.payload.manifest.repo and state.payload.manifest.repo.root
  end
  local patch = root .. "/.clankstamp/runs/" .. state.run_id .. "/patches/full.patch"
  if vim.fn.filereadable(patch) ~= 1 then
    vim.notify("clankstamp: no patch file at " .. patch, vim.log.levels.WARN)
    return
  end
  vim.cmd("tabnew " .. vim.fn.fnameescape(patch))
  vim.bo.filetype = "diff"
end

function M.open_file_at_cursor()
  if state.steps and state.steps[state.step_idx] then
    open_file_for_step(state.steps[state.step_idx])
  end
end

function M.mark_understood()
  -- TODO(review.jsonl): persist via the Go binary once `clankstamp event` lands.
  vim.notify(
    string.format("clankstamp: marked step %d understood (persistence pending)", state.step_idx),
    vim.log.levels.INFO
  )
end

function M.close()
  if state.panel_win and vim.api.nvim_win_is_valid(state.panel_win) then
    vim.api.nvim_win_close(state.panel_win, true)
  end
  state.panel_buf = nil
  state.panel_win = nil
  hl.clear_all()
end

-- Exposed for the dispatcher in init.lua.
function M.show_panel() M.open_panel() end

return M
