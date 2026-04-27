package cli

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/epuerta9/clankstamp/internal/git"
	"github.com/epuerta9/clankstamp/internal/replay"
)

// runInit handles `clankstamp init`.
//
// It scaffolds .clankstamp/ in the current repo and (by default) appends
// `.clankstamp/` to .gitignore. Idempotent: re-running on an already-initialized
// repo prints what's there and exits 0.
func runInit(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	noGitignore := fs.Bool("no-gitignore", false, "don't append .clankstamp/ to .gitignore")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	repo, err := git.Discover(cwd)
	if err != nil {
		return fmt.Errorf("init must run inside a git repo: %w", err)
	}

	store := replay.OpenStore(repo.Root)
	already := store.Exists()
	if err := store.EnsureInit(); err != nil {
		return err
	}
	if already {
		fmt.Fprintf(stdout, "clankstamp store already initialized at %s\n", store.Root)
	} else {
		fmt.Fprintf(stdout, "initialized clankstamp store at %s\n", store.Root)
	}

	if !*noGitignore {
		added, err := replay.EnsureGitignore(repo.Root)
		if err != nil {
			return fmt.Errorf("update .gitignore: %w", err)
		}
		if added {
			fmt.Fprintln(stdout, "added .clankstamp/ to .gitignore")
		}
	}
	return nil
}
