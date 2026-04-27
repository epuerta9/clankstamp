package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/epuerta9/clankstamp/internal/replay"
)

// TestCreateFromWorktree spins up a real temp git repo, makes an uncommitted
// edit, runs `create --from-worktree`, and asserts the on-disk artifact shape.
func TestCreateFromWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	repoDir := t.TempDir()
	gitInit(t, repoDir)
	writeFile(t, repoDir, "hello.go", "package main\n\nfunc main() {}\n")
	gitCmd(t, repoDir, "add", ".")
	gitCmd(t, repoDir, "commit", "-m", "initial")

	// Modify the file so the worktree diff is non-empty.
	writeFile(t, repoDir, "hello.go", "package main\n\nfunc main() { println(\"hi\") }\n")

	// Switch into the temp repo for the duration of this test.
	t.Chdir(repoDir)

	var stdout, stderr bytes.Buffer
	if err := runCreate([]string{"--from-worktree", "--title", "say hello"}, &stdout, &stderr); err != nil {
		t.Fatalf("runCreate: %v\nstderr: %s", err, stderr.String())
	}

	out := stdout.String()
	if !strings.Contains(out, "created stamp run_") {
		t.Errorf("expected 'created stamp run_' in stdout, got: %s", out)
	}

	// Find the single run dir under .clankstamp/runs/.
	runsDir := filepath.Join(repoDir, ".clankstamp", "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		t.Fatalf("read runs dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 run, found %d", len(entries))
	}
	runDir := filepath.Join(runsDir, entries[0].Name())

	// Manifest sanity.
	mb, err := os.ReadFile(filepath.Join(runDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m replay.Manifest
	if err := json.Unmarshal(mb, &m); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if m.SchemaVersion != replay.SchemaVersion {
		t.Errorf("manifest schema_version = %q", m.SchemaVersion)
	}
	if m.Title != "say hello" {
		t.Errorf("manifest title = %q", m.Title)
	}
	if m.Status != replay.StatusReady {
		t.Errorf("manifest status = %q", m.Status)
	}
	if m.Repo.BaseCommit == "" {
		t.Error("manifest base_commit is empty")
	}
	if !m.Repo.DirtyAtStart {
		t.Error("expected dirty_at_start=true (we made an uncommitted edit)")
	}

	// Patch should contain our edit.
	patch := readFile(t, runDir, "patches/full.patch")
	if !strings.Contains(patch, "+func main() { println(\"hi\") }") {
		t.Errorf("patch missing the edit:\n%s", patch)
	}

	// At least one hunk should be in hunks.jsonl.
	hunksRaw := readFile(t, runDir, "hunks.jsonl")
	if !strings.Contains(hunksRaw, `"file":"hello.go"`) {
		t.Errorf("hunks.jsonl missing hello.go entry:\n%s", hunksRaw)
	}

	// Two events expected: run.started + run.finished.
	eventLines := strings.Split(strings.TrimSpace(readFile(t, runDir, "events.jsonl")), "\n")
	if len(eventLines) != 2 {
		t.Errorf("expected 2 events, got %d:\n%s", len(eventLines), strings.Join(eventLines, "\n"))
	}
}

func TestInit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	repoDir := t.TempDir()
	gitInit(t, repoDir)
	t.Chdir(repoDir)

	var stdout, stderr bytes.Buffer
	if err := runInit(nil, &stdout, &stderr); err != nil {
		t.Fatalf("runInit: %v\nstderr: %s", err, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".clankstamp", "runs")); err != nil {
		t.Errorf(".clankstamp/runs not created: %v", err)
	}
	gi := readFile(t, repoDir, ".gitignore")
	if !strings.Contains(gi, ".clankstamp/") {
		t.Errorf(".gitignore missing .clankstamp/: %q", gi)
	}

	// Idempotency: running again should not error or duplicate the gitignore line.
	stdout.Reset()
	stderr.Reset()
	if err := runInit(nil, &stdout, &stderr); err != nil {
		t.Fatalf("runInit (second call): %v", err)
	}
	gi2 := readFile(t, repoDir, ".gitignore")
	if strings.Count(gi2, ".clankstamp/") != 1 {
		t.Errorf(".gitignore has duplicate entries: %q", gi2)
	}
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	gitCmd(t, dir, "init", "-b", "main")
	gitCmd(t, dir, "config", "user.email", "test@example.com")
	gitCmd(t, dir, "config", "user.name", "test")
	gitCmd(t, dir, "config", "commit.gpgsign", "false")
}

func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v: %s", strings.Join(args, " "), err, stderr.String())
	}
}

func writeFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}
