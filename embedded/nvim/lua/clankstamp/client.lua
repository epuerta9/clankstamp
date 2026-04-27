-- clankstamp.client — calls the Go binary and decodes its JSON output.
--
-- The Go side guarantees three stable JSON contracts (see internal/cli/{list,show}.go):
--   `clankstamp list --json`         → array of {run_id,title,status,created_at,hunks,tour_steps}
--   `clankstamp show <id> --json`    → {manifest, tour, hunks, events}
--   `clankstamp validate --json`     → array of {run_id, ok, issues, errors, warnings}
--
-- All shellouts are synchronous — the calls return quickly because the binary
-- only reads from `.clankstamp/`. If/when stamps grow large enough that this
-- matters, we can switch to async via vim.system's callback form.

local M = {}

local function binary()
  return vim.env.CLANKSTAMP_BIN or "clankstamp"
end

local function run_json(args, opts)
  opts = opts or {}
  local cmd = vim.list_extend({ binary() }, args)
  local result = vim.system(cmd, { text = true, cwd = opts.cwd }):wait()
  if result.code ~= 0 then
    error(string.format(
      "clankstamp %s failed (exit %d):\n%s",
      table.concat(args, " "), result.code, result.stderr or ""))
  end
  if result.stdout == nil or result.stdout == "" then
    return nil
  end
  local ok, decoded = pcall(vim.json.decode, result.stdout)
  if not ok then
    error("clankstamp: failed to decode JSON output of `" .. table.concat(cmd, " ") .. "`:\n" .. result.stdout)
  end
  return decoded
end

function M.list()
  return run_json({ "list", "--json" }) or {}
end

function M.show(run_id)
  return run_json({ "show", "--json", run_id })
end

function M.validate(run_id)
  local args = { "validate", "--json" }
  if run_id then table.insert(args, run_id) end
  return run_json(args) or {}
end

-- check_binary returns (ok, version_or_error). Useful for surfacing setup
-- errors via :checkhealth in the future.
function M.check_binary()
  local ok, result = pcall(function()
    return vim.system({ binary(), "--version" }, { text = true }):wait()
  end)
  if not ok then return false, tostring(result) end
  if result.code ~= 0 then return false, result.stderr or "non-zero exit" end
  return true, vim.trim(result.stdout or "")
end

return M
