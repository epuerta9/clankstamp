package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// installPlan describes what an install operation will do, without doing it.
// It's emitted by --dry-run and consumed by the actual install path.
type installPlan struct {
	Dest    string   // absolute destination directory
	Files   []string // file paths relative to Dest
	Replace bool     // true if the destination already exists (requires --force)
}

// buildInstallPlan walks the source FS and records each file's relative path.
// It does NOT touch disk.
func buildInstallPlan(src fs.FS, dest string) (*installPlan, error) {
	abs, err := filepath.Abs(dest)
	if err != nil {
		return nil, err
	}
	p := &installPlan{Dest: abs}
	st, err := os.Stat(abs)
	switch {
	case err == nil && st.IsDir():
		// Only mark Replace if the directory is non-empty — installing into
		// an empty/freshly-created dir is fine without --force.
		entries, _ := os.ReadDir(abs)
		p.Replace = len(entries) > 0
	case err == nil && !st.IsDir():
		return nil, fmt.Errorf("%s exists and is not a directory", abs)
	case errors.Is(err, fs.ErrNotExist):
		// Will be created.
	default:
		return nil, err
	}

	err = fs.WalkDir(src, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || path == "." {
			return nil
		}
		p.Files = append(p.Files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// applyInstallPlan copies every file from src into the plan's destination,
// preserving the relative directory structure.
func applyInstallPlan(src fs.FS, plan *installPlan) error {
	if err := os.MkdirAll(plan.Dest, 0o755); err != nil {
		return err
	}
	for _, rel := range plan.Files {
		if err := copyFSFile(src, rel, filepath.Join(plan.Dest, rel)); err != nil {
			return fmt.Errorf("copy %s: %w", rel, err)
		}
	}
	return nil
}

func copyFSFile(src fs.FS, rel, destPath string) error {
	in, err := src.Open(rel)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// printPlan writes a human-readable rendering of the plan to w.
func printPlan(w io.Writer, header string, plan *installPlan) {
	fmt.Fprintf(w, "%s\n", header)
	fmt.Fprintf(w, "  dest: %s\n", plan.Dest)
	if plan.Replace {
		fmt.Fprintln(w, "  note: destination is non-empty (use --force to overwrite)")
	}
	for _, f := range plan.Files {
		fmt.Fprintf(w, "  + %s\n", f)
	}
}
