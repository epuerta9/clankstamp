package replay

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// StoreDirName is the project-local directory where stamps are written.
const StoreDirName = ".clankstamp"

// Store is a handle to a project-local stamp store rooted at a repo's
// .clankstamp/ directory.
type Store struct {
	Root string // absolute path to .clankstamp/
}

// OpenStore returns a Store rooted at <repoRoot>/.clankstamp. It does NOT
// create the directory; use EnsureInit for that.
func OpenStore(repoRoot string) *Store {
	return &Store{Root: filepath.Join(repoRoot, StoreDirName)}
}

// Exists reports whether the store directory exists.
func (s *Store) Exists() bool {
	st, err := os.Stat(s.Root)
	return err == nil && st.IsDir()
}

// EnsureInit creates the store directory and the runs/ subdirectory if they
// don't exist. It is idempotent.
func (s *Store) EnsureInit() error {
	if err := os.MkdirAll(filepath.Join(s.Root, "runs"), 0o755); err != nil {
		return fmt.Errorf("create %s: %w", s.Root, err)
	}
	indexPath := filepath.Join(s.Root, "index.jsonl")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		f, err := os.Create(indexPath)
		if err != nil {
			return fmt.Errorf("create index.jsonl: %w", err)
		}
		_ = f.Close()
	}
	return nil
}

// RunDir returns the absolute path to a specific run's directory.
func (s *Store) RunDir(runID string) string {
	return filepath.Join(s.Root, "runs", runID)
}

// NewRunID returns a run id of the form `run_YYYYMMDD_HHMMSS_<suffix>`.
// Suffix may be a short task id, a short commit hash, or empty (a random
// 4-byte hex is appended in that case).
func NewRunID(now time.Time, suffix string) string {
	stamp := now.UTC().Format("20060102_150405")
	if suffix == "" {
		return fmt.Sprintf("run_%s", stamp)
	}
	return fmt.Sprintf("run_%s_%s", stamp, sanitizeSuffix(suffix))
}

func sanitizeSuffix(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_':
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return "anon"
	}
	return string(out)
}

// WriteManifest writes a manifest.json into the run's directory.
func WriteManifest(runDir string, m *Manifest) error {
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(runDir, "manifest.json")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(m)
}

// AppendJSONL appends one JSON-encoded record (followed by a newline) to the
// given file. The file is created if it doesn't exist.
func AppendJSONL(path string, rec any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	return enc.Encode(rec)
}

// EnsureGitignore appends `.clankstamp/` to the repo's .gitignore if it isn't
// already present. Per PRD §17, stamps are gitignored by default.
func EnsureGitignore(repoRoot string) (added bool, err error) {
	const line = StoreDirName + "/"
	path := filepath.Join(repoRoot, ".gitignore")
	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	for _, l := range splitLines(string(existing)) {
		if trim(l) == line || trim(l) == "/"+line {
			return false, nil
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return false, err
	}
	defer f.Close()
	prefix := ""
	if len(existing) > 0 && existing[len(existing)-1] != '\n' {
		prefix = "\n"
	}
	if _, err := f.WriteString(prefix + line + "\n"); err != nil {
		return false, err
	}
	return true, nil
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func trim(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t' || s[j-1] == '\r') {
		j--
	}
	return s[i:j]
}
