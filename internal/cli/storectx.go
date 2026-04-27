package cli

import (
	"fmt"
	"os"

	"github.com/epuerta9/clankstamp/internal/git"
	"github.com/epuerta9/clankstamp/internal/replay"
)

// openProjectStore finds the git toplevel for the current working directory
// and returns its .clankstamp store. It does not create the store directory —
// list/show/validate all want to fail gracefully if the user hasn't run init.
func openProjectStore() (*replay.Store, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	repo, err := git.Discover(cwd)
	if err != nil {
		return nil, fmt.Errorf("must run inside a git repo: %w", err)
	}
	return replay.OpenStore(repo.Root), nil
}
