package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	clankstamp "github.com/epuerta9/clankstamp"
)

// defaultNvimInstallDir is the conventional pack-path location for the
// clankstamp.nvim plugin on macOS/Linux. Neovim auto-loads any plugin under
// `~/.local/share/nvim/site/pack/*/start/*` at startup, so this works without
// requiring a third-party plugin manager (PRD §9.2).
//
// Override with --path to install elsewhere (handy for chezmoi / dotfiles
// repos, or for testing). Windows users will need to pass --path until we
// resolve a proper default for `nvim-data` there.
func defaultNvimInstallDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "nvim", "site", "pack", "clankstamp", "start", "clankstamp.nvim"), nil
	}
	return filepath.Join(home, ".local", "share", "nvim", "site", "pack", "clankstamp", "start", "clankstamp.nvim"), nil
}

// runNvimInstall implements `clankstamp nvim install`.
func runNvimInstall(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("nvim install", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		path   = fs.String("path", "", "override destination directory (the clankstamp.nvim plugin root)")
		force  = fs.Bool("force", false, "overwrite an existing non-empty destination")
		dryRun = fs.Bool("dry-run", false, "print what would be installed without writing")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}

	src, err := clankstamp.NvimFS()
	if err != nil {
		return err
	}

	dest := *path
	if dest == "" {
		dest, err = defaultNvimInstallDir()
		if err != nil {
			return err
		}
	}

	plan, err := buildInstallPlan(src, dest)
	if err != nil {
		return err
	}
	if *dryRun {
		printPlan(stdout, "would install Neovim plugin to:", plan)
		return nil
	}
	if plan.Replace && !*force {
		return fmt.Errorf("destination %s is non-empty; pass --force to overwrite", plan.Dest)
	}
	if err := applyInstallPlan(src, plan); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "installed Neovim plugin -> %s (%d files)\n", plan.Dest, len(plan.Files))
	fmt.Fprintln(stdout, "open Neovim and run `:Clankstamp` to verify the plugin loaded")
	return nil
}
