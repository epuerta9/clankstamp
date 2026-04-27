-- Drop-in lazy.nvim / LazyVim plugin spec for clankstamp.
--
-- Install:
--   cp examples/lazy.nvim/clankstamp.lua ~/.config/nvim/lua/plugins/
--   :Lazy reload
--
-- What this gets you:
--   * :Clankstamp picker on first invocation (lazy-loaded by `cmd`)
--   * <leader>r{r,n,p,d,o,u} keymaps via <Plug> mappings (lazy-loaded by `keys`)
--   * No clobbering of any existing <leader>r... binding — our plugin checks
--     `maparg` before binding the defaults, so anything you set yourself wins.
--
-- The `dir = ...` line points lazy at the local install written by
-- `clankstamp nvim install`. That keeps everything offline and matches the
-- binary you currently have on disk. Once we cut a tagged release on GitHub,
-- you can drop the `dir` line and lazy will fetch from the repo instead.

return {
  {
    "epuerta9/clankstamp",
    dir = vim.fn.expand("~/.local/share/nvim/site/pack/clankstamp/start/clankstamp.nvim"),

    cmd = {
      "Clankstamp", "ClankstampList", "ClankstampTour",
      "ClankstampNext", "ClankstampPrev", "ClankstampDiff",
      "ClankstampOpenFile", "ClankstampMarkUnderstood",
      "ClankstampNeedsReview", "ClankstampAccept",
      "ClankstampDoctor",
      "ClankStamp", "CS", -- typo-tolerant aliases
    },

    keys = {
      { "<leader>rr", "<Plug>(ClankstampList)",       desc = "clankstamp: list stamps" },
      { "<leader>rn", "<Plug>(ClankstampNext)",       desc = "clankstamp: next tour step" },
      { "<leader>rp", "<Plug>(ClankstampPrev)",       desc = "clankstamp: previous tour step" },
      { "<leader>rd", "<Plug>(ClankstampDiff)",       desc = "clankstamp: show full diff" },
      { "<leader>ro", "<Plug>(ClankstampOpen)",       desc = "clankstamp: open step file" },
      { "<leader>ru", "<Plug>(ClankstampUnderstood)", desc = "clankstamp: mark step understood" },
    },

    -- Skip the default keymap registration in plugin/clankstamp.lua, since
    -- lazy.nvim's `keys = {}` is already wiring everything up. Without this,
    -- our VimEnter autocmd would register a second set on top.
    init = function()
      vim.g.clankstamp_no_default_maps = 1
    end,
  },
}
