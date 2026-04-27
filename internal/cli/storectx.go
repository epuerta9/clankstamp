package cli

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/epuerta9/clankstamp/internal/git"
	"github.com/epuerta9/clankstamp/internal/replay"
)

// openProjectStore returns the .clankstamp store for the current working
// directory.
//
// We walk up from cwd looking for an existing .clankstamp/ directory FIRST,
// and only fall back to the git toplevel if nothing was found on the way up.
// This order matters: an example directory checked into a parent git repo
// (e.g. examples/tenant-api-keys/ inside this repo) needs to resolve to its
// own nested .clankstamp/, not the parent's nonexistent one. Walk-up gives us
// the closest store, which is what users actually mean.
//
// `init` and `create` still require git (they need base_commit metadata),
// so they don't go through this helper.
func openProjectStore() (*replay.Store, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if root, ok := findStoreParent(cwd); ok {
		return replay.OpenStore(root), nil
	}
	if repo, err := git.Discover(cwd); err == nil {
		// Git toplevel as a last resort; the store may not exist yet (callers
		// like `list` handle the missing-store case explicitly).
		return replay.OpenStore(repo.Root), nil
	}
	return nil, errors.New("no .clankstamp/ store found in this directory or any parent (run `clankstamp init` inside a git repo)")
}

// findStoreParent walks up from start looking for a directory that contains
// .clankstamp/. Returns the parent path (where .clankstamp/ lives) and ok.
func findStoreParent(start string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for {
		if st, err := os.Stat(filepath.Join(dir, replay.StoreDirName)); err == nil && st.IsDir() {
			return dir, true
		} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", false
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
