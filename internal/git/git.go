// Package git wraps the small set of `git` invocations clankstamp needs.
//
// We shell out to the system git rather than depending on a Go git library:
// the surface is tiny (toplevel, HEAD sha, dirty state, working-tree diff,
// remote URL), git is universally available on developer machines, and
// shelling out keeps the binary small.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrNotARepo is returned when the working directory is not inside a git repo.
var ErrNotARepo = errors.New("not a git repository")

// Repo identifies a git repository on disk.
type Repo struct {
	Root string
}

// Discover finds the git toplevel for the given starting directory.
func Discover(startDir string) (*Repo, error) {
	out, err := run(startDir, "rev-parse", "--show-toplevel")
	if err != nil {
		if isNotARepo(err) {
			return nil, ErrNotARepo
		}
		return nil, err
	}
	root := strings.TrimSpace(out)
	if root == "" {
		return nil, ErrNotARepo
	}
	return &Repo{Root: root}, nil
}

// HeadSHA returns the full SHA of HEAD, or empty string if there are no commits yet.
func (r *Repo) HeadSHA() (string, error) {
	out, err := run(r.Root, "rev-parse", "HEAD")
	if err != nil {
		// A fresh repo with no commits has no HEAD; treat that as empty rather than error.
		if strings.Contains(err.Error(), "unknown revision") || strings.Contains(err.Error(), "ambiguous argument 'HEAD'") {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// ShortSHA returns the abbreviated SHA for the given ref. Returns empty string
// if the ref doesn't resolve.
func (r *Repo) ShortSHA(ref string) (string, error) {
	out, err := run(r.Root, "rev-parse", "--short", ref)
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(out), nil
}

// IsDirty reports whether the working tree has uncommitted changes (staged or
// unstaged, including untracked files).
func (r *Repo) IsDirty() (bool, error) {
	out, err := run(r.Root, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

// RemoteURL returns the URL of the `origin` remote if one is configured, else "".
func (r *Repo) RemoteURL() (string, error) {
	out, err := run(r.Root, "config", "--get", "remote.origin.url")
	if err != nil {
		// Exit code 1 with no output means the key isn't set; that's fine.
		return "", nil
	}
	return strings.TrimSpace(out), nil
}

// WorktreeDiff returns the unified diff of the working tree against HEAD,
// including untracked files (via --no-renames for stable hunk identity).
//
// If there are no commits yet, this returns the diff against the empty tree.
func (r *Repo) WorktreeDiff() (string, error) {
	head, err := r.HeadSHA()
	if err != nil {
		return "", err
	}
	args := []string{"diff", "--no-color", "--no-ext-diff", "--no-renames"}
	if head == "" {
		// No commits yet — diff against the empty tree object (well-known SHA).
		args = append(args, "4b825dc642cb6eb9a060e54bf8d69288fbee4904")
	} else {
		args = append(args, "HEAD")
	}
	return run(r.Root, args...)
}

// run executes `git <args...>` in dir and returns combined-stdout output.
// Stderr is folded into the error message on failure for diagnostics.
func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func isNotARepo(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "not a git repository") || strings.Contains(msg, "fatal: not a git repository")
}
